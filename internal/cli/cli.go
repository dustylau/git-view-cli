package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	gitview "git-view-cli/internal/git"
	"golang.org/x/term"
)

// Version is set by the linker for packaged candidates.
var Version = "1.0.0"

type Reader func(context.Context, string) (gitview.Snapshot, error)

var errUsage = errors.New("invalid request")

type app struct {
	in        io.Reader
	out, diag io.Writer
	terminal  bool
	read      Reader
}

func Run(ctx context.Context, args []string, stdout, stderr io.Writer, read Reader) int {
	outFile, ok := stdout.(*os.File)
	terminal := ok && term.IsTerminal(int(outFile.Fd())) && term.IsTerminal(int(os.Stdin.Fd()))
	a := app{os.Stdin, stdout, stderr, terminal, read}
	r, err := parseRequest(args)
	if err != nil {
		diagnostic(stderr, err)
		return 2
	}
	err = a.execute(ctx, r)
	if err == nil {
		return 0
	}
	diagnostic(stderr, err)
	switch {
	case errors.Is(err, context.Canceled):
		return 130
	case errors.Is(err, errUsage), errors.Is(err, gitview.ErrMissingGit), errors.Is(err, gitview.ErrNoWorktree):
		return 2
	default:
		return 1
	}
}

func (a *app) execute(ctx context.Context, r request) error {
	switch r.name {
	case "help":
		_, err := io.WriteString(a.out, helpText)
		return err
	case "version":
		_, err := fmt.Fprintln(a.out, "git-view", Version)
		return err
	case "watch", "interactive", "edit", "resolve":
		if !a.terminal {
			return fmt.Errorf("%w: %s requires a terminal", errUsage, r.name)
		}
	}
	s, err := a.read(ctx, r.repo)
	if err != nil {
		return err
	}
	w := gitview.Workspace{Repository: s.Repository, Runner: gitview.Process{}}
	switch r.name {
	case "stage", "unstage":
		return a.index(ctx, w, s, r)
	case "edit", "resolve":
		return a.foreground(ctx, w, s, r)
	case "watch", "interactive":
		return a.session(ctx, r, s)
	}
	output, err := renderView(ctx, w, s, r)
	if err != nil {
		return err
	}
	_, err = io.WriteString(a.out, output)
	return err
}

func diagnostic(out io.Writer, err error) {
	_, _ = fmt.Fprintln(out, "git-view:", strconv.QuoteToGraphic(err.Error()))
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
