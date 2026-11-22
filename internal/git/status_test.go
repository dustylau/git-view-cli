package git

import (
	"reflect"
	"strings"
	"testing"
)

const headers = "# branch.oid abcdef\x00# branch.head main\x00"

func TestParseStatus(t *testing.T) {
	for _, tc := range []struct {
		name, records string
		branch        Branch
		changes       []Change
	}{
		{"clean", headers, Branch{Name: "main", OID: "abcdef"}, nil},
		{"unborn", "# branch.oid (initial)\x00# branch.head main\x00", Branch{Name: "main", Unborn: true}, nil},
		{"detached", "# branch.head (detached)\x00# branch.oid deadbeef\x00", Branch{OID: "deadbeef", Detached: true}, nil},
		{"upstream", headers + "# branch.upstream origin/main\x00# branch.ab +12 -3\x00# future.header ignored\x00", Branch{Name: "main", OID: "abcdef", Upstream: "origin/main", CountsKnown: true, Ahead: 12, Behind: 3}, nil},
		{"paths", headers + "1 MM N... 100644 100644 100644 a b file Ω with\tspace\n\xff\x00? -new\x1b[31m\x00", Branch{Name: "main", OID: "abcdef"}, []Change{{Index: 'M', Worktree: 'M', Path: "file Ω with\tspace\n\xff", Submodule: "N..."}, {Index: '?', Worktree: '?', Path: "-new\x1b[31m", Untracked: true}}},
		{"rename", headers + "2 R. N... 100644 100644 100644 a b R100 new\nname\x00old\tname\x00", Branch{Name: "main", OID: "abcdef"}, []Change{{Index: 'R', Worktree: '.', Path: "new\nname", Original: "old\tname", Submodule: "N..."}}},
		{"unmerged", headers + "u DU N... 100644 100644 100644 100644 a b c deleted ours\x00", Branch{Name: "main", OID: "abcdef"}, []Change{{Index: 'D', Worktree: 'U', Path: "deleted ours", Submodule: "N...", Conflict: true}}},
		{"submodule", headers + "1 .M S.MU 160000 160000 160000 a b child\x00", Branch{Name: "main", OID: "abcdef"}, []Change{{Index: '.', Worktree: 'M', Path: "child", Submodule: "S.MU"}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ParseStatus([]byte(tc.records))
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got.Branch, tc.branch) || !reflect.DeepEqual(got.Changes, tc.changes) {
				t.Fatalf("got %#v, want branch %#v changes %#v", got, tc.branch, tc.changes)
			}
		})
	}
}

func TestParseStatusRejectsMalformed(t *testing.T) {
	for _, data := range []string{
		"", headers[:len(headers)-1], "# branch.head main\x00",
		headers + "1 M N... 1 2 3 a b path\x00",
		headers + "1 XX N... 1 2 3 a b path\x00",
		headers + "1 .M N... 1 2 3 a b \x00",
		headers + "1 .M bad 1 2 3 a b path\x00",
		headers + "2 R. N... 1 2 3 a b R100 new\x00",
		headers + "2 R. N... 1 2 3 a b R100 new\x00\x00",
		headers + "u UU N... short\x00", headers + "? \x00",
		headers + "# branch.ab +oops -1\x00", headers + "# branch.ab +1 2\x00",
		headers + "# branch.ab +1 -1 extra\x00", headers + "x unexpected\x00",
		headers + "# branch.head other\x00",
	} {
		t.Run(strings.ReplaceAll(data, "\x00", "|"), func(t *testing.T) {
			if got, err := ParseStatus([]byte(data)); err == nil {
				t.Fatalf("accepted malformed status: %#v", got)
			}
		})
	}
}
