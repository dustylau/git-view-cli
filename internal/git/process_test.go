package git

import (
	"context"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestProcessHelper(t *testing.T) {
	for i, arg := range os.Args {
		if arg != "--git-view-helper" {
			continue
		}
		switch os.Args[i+1] {
		case "args":
			fmt.Print(strings.Join(os.Args[i+2:], "|"))
		case "large":
			fmt.Print(strings.Repeat("x", 1024))
		case "stderr":
			fmt.Fprint(os.Stderr, strings.Repeat("e", 1024))
			os.Exit(4)
		case "wait":
			time.Sleep(30 * time.Second)
		case "failure":
			fmt.Fprint(os.Stderr, "useful diagnostic\x1b[31m")
			os.Exit(4)
		case "readenv":
			var values []string
			for _, key := range []string{"GIT_DIR", "GIT_WORK_TREE", "GIT_COMMON_DIR", "GIT_INDEX_FILE", "GIT_PREFIX", "GIT_NO_LAZY_FETCH", "GIT_TERMINAL_PROMPT"} {
				values = append(values, os.Getenv(key))
			}
			fmt.Print(strings.Join(values, "|"))
		}
		os.Exit(0)
	}
}

func TestProcess(t *testing.T) {
	bin, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	p := Process{Executable: bin, Limit: 128}
	args := []string{"-test.run=TestProcessHelper", "--", "--git-view-helper"}
	t.Run("literal arguments", func(t *testing.T) {
		got, err := p.Run(context.Background(), "", append(args, "args", "space Ω", ";$(not-a-shell)")...)
		if err != nil || string(got) != "space Ω|;$(not-a-shell)" {
			t.Fatalf("%q %v", got, err)
		}
	})
	t.Run("read-only environment", func(t *testing.T) {
		for _, key := range []string{"GIT_DIR", "GIT_WORK_TREE", "GIT_COMMON_DIR", "GIT_INDEX_FILE", "GIT_PREFIX"} {
			t.Setenv(key, "unrelated repository")
		}
		t.Setenv("GIT_NO_LAZY_FETCH", "0")
		t.Setenv("GIT_TERMINAL_PROMPT", "1")
		got, err := p.Run(context.Background(), "", append(args, "readenv")...)
		if err != nil || string(got) != "|||||1|0" {
			t.Fatalf("got %q %v", got, err)
		}
	})
	for _, mode := range []string{"large", "stderr"} {
		t.Run(mode, func(t *testing.T) {
			got, err := p.Run(context.Background(), "", append(args, mode)...)
			if !errors.Is(err, ErrOutputLimit) || got != nil {
				t.Fatalf("got %q, %v", got, err)
			}
		})
	}
	t.Run("cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		_, err := p.Run(ctx, "", append(args, "wait")...)
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("got %v", err)
		}
	})
	t.Run("deadline", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()
		_, err := p.Run(ctx, "", append(args, "wait")...)
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Fatalf("got %v", err)
		}
	})
	t.Run("own timeout", func(t *testing.T) {
		_, err := (Process{Executable: bin, Timeout: 50 * time.Millisecond}).Run(context.Background(), "", append(args, "wait")...)
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Fatalf("got %v", err)
		}
	})
	t.Run("diagnostic", func(t *testing.T) {
		_, err := p.Run(context.Background(), "", append(args, "failure")...)
		if err == nil || !strings.Contains(err.Error(), "useful diagnostic") {
			t.Fatalf("got %v", err)
		}
	})
	t.Run("missing executable", func(t *testing.T) {
		_, err := (Process{Executable: "git-view-no-such-executable"}).Run(context.Background(), "")
		if !errors.Is(err, ErrMissingGit) {
			t.Fatalf("got %v", err)
		}
	})
}

type fakeRunner struct {
	calls   [][]string
	dirs    []string
	results [][]byte
	err     error
}

func (f *fakeRunner) Run(_ context.Context, dir string, args ...string) ([]byte, error) {
	f.calls = append(f.calls, args)
	f.dirs = append(f.dirs, dir)
	if f.err != nil {
		return nil, f.err
	}
	if len(f.results) == 0 {
		return nil, errors.New("unexpected call")
	}
	result := f.results[0]
	f.results = f.results[1:]
	return result, nil
}

func TestRead(t *testing.T) {
	f := &fakeRunner{results: [][]byte{[]byte("true\n"), []byte("/root with\nnewline\n"), []byte("/metadata/worktrees/linked\n"), []byte("/metadata\n"), []byte(headers)}}
	got, err := Read(context.Background(), f, "/nested")
	if err != nil {
		t.Fatal(err)
	}
	if got.Repository != (Repository{Root: "/root with\nnewline", GitDir: "/metadata/worktrees/linked", CommonDir: "/metadata"}) {
		t.Fatalf("identity: %#v", got.Repository)
	}
	want := []string{"--no-optional-locks", "-c", "core.fsmonitor=false", "-c", "core.untrackedCache=false", "status", "--porcelain=v2", "--branch", "-z", "--untracked-files=all", "--renames", "--ignore-submodules=none", "--"}
	if !reflect.DeepEqual(f.calls[len(f.calls)-1], want) {
		t.Fatalf("status args %v", f.calls)
	}
	if f.dirs[len(f.dirs)-1] != got.Repository.Root {
		t.Fatalf("status dir %q", f.dirs)
	}
}

func TestReadFailures(t *testing.T) {
	for _, tc := range []struct {
		name   string
		runner *fakeRunner
		want   error
	}{
		{"missing git", &fakeRunner{err: ErrMissingGit}, ErrMissingGit},
		{"cancelled", &fakeRunner{err: context.Canceled}, context.Canceled},
		{"no repo", &fakeRunner{err: errors.New("dubious ownership")}, ErrNoWorktree},
		{"bare", &fakeRunner{results: [][]byte{[]byte("false\n")}}, ErrNoWorktree},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := Read(context.Background(), tc.runner, ".")
			if !errors.Is(err, tc.want) {
				t.Fatalf("got %v", err)
			}
			if tc.name == "no repo" && !strings.Contains(err.Error(), "dubious ownership") {
				t.Fatal("lost Git diagnostic")
			}
		})
	}
}
