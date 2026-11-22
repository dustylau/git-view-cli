package cli

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	gitview "git-view-cli/internal/git"
)

func TestArguments(t *testing.T) {
	for _, tc := range []struct {
		args          []string
		command, repo string
		code          int
	}{
		{nil, "status", ".", 0}, {[]string{"status"}, "status", ".", 0},
		{[]string{"--repo", "space Ω", "changes"}, "changes", "space Ω", 0},
		{[]string{"changes", "--repo=space Ω"}, "changes", "space Ω", 0},
		{[]string{"--help"}, "help", "", 0}, {[]string{"-h"}, "help", "", 0}, {[]string{"help"}, "help", "", 0}, {[]string{"--version"}, "version", "", 0},
		{[]string{"watch"}, "", "", 2}, {[]string{"--json"}, "", "", 2}, {[]string{"--color=always"}, "", "", 2},
		{[]string{"--repo"}, "", "", 2}, {[]string{"--repo="}, "", "", 2}, {[]string{"--repo", ""}, "", "", 2},
		{[]string{"status", "changes"}, "", "", 2}, {[]string{"status", "extra"}, "", "", 2},
		{[]string{"--repo", "--help"}, "", "", 2}, {[]string{"--help", "--unknown"}, "", "", 2},
	} {
		t.Run(strings.Join(tc.args, " "), func(t *testing.T) {
			var out, diagnostic bytes.Buffer
			called := false
			code := Run(context.Background(), tc.args, &out, &diagnostic, func(_ context.Context, repo string) (gitview.Snapshot, error) {
				called = true
				if repo != tc.repo {
					t.Errorf("repo %q want %q", repo, tc.repo)
				}
				return gitview.Snapshot{}, nil
			})
			if code != tc.code {
				t.Fatalf("code %d want %d: %s", code, tc.code, diagnostic.String())
			}
			if tc.code == 2 {
				if called || out.Len() != 0 || diagnostic.Len() == 0 {
					t.Fatalf("invalid arguments caused read/output: %v %q %q", called, out.String(), diagnostic.String())
				}
				return
			}
			if diagnostic.Len() != 0 {
				t.Fatal(diagnostic.String())
			}
			switch tc.command {
			case "help":
				if called || !strings.Contains(out.String(), "Usage: git-view") {
					t.Fatal("help contract")
				}
			case "version":
				if called || out.String() != "git-view "+Version+"\n" {
					t.Fatal("version contract")
				}
			default:
				if !called {
					t.Fatal("repository not read")
				}
			}
		})
	}
}

func TestRender(t *testing.T) {
	snapshot := gitview.Snapshot{
		Repository: gitview.Repository{Root: "/space Ω\n\x1b[31m"},
		Branch:     gitview.Branch{Name: "main", Unborn: true},
		Changes: []gitview.Change{
			{Index: 'M', Worktree: 'M', Path: "partial\t\x1b[31m\xff"},
			{Index: '?', Worktree: '?', Untracked: true, Path: "-new Ω"},
			{Index: 'U', Worktree: 'U', Conflict: true, Path: "conflict"},
			{Index: 'R', Worktree: '.', Path: "new", Original: "old\nname"},
			{Index: '.', Worktree: 'M', Path: "child", Submodule: "S.MU"},
		},
	}
	for _, tc := range []struct {
		command  string
		contains []string
	}{
		{"status", []string{`Repository: "/space Ω\n\x1b[31m"`, `Branch: "main" (unborn)`, "Upstream: none", "Ahead/behind: unavailable", "State: dirty", "Changes: staged 2, unstaged 2, untracked 1, conflicts 1"}},
		{"changes", []string{`M M "partial\t\x1b[31m\xff"`, `? ? "-new Ω"`, `U U "conflict"`, `R . "old\nname" -> "new"`, `. M "child" [submodule S.MU]`}},
	} {
		t.Run(tc.command, func(t *testing.T) {
			var out, diagnostic bytes.Buffer
			code := Run(context.Background(), []string{tc.command}, &out, &diagnostic, func(context.Context, string) (gitview.Snapshot, error) { return snapshot, nil })
			if code != 0 || diagnostic.Len() != 0 {
				t.Fatalf("%d %s", code, diagnostic.String())
			}
			for _, want := range tc.contains {
				if !strings.Contains(out.String(), want) {
					t.Errorf("missing %q in %q", want, out.String())
				}
			}
			if strings.ContainsRune(out.String(), '\x1b') {
				t.Fatal("unsafe terminal control")
			}
		})
	}
}

func TestErrors(t *testing.T) {
	for _, tc := range []struct {
		err  error
		code int
	}{
		{gitview.ErrMissingGit, 2}, {gitview.ErrNoWorktree, 2}, {context.Canceled, 130},
		{context.DeadlineExceeded, 1}, {gitview.ErrOutputLimit, 1}, {errors.New("failure\x1b[31m\n"), 1},
	} {
		t.Run(tc.err.Error(), func(t *testing.T) {
			var out, diagnostic bytes.Buffer
			code := Run(context.Background(), nil, &out, &diagnostic, func(context.Context, string) (gitview.Snapshot, error) { return gitview.Snapshot{}, tc.err })
			if code != tc.code || out.Len() != 0 || diagnostic.Len() == 0 || strings.ContainsRune(diagnostic.String(), '\x1b') {
				t.Fatalf("%d %q %q", code, out.String(), diagnostic.String())
			}
		})
	}
}

type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) { return 0, errors.New("closed output") }

func TestOutputFailure(t *testing.T) {
	if code := Run(context.Background(), []string{"--help"}, failingWriter{}, io.Discard, nil); code != 1 {
		t.Fatalf("got exit %d", code)
	}
}
