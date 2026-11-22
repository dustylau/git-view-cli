package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	gitview "git-view-cli/internal/git"
)

const Version = "0.1.0-dev"

type Reader func(context.Context, string) (gitview.Snapshot, error)

func Run(ctx context.Context, args []string, stdout, stderr io.Writer, read Reader) int {
	command, repo, err := parse(args)
	if err != nil {
		diagnostic(stderr, err)
		return 2
	}
	var output string
	switch command {
	case "help":
		output = "Usage: git-view [--repo <directory>] [status|changes]\n\nCommands:\n  status   Show repository, branch, and change counts (default)\n  changes  Show index/worktree codes and quoted paths\n\nOptions:\n  --repo <directory>  Select a working tree (also accepted after command)\n  --help, -h         Show help\n  --version          Show version\n"
	case "version":
		output = "git-view " + Version + "\n"
	default:
		snapshot, err := read(ctx, repo)
		if err != nil {
			diagnostic(stderr, err)
			switch {
			case errors.Is(err, context.Canceled):
				return 130
			case errors.Is(err, gitview.ErrMissingGit), errors.Is(err, gitview.ErrNoWorktree):
				return 2
			default:
				return 1
			}
		}
		if command == "changes" {
			output = renderChanges(snapshot)
		} else {
			output = renderStatus(snapshot)
		}
	}
	if _, err := io.WriteString(stdout, output); err != nil {
		diagnostic(stderr, fmt.Errorf("write output: %w", err))
		return 1
	}
	return 0
}

func diagnostic(out io.Writer, err error) {
	_, _ = fmt.Fprintln(out, "git-view:", strconv.QuoteToGraphic(err.Error()))
}

func parse(args []string) (command, repo string, err error) {
	command, repo = "status", "."
	selected := ""
	commandSeen := false
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--repo":
			i++
			if i >= len(args) || args[i] == "" || strings.HasPrefix(args[i], "-") {
				return "", "", errors.New("--repo requires a directory; use --repo= for paths beginning with a dash")
			}
			repo = args[i]
		case strings.HasPrefix(arg, "--repo="):
			repo = strings.TrimPrefix(arg, "--repo=")
			if repo == "" {
				return "", "", errors.New("--repo requires a non-empty directory")
			}
		case arg == "--help" || arg == "-h":
			selected = "help"
		case arg == "--version":
			selected = "version"
		case strings.HasPrefix(arg, "-"):
			return "", "", fmt.Errorf("unknown option %s; see --help", arg)
		case arg == "status" || arg == "changes" || arg == "help":
			if commandSeen {
				return "", "", errors.New("expected one command; see --help")
			}
			command = arg
			commandSeen = true
		default:
			return "", "", fmt.Errorf("unknown command or operand %s; see --help", arg)
		}
	}
	if selected != "" {
		command = selected
	}
	return command, repo, nil
}

func renderStatus(snapshot gitview.Snapshot) string {
	var out strings.Builder
	fmt.Fprintln(&out, "Repository:", strconv.QuoteToGraphic(snapshot.Repository.Root))
	branch := snapshot.Branch
	switch {
	case branch.Detached:
		fmt.Fprintf(&out, "Branch: detached (%s)\n", strconv.QuoteToGraphic(branch.OID))
	default:
		fmt.Fprint(&out, "Branch: ", strconv.QuoteToGraphic(branch.Name))
		if branch.Unborn {
			out.WriteString(" (unborn)")
		}
		out.WriteByte('\n')
	}
	if branch.Upstream == "" {
		out.WriteString("Upstream: none\n")
	} else {
		fmt.Fprintln(&out, "Upstream:", strconv.QuoteToGraphic(branch.Upstream))
	}
	if branch.CountsKnown {
		fmt.Fprintf(&out, "Ahead/behind: +%d -%d (local refs)\n", branch.Ahead, branch.Behind)
	} else {
		out.WriteString("Ahead/behind: unavailable\n")
	}
	state := "clean"
	if len(snapshot.Changes) > 0 {
		state = "dirty"
	}
	fmt.Fprintln(&out, "State:", state)
	var staged, unstaged, untracked, conflicts int
	for _, change := range snapshot.Changes {
		switch {
		case change.Conflict:
			conflicts++
		case change.Untracked:
			untracked++
		default:
			if change.Index != '.' {
				staged++
			}
			if change.Worktree != '.' {
				unstaged++
			}
		}
	}
	fmt.Fprintf(&out, "Changes: staged %d, unstaged %d, untracked %d, conflicts %d\n", staged, unstaged, untracked, conflicts)
	return out.String()
}

func renderChanges(snapshot gitview.Snapshot) string {
	if len(snapshot.Changes) == 0 {
		return "No changes.\n"
	}
	var out strings.Builder
	out.WriteString("Index Worktree Path\n")
	for _, change := range snapshot.Changes {
		fmt.Fprintf(&out, "%c %c ", change.Index, change.Worktree)
		if change.Original != "" {
			fmt.Fprint(&out, strconv.QuoteToGraphic(change.Original), " -> ")
		}
		out.WriteString(strconv.QuoteToGraphic(change.Path))
		if strings.HasPrefix(change.Submodule, "S") {
			// The adapter validates this four-character protocol enum.
			fmt.Fprintf(&out, " [submodule %s]", change.Submodule)
		}
		out.WriteByte('\n')
	}
	return out.String()
}
