package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	gitview "git-view-cli/internal/git"
)

func TestSessionArguments(t *testing.T) {
	args, err := splitLine(`history -- "dir/space Ω\tname"`)
	if err != nil || !slices.Equal(args, []string{"history", "--", "dir/space Ω\tname"}) {
		t.Fatalf("%q %v", args, err)
	}
	if _, err := splitLine(`edit -- "unfinished`); err == nil {
		t.Fatal("unfinished quote accepted")
	}
}

type notifyWriter struct{ ch chan string }

func (w notifyWriter) Write(data []byte) (int, error) { w.ch <- string(data); return len(data), nil }

func TestLiveSession(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	in, send := io.Pipe()
	defer in.Close()
	defer send.Close()
	out := notifyWriter{make(chan string, 64)}
	calls := 0
	base := gitview.Snapshot{Repository: gitview.Repository{Root: t.TempDir()}, Branch: gitview.Branch{Name: "first"}}
	a := app{in: in, out: out, diag: out, terminal: true, read: func(context.Context, string) (gitview.Snapshot, error) {
		calls++
		if calls == 3 {
			return gitview.Snapshot{}, fmt.Errorf("busy")
		}
		snap := base
		snap.Branch.Name = fmt.Sprintf("refresh-%d", calls)
		return snap, nil
	}}
	r, _ := parseRequest([]string{"interactive"})
	done := make(chan error, 1)
	go func() { done <- a.runSession(ctx, r, base, 20*time.Millisecond) }()
	wait := func(needle string) {
		t.Helper()
		var seen strings.Builder
		for {
			select {
			case s := <-out.ch:
				seen.WriteString(s)
				if strings.Contains(seen.String(), needle) {
					return
				}
			case <-ctx.Done():
				t.Fatalf("waiting for %q, got %q", needle, seen.String())
			case err := <-done:
				t.Fatalf("session ended: %v", err)
			}
		}
	}
	wait("Branch:")
	fmt.Fprintln(send, "changes")
	wait("No changes.")
	fmt.Fprintln(send, "watch --views status")
	wait("Branch:")
	wait("Stale")
	wait("refresh-4")
	fmt.Fprintln(send, "pause")
	wait("paused")
	fmt.Fprintln(send, "resume")
	wait("Branch:")
	fmt.Fprintln(send, "back")
	wait("No changes.")
	fmt.Fprintln(send, "quit")
	if err := <-done; err != nil {
		t.Fatal(err)
	}
}

func TestEditorHandoff(t *testing.T) {
	dir := t.TempDir()
	p := "space Ω.txt"
	if err := os.WriteFile(dir+"/"+p, []byte("content"), 0600); err != nil {
		t.Fatal(err)
	}
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	setTestEditor(t, exe)
	var out, diag bytes.Buffer
	s := gitview.Snapshot{Repository: gitview.Repository{Root: dir}}
	a := app{in: strings.NewReader(""), out: &out, diag: &diag, terminal: true, read: func(context.Context, string) (gitview.Snapshot, error) { return s, nil }}
	req, _ := parseRequest([]string{"edit", "--", p})
	if err := a.execute(context.Background(), req); err != nil {
		t.Fatalf("%v %s", err, &diag)
	}
	if !strings.Contains(out.String(), p) || !strings.Contains(out.String(), "No changes.") {
		t.Fatalf("handoff output %q", out.String())
	}
}
func setTestEditor(t *testing.T, exe string) {
	t.Helper()
	args, _ := json.Marshal([]string{exe, "-test.run=^TestForegroundHelper$", "--", "--print-path"})
	t.Setenv("GIT_VIEW_EDITOR", string(args))
}
func TestForegroundHelper(t *testing.T) {
	if i := slices.Index(os.Args, "--print-path"); i >= 0 {
		if len(os.Args) != i+2 || !filepath.IsAbs(os.Args[i+1]) {
			os.Exit(3)
		}
		fmt.Println(os.Args[i+1])
		os.Exit(0)
	}
}
