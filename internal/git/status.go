package git

import (
	"bytes"
	"errors"
	"strconv"
	"strings"
)

type Repository struct{ Root, GitDir, CommonDir string }
type Branch struct {
	Name, OID, Upstream           string
	Unborn, Detached, CountsKnown bool
	Ahead, Behind                 int
}
type Change struct {
	Index, Worktree           byte
	Path, Original, Submodule string
	Untracked, Conflict       bool
}
type Snapshot struct {
	Repository Repository
	Branch     Branch
	Changes    []Change
}

func ParseStatus(data []byte) (Snapshot, error) {
	var snapshot Snapshot
	bad := errors.New("invalid or incomplete Git porcelain v2 status")
	if len(data) == 0 || data[len(data)-1] != 0 {
		return snapshot, bad
	}
	records := bytes.Split(data[:len(data)-1], []byte{0})
	seen := map[string]bool{}
	for i := 0; i < len(records); i++ {
		record := string(records[i])
		if strings.HasPrefix(record, "# ") {
			key, value, ok := strings.Cut(record[2:], " ")
			if !ok || value == "" {
				return Snapshot{}, bad
			}
			if seen[key] {
				return Snapshot{}, bad
			}
			seen[key] = true
			switch key {
			case "branch.oid":
				snapshot.Branch.Unborn = value == "(initial)"
				if !snapshot.Branch.Unborn {
					snapshot.Branch.OID = value
				}
			case "branch.head":
				snapshot.Branch.Detached = value == "(detached)"
				if !snapshot.Branch.Detached {
					snapshot.Branch.Name = value
				}
			case "branch.upstream":
				snapshot.Branch.Upstream = value
			case "branch.ab":
				fields := strings.Split(value, " ")
				if len(fields) != 2 || len(fields[0]) < 2 || len(fields[1]) < 2 || fields[0][0] != '+' || fields[1][0] != '-' {
					return Snapshot{}, bad
				}
				ahead, aerr := strconv.Atoi(fields[0][1:])
				behind, berr := strconv.Atoi(fields[1][1:])
				if aerr != nil || berr != nil || ahead < 0 || behind < 0 {
					return Snapshot{}, bad
				}
				snapshot.Branch.Ahead, snapshot.Branch.Behind, snapshot.Branch.CountsKnown = ahead, behind, true
			}
			continue
		}
		if len(record) < 3 || record[1] != ' ' {
			return Snapshot{}, bad
		}
		if record[0] == '?' {
			snapshot.Changes = append(snapshot.Changes, Change{Index: '?', Worktree: '?', Path: record[2:], Untracked: true})
			continue
		}
		fieldCount := 0
		switch record[0] {
		case '1':
			fieldCount = 9
		case '2':
			fieldCount = 10
		case 'u':
			fieldCount = 11
		default:
			return Snapshot{}, bad
		}
		fields := strings.SplitN(record, " ", fieldCount)
		if len(fields) != fieldCount || len(fields[1]) != 2 || fields[fieldCount-1] == "" {
			return Snapshot{}, bad
		}
		for _, code := range fields[1] {
			if !strings.ContainsRune(".MADRCUT", code) {
				return Snapshot{}, bad
			}
		}
		sub := fields[2]
		if sub != "N..." && (len(sub) != 4 || sub[0] != 'S' || !strings.ContainsRune(".C", rune(sub[1])) || !strings.ContainsRune(".M", rune(sub[2])) || !strings.ContainsRune(".U", rune(sub[3]))) {
			return Snapshot{}, bad
		}
		change := Change{Index: fields[1][0], Worktree: fields[1][1], Submodule: sub, Path: fields[fieldCount-1], Conflict: record[0] == 'u'}
		if record[0] == '2' {
			i++
			if i >= len(records) || len(records[i]) == 0 {
				return Snapshot{}, bad
			}
			change.Original = string(records[i])
		}
		snapshot.Changes = append(snapshot.Changes, change)
	}
	if !seen["branch.oid"] || !seen["branch.head"] {
		return Snapshot{}, bad
	}
	return snapshot, nil
}
