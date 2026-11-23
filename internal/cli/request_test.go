package cli

import (
	"bytes"
	"context"
	"strings"
	"testing"

	gitview "git-view-cli/internal/git"
)

func TestV1Grammar(t *testing.T) {
	for _, tc := range []struct {
		args  []string
		valid bool
	}{
		{[]string{"help"}, true},
		{[]string{"log", "--limit", "1", "--skip", "0"}, true},
		{[]string{"history", "--limit", "2", "--", "-odd name"}, true},
		{[]string{"diff", "--staged", "--", "*.txt"}, true},
		{[]string{"show", "--repo=.", "HEAD~2"}, true},
		{[]string{"watch", "--views", "files,files,commits"}, true},
		{[]string{"resolve", "--tool", "--", "file"}, true},
		{[]string{"stage", "--remove", "--yes", "--", "missing"}, true},
		{[]string{"log", "--limit", "101"}, false},
		{[]string{"log", "--skip", "-1"}, false},
		{[]string{"stage", "--yes", "--", "."}, false},
		{[]string{"resolve", "--", "file"}, false},
		{[]string{"resolve", "--editor", "--tool", "--", "file"}, false},
		{[]string{"watch", "--views", "changes,"}, false},
		{[]string{"log", "--staged"}, false},
		{[]string{"files", "--", "../outside"}, false},
		{[]string{"stage", "--", "a", "dir/.git/config"}, false},
	} {
		t.Run(strings.Join(tc.args, " "), func(t *testing.T) {
			_, err := parseRequest(tc.args)
			if (err == nil) != tc.valid {
				t.Fatalf("valid=%t err=%v", tc.valid, err)
			}
		})
	}
}

func TestNewCommandsDiscoverable(t *testing.T) {
	var out, diag bytes.Buffer
	Run(context.Background(), []string{"help"}, &out, &diag, func(context.Context, string) (gitview.Snapshot, error) { return gitview.Snapshot{}, nil })
	for _, name := range []string{"files", "edit", "log", "history", "show", "diff", "conflicts", "resolve", "stage", "unstage", "watch", "interactive"} {
		if !strings.Contains(out.String(), name) {
			t.Errorf("help missing %q", name)
		}
	}
}
