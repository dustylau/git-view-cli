package git

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
)

type Workspace struct {
	Repository
	Runner Runner
}

func (w Workspace) run(ctx context.Context, args ...string) ([]byte, error) {
	return w.Runner.Run(ctx, w.Root, append(safeArgs("--literal-pathspecs"), args...)...)
}

// PathInfo checks every existing component before passing a literal path to Git
// or a foreground tool. A missing suffix is valid for historical/deleted paths.
func (w Workspace) PathInfo(name string, missingOK bool) (fs.FileInfo, error) {
	root, err := os.OpenRoot(w.Root)
	if err != nil {
		return nil, err
	}
	defer root.Close()
	if name == "." {
		return root.Stat(".")
	}
	if !filepath.IsLocal(name) {
		return nil, errors.New("expected a root-relative path")
	}
	var info fs.FileInfo
	parts := strings.Split(filepath.ToSlash(name), "/")
	prefix := ""
	for i, component := range parts {
		if component == "" || component == "." || component == ".." || strings.EqualFold(strings.TrimRight(component, " ."), ".git") {
			return nil, errors.New("invalid worktree path")
		}
		prefix = filepath.Join(prefix, component)
		info, err = root.Lstat(prefix)
		if errors.Is(err, fs.ErrNotExist) && missingOK {
			return nil, nil
		}
		if err != nil {
			return nil, err
		}
		if info.IsDir() {
			for _, metadata := range []string{w.GitDir, w.CommonDir} {
				if metadata != "" {
					if st, err := os.Stat(metadata); err == nil && os.SameFile(info, st) {
						return nil, errors.New("cannot enter Git metadata")
					}
				}
			}
			if _, err := root.Lstat(filepath.Join(prefix, ".git")); err == nil {
				return nil, errors.New("nested repository: select its root with --repo")
			} else if !errors.Is(err, fs.ErrNotExist) {
				return nil, err
			}
		} else if i < len(parts)-1 {
			return nil, errors.New("cannot traverse a file or symbolic link")
		}
	}
	return info, nil
}

type Entry struct{ Path, Kind, State string }

func (w Workspace) Files(ctx context.Context, dir string) ([]Entry, error) {
	st, err := w.PathInfo(dir, false)
	if err != nil {
		return nil, err
	}
	if !st.IsDir() {
		return nil, errors.New("files requires a directory")
	}
	root, err := os.OpenRoot(w.Root)
	if err != nil {
		return nil, err
	}
	defer root.Close()
	folder, err := root.Open(dir)
	if err != nil {
		return nil, err
	}
	defer folder.Close()
	entries, err := folder.ReadDir(10001)
	if err != nil && err != io.EOF {
		return nil, err
	}
	if len(entries) > 10000 {
		return nil, errors.New("directory exceeds 10000 entries")
	}
	tracked, err := w.run(ctx, "ls-files", "--stage", "-z", "--", dir)
	if err != nil {
		return nil, err
	}
	ignored, err := w.run(ctx, "ls-files", "--others", "--ignored", "--exclude-standard", "--directory", "-z", "--", dir)
	if err != nil {
		return nil, err
	}
	prefix := ""
	if dir != "." {
		prefix = dir + "/"
	}
	// A tracked descendant makes its containing directory navigable as tracked.
	states := map[string]string{}
	modules := map[string]bool{}
	for rec := range strings.SplitSeq(strings.TrimSuffix(string(ignored), "\x00"), "\x00") {
		states[strings.TrimSuffix(rec, "/")] = "ignored"
		if strings.HasSuffix(rec, "/") && strings.HasPrefix(prefix, rec) {
			states[dir] = "ignored"
		}
	}
	for rec := range strings.SplitSeq(strings.TrimSuffix(string(tracked), "\x00"), "\x00") {
		if rec == "" {
			continue
		}
		header, name, ok := strings.Cut(rec, "\t")
		if !ok {
			return nil, errors.New("invalid ls-files record")
		}
		child, _, _ := strings.Cut(strings.TrimPrefix(name, prefix), "/")
		states[prefix+child] = "tracked"
		modules[name] = strings.HasPrefix(header, "160000 ")
	}
	var list []Entry
	for _, entry := range entries {
		if strings.EqualFold(entry.Name(), ".git") {
			continue
		}
		name := path.Join(dir, entry.Name())
		kind := "file"
		switch {
		case modules[name]:
			kind = "submodule"
		case entry.Type()&os.ModeSymlink != 0:
			kind = "symlink"
		case entry.IsDir():
			kind = "directory"
		}
		state := states[name]
		if state == "" {
			if states[dir] == "ignored" {
				state = "ignored"
			} else {
				state = "untracked"
			}
		}
		list = append(list, Entry{name, kind, state})
	}
	slices.SortFunc(list, func(a, b Entry) int { return strings.Compare(a.Path, b.Path) })
	return list, nil
}

