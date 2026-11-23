package cli

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	gitview "git-view-cli/internal/git"
)

func TestIndexSafety(t *testing.T) {
	root := t.TempDir()
	t.Setenv("GIT_CONFIG_GLOBAL", os.DevNull)
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	git := func(args ...string) string {
		t.Helper()
		c := exec.Command("git", args...)
		c.Dir = root
		out, err := c.CombinedOutput()
		if err != nil {
			t.Fatalf("%v: %s", err, out)
		}
		return string(out)
	}
	git("init", "-b", "main")
	for _, p := range []string{"[x]", "x"} {
		if err := os.WriteFile(filepath.Join(root, p), []byte(p), 0600); err != nil {
			t.Fatal(err)
		}
	}
	var out, diag bytes.Buffer
	do := func(args ...string) int {
		out.Reset()
		diag.Reset()
		return Run(context.Background(), append([]string{"--repo", root}, args...), &out, &diag, func(ctx context.Context, d string) (gitview.Snapshot, error) {
			return gitview.Read(ctx, gitview.Process{}, d)
		})
	}
	if code := do("stage", "--", "[x]"); code != 2 || git("ls-files") != "" {
		t.Fatalf("unconfirmed mutation %d %s", code, &out)
	}
	if code := do("stage", "--yes", "--", "[x]"); code != 0 || git("ls-files") != "[x]\n" {
		t.Fatalf("literal stage %d %s %s", code, &out, &diag)
	}
	if code := do("unstage", "--yes", "--", "[x]"); code != 0 || git("ls-files") != "" {
		t.Fatalf("unborn unstage %d %s", code, &diag)
	}
	if b, err := os.ReadFile(filepath.Join(root, "[x]")); err != nil || string(b) != "[x]" {
		t.Fatal("unstage changed file")
	}
	git("add", "--all")
	git("-c", "user.name=T", "-c", "user.email=t@example.invalid", "-c", "commit.gpgSign=false", "commit", "-m", "base")
	if err := os.Remove(filepath.Join(root, "x")); err != nil {
		t.Fatal(err)
	}
	if code := do("stage", "--yes", "--", "x"); code == 0 || git("diff", "--cached", "--name-only") != "" {
		t.Fatal("implicit deletion")
	}
	if code := do("stage", "--yes", "--remove", "--", "x"); code != 0 || git("diff", "--cached", "--name-only") != "x\n" {
		t.Fatalf("explicit deletion %d %s", code, &diag)
	}
}
func TestDirectoryReplacementIsScoped(t *testing.T) {
	root := t.TempDir()
	t.Setenv("GIT_CONFIG_GLOBAL", os.DevNull)
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	cmd := exec.Command("git", "init", "-b", "main", root)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("%s %v", out, err)
	}
	p := filepath.Join(root, "parent")
	if err := os.WriteFile(p, []byte("indexed"), 0600); err != nil {
		t.Fatal(err)
	}
	cmd = exec.Command("git", "-C", root, "add", "parent")
	if err := cmd.Run(); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(p); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(p, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(p, "child"), []byte("new"), 0600); err != nil {
		t.Fatal(err)
	}
	var out, diag bytes.Buffer
	code := Run(context.Background(), []string{"--repo", root, "stage", "--yes", "--", "parent/child"}, &out, &diag, func(ctx context.Context, d string) (gitview.Snapshot, error) {
		return gitview.Read(ctx, gitview.Process{}, d)
	})
	if code == 0 {
		t.Fatal("staging child silently replaced an independently indexed parent")
	}
	indexed, err := exec.Command("git", "-C", root, "ls-files").Output()
	if err != nil || string(indexed) != "parent\n" {
		t.Fatalf("%s %v", indexed, err)
	}
}

func TestEditorArgs(t *testing.T) {
	t.Setenv("GIT_VIEW_EDITOR", `["program with space","--wait"]`)
	args, err := editorArgs(context.Background(), gitview.Workspace{})
	if err != nil || fmt.Sprint(args) != "[program with space --wait]" {
		t.Fatalf("%q %v", args, err)
	}
	t.Setenv("GIT_VIEW_EDITOR", "[]")
	if _, err = editorArgs(context.Background(), gitview.Workspace{}); err == nil {
		t.Fatal("accepted empty editor")
	}
}
func TestPatchEscaping(t *testing.T) {
	got := renderPatch([]byte("one\n+\x1b[2J\n+\xff\n"))
	if strings.ContainsRune(got, 27) || !strings.Contains(got, "+\\x1b[2J") || !strings.Contains(got, "+\\xff") {
		t.Fatalf("%q", got)
	}
}
func TestConflictToolRefresh(t *testing.T) {
	root := t.TempDir()
	t.Setenv("GIT_CONFIG_GLOBAL", os.DevNull)
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	git := func(args ...string) error {
		cmd := exec.Command("git", args...)
		cmd.Dir = root
		out, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("%v: %w %s", args, err, out)
		}
		return nil
	}
	must := func(args ...string) {
		t.Helper()
		if err := git(args...); err != nil {
			t.Fatal(err)
		}
	}
	must("init", "-b", "main")
	must("config", "user.name", "Fixture")
	must("config", "user.email", "test@example.invalid")
	p := "unmerged file.txt"
	commit := func(value string) {
		if err := os.WriteFile(filepath.Join(root, p), []byte(value+"\n"), 0600); err != nil {
			t.Fatal(err)
		}
		must("add", "--", p)
		must("commit", "-m", value)
	}
	commit("base")
	must("switch", "-c", "other")
	commit("other")
	must("switch", "main")
	commit("main")
	if err := git("merge", "other"); err == nil {
		t.Fatal("expected conflict")
	}
	must("config", "merge.tool", "fixture")
	must("config", "mergetool.fixture.cmd", `cat "$REMOTE" > "$MERGED"`)
	must("config", "mergetool.fixture.trustExitCode", "true")
	var out, diag bytes.Buffer
	a := app{in: strings.NewReader(""), out: &out, diag: &diag, terminal: true, read: func(ctx context.Context, d string) (gitview.Snapshot, error) {
		return gitview.Read(ctx, gitview.Process{}, d)
	}}
	run := func(args ...string) {
		t.Helper()
		req, err := parseRequest(append([]string{"--repo", root}, args...))
		if err != nil {
			t.Fatal(err)
		}
		out.Reset()
		if err = a.execute(context.Background(), req); err != nil {
			t.Fatalf("%v %s", err, &diag)
		}
	}
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	setTestEditor(t, exe)
	run("resolve", "--editor", "--", p)
	if !strings.Contains(out.String(), `U U "unmerged file.txt"`) {
		t.Fatal(out.String())
	}
	run("resolve", "--tool", "--", p)
	if !strings.Contains(out.String(), `M . "unmerged file.txt"`) {
		t.Fatal(out.String())
	}
	must("rev-parse", "--verify", "MERGE_HEAD")
}
