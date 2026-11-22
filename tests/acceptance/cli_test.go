package acceptance

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/cucumber/godog"
)

type fixture struct {
	root, cwd, selected, binary string
	env                         []string
	out, err                    bytes.Buffer
	code                        int
	before                      map[string][32]byte
}

func cleanEnv() []string {
	var env []string
	for _, item := range os.Environ() {
		if !strings.HasPrefix(strings.ToUpper(item), "GIT_") {
			env = append(env, item)
		}
	}
	return append(env, "GIT_CONFIG_NOSYSTEM=1", "GIT_CONFIG_GLOBAL="+os.DevNull, "GIT_TERMINAL_PROMPT=0", "GIT_ATTR_NOSYSTEM=1")
}

func (f *fixture) git(dir string, args ...string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir, cmd.Env = dir, f.env
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git %v: %w: %s", args, err, out)
	}
	return nil
}

func (f *fixture) write(name, content string) error {
	return os.WriteFile(filepath.Join(f.selected, name), []byte(content), 0o600)
}

func (f *fixture) create(t *testing.T, kind string) error {
	f.root = t.TempDir()
	f.selected = filepath.Join(f.root, "repo space Ω")
	f.cwd, f.env = f.selected, cleanEnv()
	if err := os.Mkdir(f.selected, 0o700); err != nil {
		return err
	}
	if kind == "outside" || kind == "no git" {
		if kind == "no git" {
			for i, item := range f.env {
				if strings.HasPrefix(strings.ToUpper(item), "PATH=") {
					f.env[i] = "PATH=" + filepath.Join(f.root, "missing")
				}
			}
		}
		return nil
	}
	if kind == "bare" {
		if err := f.git(f.selected, "init", "--bare"); err != nil {
			return err
		}
		var err error
		f.before, err = snapshot(f.root)
		return err
	}
	if err := f.git(f.selected, "init", "--initial-branch=main"); err != nil {
		return err
	}
	for _, args := range [][]string{{"config", "user.name", "Fixture"}, {"config", "user.email", "fixture@example.invalid"}, {"config", "commit.gpgSign", "false"}, {"config", "core.autocrlf", "false"}} {
		if err := f.git(f.selected, args...); err != nil {
			return err
		}
	}
	if kind != "unborn" {
		for name, content := range map[string]string{"tracked.txt": "base\n", "deleted.txt": "keep\n", ".gitignore": "ignored.txt\n"} {
			if err := f.write(name, content); err != nil {
				return err
			}
		}
		if err := f.git(f.selected, "add", "."); err != nil {
			return err
		}
		if err := f.git(f.selected, "commit", "-m", "base"); err != nil {
			return err
		}
	}
	switch kind {
	case "unborn":
	case "nested":
		f.cwd = filepath.Join(f.selected, "nested")
		if err := os.Mkdir(f.cwd, 0o700); err != nil {
			return err
		}
	case "mixed":
		if err := f.write("tracked.txt", "staged\n"); err != nil {
			return err
		}
		if err := f.git(f.selected, "add", "tracked.txt"); err != nil {
			return err
		}
		if err := f.write("tracked.txt", "unstaged\n"); err != nil {
			return err
		}
		if err := os.Remove(filepath.Join(f.selected, "deleted.txt")); err != nil {
			return err
		}
		if err := f.write("-new Ω file.txt", "new\n"); err != nil {
			return err
		}
		if err := f.write("ignored.txt", "ignored\n"); err != nil {
			return err
		}
	case "rename":
		if err := f.git(f.selected, "-c", "status.renames=false", "mv", "tracked.txt", "renamed Ω file.txt"); err != nil {
			return err
		}
		if err := f.git(f.selected, "config", "status.renames", "false"); err != nil {
			return err
		}
	case "staged":
		if err := f.write("added.txt", "new content\n"); err != nil {
			return err
		}
		if err := f.git(f.selected, "add", "added.txt"); err != nil {
			return err
		}
		if err := f.git(f.selected, "rm", "deleted.txt"); err != nil {
			return err
		}
	case "submodule":
		if err := f.git(f.selected, "-c", "protocol.file.allow=always", "submodule", "add", "--", f.selected, "child"); err != nil {
			return err
		}
		if err := f.git(f.selected, "commit", "-am", "add submodule"); err != nil {
			return err
		}
		if err := f.write("child/tracked.txt", "modified\n"); err != nil {
			return err
		}
		if err := f.write("child/new.txt", "untracked\n"); err != nil {
			return err
		}
		if err := f.git(f.selected, "config", "submodule.child.ignore", "all"); err != nil {
			return err
		}
	case "detached":
		if err := f.git(f.selected, "checkout", "--detach"); err != nil {
			return err
		}
	case "worktree":
		f.selected = filepath.Join(f.root, "linked Ω tree")
		if err := f.git(f.cwd, "worktree", "add", "-b", "linked", f.selected); err != nil {
			return err
		}
		f.cwd = t.TempDir()
	case "upstream":
		if err := f.git(f.selected, "remote", "add", "origin", "https://example.invalid/no-network"); err != nil {
			return err
		}
		if err := f.git(f.selected, "update-ref", "refs/remotes/origin/main", "HEAD"); err != nil {
			return err
		}
		if err := f.git(f.selected, "branch", "--set-upstream-to=origin/main"); err != nil {
			return err
		}
		if err := f.write("tracked.txt", "ahead\n"); err != nil {
			return err
		}
		if err := f.git(f.selected, "commit", "-am", "ahead"); err != nil {
			return err
		}
	case "conflict":
		if err := f.git(f.selected, "checkout", "-b", "other"); err != nil {
			return err
		}
		if err := f.write("tracked.txt", "other\n"); err != nil {
			return err
		}
		if err := f.git(f.selected, "commit", "-am", "other"); err != nil {
			return err
		}
		if err := f.git(f.selected, "checkout", "main"); err != nil {
			return err
		}
		if err := f.write("tracked.txt", "main\n"); err != nil {
			return err
		}
		if err := f.git(f.selected, "commit", "-am", "main"); err != nil {
			return err
		}
		if err := f.git(f.selected, "merge", "other"); err == nil {
			return fmt.Errorf("fixture merge should conflict")
		}
	default:
		return fmt.Errorf("unknown fixture %q", kind)
	}
	var err error
	f.before, err = snapshot(f.root)
	return err
}