type Commit struct{ OID, Time, Author, Subject string }

func (w Workspace) Log(ctx context.Context, b Branch, file string, limit, skip int) ([]Commit, error) {
	if file != "" {
		if _, err := w.PathInfo(file, true); err != nil {
			return nil, err
		}
	}
	if b.Unborn {
		return nil, nil
	}
	args := []string{"log", "--no-color", "--no-notes", "--no-show-signature", "--no-decorate", "--no-ext-diff", "--no-textconv", "--no-patch", "-z", "--format=tformat:%H%x00%aI%x00%an%x00%s", "--max-count=" + strconv.Itoa(limit), "--skip=" + strconv.Itoa(skip)}
	if file != "" {
		args = append(args, "--follow")
	}
	args = append(args, b.OID, "--")
	if file != "" {
		args = append(args, file)
	}
	data, err := w.run(ctx, args...)
	if err != nil {
		return nil, err
	}
	var page []Commit
	for len(data) > 0 {
		var fields [4]string
		for i := range fields {
			field, rest, ok := bytes.Cut(data, []byte{0})
			if !ok {
				return nil, errors.New("invalid git log record")
			}
			fields[i] = string(field)
			data = rest
		}
		page = append(page, Commit{fields[0], fields[1], fields[2], fields[3]})
	}
	return page, nil
}

const PatchLimit = 256 * 1024

func (w Workspace) Patch(ctx context.Context, rev, file string, staged bool) ([]byte, bool, error) {
	args := []string{"diff"}
	if rev != "" {
		id, err := w.run(ctx, "rev-parse", "--verify", "--end-of-options", rev+"^{commit}")
		if err != nil {
			return nil, false, err
		}
		args = []string{"show", "--format=fuller", "--no-notes", "--no-show-signature", "--no-decorate", strings.TrimSpace(string(id))}
	} else if staged {
		args = append(args, "--cached")
	}
	args = append(args, "--patch", "--no-color", "--no-ext-diff", "--no-textconv", "--submodule=short", "--")
	if file != "" {
		if _, err := w.PathInfo(file, true); err != nil {
			return nil, false, err
		}
		args = append(args, file)
	}
	data, err := w.run(ctx, args...)
	if errors.Is(err, ErrOutputLimit) || len(data) > PatchLimit {
		return nil, true, nil
	}
	return data, false, err
}

type Conflict struct {
	Path   string
	Stages [4]string
}

func (w Workspace) Conflicts(ctx context.Context) ([]Conflict, error) {
	data, err := w.run(ctx, "ls-files", "--unmerged", "-z", "--")
	if err != nil {
		return nil, err
	}
	var result []Conflict
	for rec := range strings.SplitSeq(strings.TrimSuffix(string(data), "\x00"), "\x00") {
		if rec == "" {
			continue
		}
		header, name, ok := strings.Cut(rec, "\t")
		fields := strings.Fields(header)
		if !ok || len(fields) != 3 || len(fields[2]) != 1 || fields[2][0] < '1' || fields[2][0] > '3' {
			return nil, fmt.Errorf("invalid conflict stage %q", header)
		}
		if len(result) == 0 || result[len(result)-1].Path != name {
			result = append(result, Conflict{Path: name})
		}
		result[len(result)-1].Stages[fields[2][0]-'0'] = fields[1]
	}
	return result, nil
}
