package acceptance

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBrowseEdges(t *testing.T) {
	f := fixture{binary: buildCLI(t)}
	if err := f.create(t, "mixed"); err != nil {
		t.Fatal(err)
	}
	run := func(code int, args ...string) string {
		t.Helper()
		if err := f.runArgs(args...); err != nil || f.code != code {
			t.Fatalf("%v: exit %d, err %v, %s", args, f.code, err, &f.err)
		}
		return f.out.String()
	}
	if err := f.git(f.selected, "config", "diff.external", "nonexistent-diff-helper"); err != nil {
		t.Fatal(err)
	}
	if err := f.write("tracked.txt", string([]byte{0, 1, 2})); err != nil {
		t.Fatal(err)
	}
	if got := run(0, "diff", "--", "tracked.txt"); !strings.Contains(got, "Binary files") {
		t.Fatalf("%q", got)
	}
	if err := f.write("tracked.txt", strings.Repeat("large\n", 60000)); err != nil {
		t.Fatal(err)
	}
	if got := run(0, "diff", "--", "tracked.txt"); !strings.Contains(got, "display limit") {
		t.Fatal("missing oversized patch summary")
	}
	if got := run(0, "diff", "--", "-new Ω file.txt"); !strings.Contains(got, "Untracked") {
		t.Fatal(got)
	}
	if err := f.git(f.selected, "rm", "--", "deleted.txt"); err != nil {
		t.Fatal(err)
	}
	if err := f.write("deleted.txt", "replacement"); err != nil {
		t.Fatal(err)
	}
	if got := run(0, "diff", "--staged", "--", "deleted.txt"); !strings.Contains(got, "-keep") {
		t.Fatalf("staged deletion hidden: %s", got)
	}
	run(2, "files", "--", ".git")
	if err := os.MkdirAll(filepath.Join(f.selected, "sub", "ignored-dir"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := f.write(".gitignore", "ignored.txt\nignored-dir/\n"); err != nil {
		t.Fatal(err)
	}
	if err := f.write("sub/ignored-dir/file", "ignored"); err != nil {
		t.Fatal(err)
	}
	if got := run(0, "files", "--", "sub/ignored-dir"); !strings.Contains(got, "ignored file") {
		t.Fatal(got)
	}
	if err := os.Symlink(f.root, filepath.Join(f.selected, "outside")); err == nil {
		run(1, "files", "--", "outside")
		run(1, "diff", "--", "outside/file")
	}
	if err := f.git(f.selected, "restore", "--source=HEAD", "--staged", "--worktree", "--", "tracked.txt"); err != nil {
		t.Fatal(err)
	}
	if err := f.git(f.selected, "mv", "--", "tracked.txt", "renamed.txt"); err != nil {
		t.Fatal(err)
	}
	if err := f.git(f.selected, "commit", "-m", "rename"); err != nil {
		t.Fatal(err)
	}
	if got := run(0, "history", "--", "renamed.txt"); !strings.Contains(got, "base") || !strings.Contains(got, "rename") {
		t.Fatal(got)
	}
}

func TestConflictKinds(t *testing.T) {
	bin := buildCLI(t)
	for _, kind := range []string{"add/add", "modify/delete"} {
		t.Run(kind, func(t *testing.T) {
			f := fixture{binary: bin}
			if err := f.create(t, "nested"); err != nil {
				t.Fatal(err)
			}
			f.cwd = f.selected
			name := "tracked.txt"
			if kind == "add/add" {
				name = "new.txt"
			}
			git := func(args ...string) {
				t.Helper()
				if err := f.git(f.selected, args...); err != nil {
					t.Fatal(err)
				}
			}
			change := func(content string) {
				t.Helper()
				if err := f.write(name, content); err != nil {
					t.Fatal(err)
				}
				git("add", "--", name)
				git("commit", "-m", content)
			}
			git("checkout", "-b", "other")
			change("other")
			git("checkout", "main")
			if kind == "modify/delete" {
				git("rm", "--", name)
				git("commit", "-m", "delete")
			} else {
				change("main")
			}
			if err := f.git(f.selected, "merge", "other"); err == nil {
				t.Fatal("expected conflict")
			}
			if err := f.run("conflicts"); err != nil || f.code != 0 {
				t.Fatalf("%v %s", err, &f.err)
			}
			if !bytes.Contains(f.out.Bytes(), []byte("theirs")) || bytes.Contains(f.out.Bytes(), []byte("base")) != (kind == "modify/delete") || bytes.Contains(f.out.Bytes(), []byte("ours")) != (kind == "add/add") {
				t.Fatal(f.out.String())
			}
			args := []string{"stage", "--yes"}
			if kind == "modify/delete" {
				if err := os.Remove(filepath.Join(f.selected, name)); err != nil {
					t.Fatal(err)
				}
				args = append(args, "--remove")
			}
			if err := f.runArgs(append(args, "--", name)...); err != nil || f.code != 0 {
				t.Fatalf("%v %s", err, &f.err)
			}
			if unmerged, err := f.gitOutput(f.selected, "ls-files", "-u"); err != nil || len(unmerged) != 0 {
				t.Fatalf("%s %v", unmerged, err)
			}
			git("rev-parse", "--verify", "MERGE_HEAD")
		})
	}
}
