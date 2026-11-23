package git

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"strings"
)

// CheckTargets rejects directory pathspecs, including directories that exist
// only in the index or HEAD. A single Git command then updates the whole batch.
func (w Workspace) CheckTargets(ctx context.Context, b Branch, paths []string, unstage, remove bool) error {
	for _, p := range paths {
		st, err := w.PathInfo(p, true)
		if err != nil {
			return err
		}
		if st != nil && !st.Mode().IsRegular() && st.Mode()&fs.ModeSymlink == 0 {
			return fmt.Errorf("select individual files: %q", p)
		}
		if !unstage && (st == nil) != remove {
			return fmt.Errorf("use --remove only for a missing file: %q", p)
		}
		// Git add may implicitly replace an indexed parent file when the worktree
		// now contains a directory. Inspect this subtree before touching the index.
		top, _, _ := strings.Cut(p, "/")
		data, err := w.run(ctx, "ls-files", "--stage", "-z", "--", top)
		if err != nil {
			return err
		}
		indexed := false
		for r := range strings.SplitSeq(strings.TrimSuffix(string(data), "\x00"), "\x00") {
			if r == "" {
				continue
			}
			meta, name, ok := strings.Cut(r, "\t")
			if !ok {
				return errors.New("invalid ls-files record")
			}
			if strings.HasPrefix(p, name+"/") || strings.HasPrefix(name, p+"/") {
				return fmt.Errorf("path would replace other index entries: %q", p)
			}
			if name != p {
				continue
			}
			if strings.HasPrefix(meta, "160000 ") {
				return fmt.Errorf("select an individual indexed file: %q", p)
			}
			indexed = true
		}
		if remove && !indexed {
			return fmt.Errorf("missing index entry for %q", p)
		}
		if !b.Unborn {
			head, err := w.run(ctx, "ls-tree", "-z", b.OID, "--", p)
			if err != nil {
				return err
			}
			if len(head) > 0 && !strings.Contains(strings.SplitN(string(head), "\t", 2)[0], " blob ") {
				return fmt.Errorf("path is a directory or submodule in HEAD: %q", p)
			}
		}
	}
	return nil
}

func (w Workspace) UpdateIndex(ctx context.Context, b Branch, paths []string, unstage, remove bool) error {
	if err := w.CheckTargets(ctx, b, paths, unstage, remove); err != nil {
		return err
	}
	var args []string
	switch {
	case unstage && !b.Unborn:
		args = []string{"restore", "--source=" + b.OID, "--staged"}
	case remove || unstage:
		args = []string{"rm", "--cached", "--force"}
	default:
		args = []string{"add"}
	}
	_, err := w.run(ctx, append(append(args, "--"), paths...)...)
	return err
}

func (w Workspace) GitEditor(ctx context.Context) (string, error) {
	data, err := w.run(ctx, "var", "GIT_EDITOR")
	if err != nil {
		return "", err
	}
	value := strings.TrimSuffix(string(data), "\n")
	if value == "" {
		return "", errors.New("empty Git editor")
	}
	return value, nil
}
