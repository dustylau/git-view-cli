package cli

import (
	"errors"
	"fmt"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
)

const helpText = `Usage: git-view [--repo <directory>] [command]

  status                         Repository, branch and change counts (default)
  changes                        Index and worktree status by path
  files [-- <directory>]         List entries relative to the repository root
  log [--limit N] [--skip N]     Page through commits (default 20, maximum 100)
  history [page options] -- FILE Follow a file's history across renames
  show COMMIT                    Show commit details and patch
  diff [--staged] [-- FILE]      Show working or staged patch
  conflicts                      List unmerged paths and their Git stages
  edit -- FILE                   Open a file in the foreground editor
  resolve --editor|--tool -- FILE  Open an unmerged path and report its state
  stage [--remove] [--yes] -- FILE...  Confirm selected files or deletions
  unstage [--yes] -- FILE...     Unstage selected paths, keeping working files
  watch [--views LIST]           Refresh status,changes,files,commits,conflicts
  interactive                    Navigate using these commands in one terminal
  help, --help, -h               Show help
  --version                      Show version
`

type request struct {
	name, repo, revision              string
	paths, views                      []string
	limit, skip                       int
	staged, yes, remove, tool, editor bool
}

func parseRequest(args []string) (request, error) {
	r := request{name: "status", repo: ".", limit: 20, views: []string{"status", "changes"}}
	var words []string
	override := ""
	before, paths, hasPaths := cutPaths(args)
	for i := 0; i < len(before); i++ {
		arg := before[i]
		switch {
		case arg == "--repo":
			if i+1 == len(before) || before[i+1] == "" || strings.HasPrefix(before[i+1], "-") {
				return r, errors.New("--repo requires a directory")
			}
			i++
			r.repo = before[i]
		case strings.HasPrefix(arg, "--repo="):
			r.repo = strings.TrimPrefix(arg, "--repo=")
			if r.repo == "" {
				return r, errors.New("--repo requires a directory")
			}
		case arg == "--help" || arg == "-h":
			override = "help"
		case arg == "--version":
			override = "version"
		default:
			words = append(words, arg)
		}
	}
	if len(words) > 0 {
		r.name = words[0]
		words = words[1:]
	}
	if !slices.Contains(strings.Fields("status changes help files edit log history show diff conflicts resolve stage unstage watch interactive"), r.name) {
		if strings.HasPrefix(r.name, "-") {
			return r, fmt.Errorf("unknown option %s; see --help", r.name)
		}
		return r, fmt.Errorf("unknown command %s; see --help", r.name)
	}
	for len(words) > 0 {
		arg := words[0]
		words = words[1:]
		switch {
		case (arg == "--limit" || arg == "--skip") && (r.name == "log" || r.name == "history"):
			if len(words) == 0 {
				return r, fmt.Errorf("%s requires a number", arg)
			}
			n, err := strconv.Atoi(words[0])
			words = words[1:]
			if err != nil || n < 0 || (arg == "--limit" && (n < 1 || n > 100)) {
				return r, fmt.Errorf("invalid %s", arg)
			}
			if arg == "--limit" {
				r.limit = n
			} else {
				r.skip = n
			}
		case arg == "--views" && r.name == "watch":
			if len(words) == 0 {
				return r, errors.New("--views requires a list")
			}
			r.views = nil
			for _, v := range strings.Split(words[0], ",") {
				if !slices.Contains([]string{"status", "changes", "files", "commits", "conflicts"}, v) {
					return r, fmt.Errorf("unknown Watch view %q", v)
				}
				if !slices.Contains(r.views, v) {
					r.views = append(r.views, v)
				}
			}
			words = words[1:]
		case arg == "--staged" && r.name == "diff":
			r.staged = true
		case arg == "--remove" && r.name == "stage":
			r.remove = true
		case arg == "--yes" && (r.name == "stage" || r.name == "unstage"):
			r.yes = true
		case arg == "--tool" && r.name == "resolve":
			r.tool = true
		case arg == "--editor" && r.name == "resolve":
			r.editor = true
		case r.name == "show" && r.revision == "" && arg != "" && !strings.HasPrefix(arg, "-"):
			r.revision = arg
		default:
			return r, fmt.Errorf("unexpected argument %q for %s; see --help", arg, r.name)
		}
	}
	if override != "" && !hasPaths {
		r.name = override
		return r, nil
	}
	if r.name == "show" && r.revision == "" {
		return r, errors.New("show requires a commit")
	}
	if r.name == "resolve" && r.tool == r.editor {
		return r, errors.New("resolve requires exactly one of --tool or --editor")
	}
	if hasPaths && len(paths) == 0 {
		return r, errors.New("expected a path after --")
	}
	switch r.name {
	case "edit", "history", "resolve":
		if len(paths) != 1 {
			return r, errors.New("expected -- followed by one file")
		}
	case "files", "diff":
		if len(paths) > 1 {
			return r, errors.New("expected at most one path")
		}
	case "stage", "unstage":
		if len(paths) == 0 {
			return r, errors.New("expected -- followed by one or more files")
		}
	default:
		if hasPaths {
			return r, fmt.Errorf("%s takes no paths", r.name)
		}
	}
	for _, p := range paths {
		name := filepath.ToSlash(p)
		if !filepath.IsLocal(p) || strings.IndexByte(p, 0) >= 0 {
			return r, errors.New("expected a root-relative worktree path")
		}
		if name == "." && r.name == "files" {
			continue
		}
		for _, part := range strings.Split(name, "/") {
			if part == ".." || part == "." || part == "" || strings.EqualFold(strings.TrimRight(part, " ."), ".git") {
				return r, fmt.Errorf("invalid worktree path %q", p)
			}
		}
		r.paths = append(r.paths, name)
	}
	return r, nil
}

func cutPaths(args []string) (before, paths []string, found bool) {
	if i := slices.Index(args, "--"); i >= 0 {
		return args[:i], args[i+1:], true
	}
	return args, nil, false
}