func snapshot(root string) (map[string][32]byte, error) {
	files := map[string][32]byte{}
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		files[path] = sha256.Sum256(data)
		return nil
	})
	return files, err
}

func (f *fixture) run(mode string) error {
	args, ok := map[string][]string{
		"default": {}, "status": {"status"}, "changes": {"changes"},
		"selected changes": {"changes", "--repo", f.selected},
		"selected status":  {"--repo=" + f.selected, "status"},
		"help":             {"--help"}, "version": {"--version"}, "invalid": {"--unknown"},
	}[mode]
	if !ok {
		return fmt.Errorf("unknown invocation %q", mode)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	f.out.Reset()
	f.err.Reset()
	f.code = 0
	cmd := exec.CommandContext(ctx, f.binary, args...)
	cmd.Dir, cmd.Env, cmd.Stdout, cmd.Stderr = f.cwd, f.env, &f.out, &f.err
	if err := cmd.Run(); err != nil {
		if exit, ok := err.(*exec.ExitError); ok {
			f.code = exit.ExitCode()
		} else {
			return err
		}
	}
	if ctx.Err() != nil {
		return ctx.Err()
	}
	return nil
}

func buildCLI(t *testing.T) string {
	t.Helper()
	binary := filepath.Join(t.TempDir(), "git-view")
	if runtime.GOOS == "windows" {
		binary += ".exe"
	}
	build := exec.Command("go", "build", "-o", binary, "../../cmd/git-view")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build harness: %v: %s", err, out)
	}
	return binary
}

func TestPOSIXPaths(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows forbids these control/non-UTF-8 filename cases")
	}
	f := &fixture{binary: buildCLI(t)}
	if err := f.create(t, "unborn"); err != nil {
		t.Fatal(err)
	}
	newRoot := filepath.Join(f.root, "repo\nwith\twhitespace")
	if err := os.Rename(f.selected, newRoot); err != nil {
		t.Fatal(err)
	}
	f.selected, f.cwd = newRoot, newRoot
	name := "-path Ω\tline\n\x1b[31m\xff"
	if err := f.write(name, "content\n"); err != nil {
		t.Fatal(err)
	}
	before, err := snapshot(f.root)
	if err != nil {
		t.Fatal(err)
	}
	for mode, want := range map[string]string{"changes": "? ? " + strconv.QuoteToGraphic(name), "status": "Repository: " + strconv.QuoteToGraphic(newRoot)} {
		if err := f.run(mode); err != nil {
			t.Fatal(err)
		}
		if f.code != 0 || f.err.Len() != 0 || !strings.Contains(f.out.String(), want) || strings.ContainsRune(f.out.String(), '\x1b') {
			t.Fatalf("%s: code %d stdout %q stderr %q", mode, f.code, f.out.String(), f.err.String())
		}
	}
	after, err := snapshot(f.root)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(before, after) {
		t.Fatal("fixture changed")
	}
}

func TestCLI(t *testing.T) {
	binary := buildCLI(t)
	var f *fixture
	suite := godog.TestSuite{
		Name: "cli", Options: &godog.Options{Format: "progress", Paths: []string{"../features/cli.feature"}, Strict: true, TestingT: t},
		ScenarioInitializer: func(sc *godog.ScenarioContext) {
			sc.Before(func(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
				f = &fixture{binary: binary}
				return ctx, nil
			})
			sc.Step(`^a Git fixture "([^"]*)"$`, func(kind string) error { return f.create(t, kind) })
			sc.Step(`^I run git-view with "([^"]*)"$`, func(mode string) error { return f.run(mode) })
			sc.Step(`^the exit code is (\d+)$`, func(want int) error {
				if f.code != want {
					return fmt.Errorf("exit %d, want %d; stderr: %s", f.code, want, f.err.String())
				}
				return nil
			})
			sc.Step(`^(stdout|stderr) contains (".*")$`, func(stream, want string) error {
				value, err := strconv.Unquote(want)
				if err != nil {
					return err
				}
				actual := f.out.String()
				if stream == "stderr" {
					actual = f.err.String()
				}
				if !strings.Contains(actual, value) {
					return fmt.Errorf("%s missing %q: %s", stream, value, actual)
				}
				return nil
			})
			sc.Step(`^stdout excludes "([^"]*)"$`, func(value string) error {
				if strings.Contains(f.out.String(), value) {
					return fmt.Errorf("unexpected %q: %s", value, f.out.String())
				}
				return nil
			})
			sc.Step(`^(stdout|stderr) is empty$`, func(stream string) error {
				value := f.out.String()
				if stream == "stderr" {
					value = f.err.String()
				}
				if value != "" {
					return fmt.Errorf("%s is not empty: %s", stream, value)
				}
				return nil
			})
			sc.Step(`^the fixture is unchanged$`, func() error {
				after, err := snapshot(f.root)
				if err != nil {
					return err
				}
				if !reflect.DeepEqual(f.before, after) {
					return fmt.Errorf("read changed fixture contents")
				}
				return nil
			})
		},
	}
	if suite.Run() != 0 {
		t.Fatal("CLI acceptance failed")
	}
}
