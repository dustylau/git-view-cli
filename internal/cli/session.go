package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"slices"
	"strconv"
	"strings"
	"time"

	gitview "git-view-cli/internal/git"
)

func splitLine(line string) ([]string, error) {
	var args []string
	for line = strings.TrimSpace(line); line != ""; line = strings.TrimLeft(line, " \t") {
		end := strings.IndexAny(line, " \t")
		if line[0] != '"' {
			if end < 0 {
				end = len(line)
			}
			args = append(args, line[:end])
			line = line[end:]
			continue
		}
		quoted, err := strconv.QuotedPrefix(line)
		if err != nil {
			return nil, err
		}
		value, err := strconv.Unquote(quoted)
		if err != nil {
			return nil, err
		}
		args = append(args, value)
		line = line[len(quoted):]
		if line != "" && line[0] != ' ' && line[0] != '\t' {
			return nil, errors.New("expected a space after quoted argument")
		}
	}
	return args, nil
}

type refreshResult struct {
	text string
	err  error
}
type session struct {
	app                                  *app
	root                                 string
	view                                 request
	interactive, watching, paused, stale bool
	views                                []string
	last                                 string
	pending                              <-chan refreshResult
	cancel                               context.CancelFunc
}

func (s *session) start(ctx context.Context, initial *gitview.Snapshot) {
	if s.pending != nil || s.paused {
		return
	}
	ctx, s.cancel = context.WithCancel(ctx)
	result := make(chan refreshResult, 1)
	s.pending = result
	queries := []request{s.view}
	if s.watching {
		queries = nil
		for _, view := range s.views {
			q := request{name: view, limit: 20}
			if view == "commits" {
				q.name = "log"
			}
			queries = append(queries, q)
		}
	}
	// Only the session loop writes to the terminal. The worker publishes one
	// complete result so failed refreshes cannot replace part of a live view.
	go func() {
		var snap gitview.Snapshot
		var err error
		if initial != nil {
			snap = *initial
		} else {
			snap, err = s.app.read(ctx, s.root)
		}
		var buf strings.Builder
		if err == nil {
			w := gitview.Workspace{Repository: snap.Repository, Runner: gitview.Process{}}
			for _, q := range queries {
				var text string
				text, err = renderView(ctx, w, snap, q)
				if err != nil {
					break
				}
				fmt.Fprintf(&buf, "[%s]\n%s", q.name, text)
			}
		}
		result <- refreshResult{buf.String(), err}
	}()
}

func (s *session) stop() {
	if s.pending != nil {
		s.cancel()
		<-s.pending
		s.pending = nil
	}
}

func (s *session) display(result refreshResult) error {
	if result.err != nil {
		s.stale = true
		_, err := fmt.Fprintf(s.app.diag, "Stale view: %s\n", strconv.QuoteToGraphic(result.err.Error()))
		return err
	}
	if s.watching && !s.stale && s.last == result.text {
		return nil
	}
	s.last, s.stale = result.text, false
	if s.watching {
		if _, err := fmt.Fprintln(s.app.out, "Updated", time.Now().Format("15:04:05")); err != nil {
			return err
		}
	}
	_, err := io.WriteString(s.app.out, result.text)
	return err
}

func (a *app) session(ctx context.Context, r request, initial gitview.Snapshot) error {
	return a.runSession(ctx, r, initial, 2*time.Second)
}

func (a *app) runSession(ctx context.Context, r request, initial gitview.Snapshot, interval time.Duration) error {
	s := session{app: a, root: initial.Repository.Root, view: request{name: "status", limit: 20}, interactive: r.name == "interactive", watching: r.name == "watch", views: r.views}
	var input <-chan lineResult
	if s.interactive {
		if _, err := fmt.Fprintln(a.out, "Enter a command and press Return. Type help for commands or quit to exit."); err != nil {
			return err
		}
		input = a.readLine()
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	s.start(ctx, &initial)
	defer s.stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case result := <-s.pending:
			s.cancel()
			s.pending = nil
			if err := s.display(result); err != nil {
				return err
			}
			if s.interactive && !s.watching {
				if _, err := fmt.Fprint(a.out, "git-view> "); err != nil {
					return err
				}
			}
		case <-ticker.C:
			if s.watching {
				s.start(ctx, nil)
			}
		case line := <-input:
			if errors.Is(line.err, io.EOF) {
				return nil
			}
			if line.err != nil {
				return line.err
			}
			args, err := splitLine(line.text)
			if err == nil {
				if slices.Equal(args, []string{"quit"}) || slices.Equal(args, []string{"q"}) {
					return nil
				}
				err = s.command(ctx, args)
			}
			if err != nil {
				diagnostic(a.diag, err)
			}
			input = a.readLine()
		}
	}
}

func (s *session) command(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return nil
	}
	if s.watching {
		if len(args) != 1 {
			return errors.New("watch commands: pause, resume, back, quit")
		}
		switch args[0] {
		case "back":
			s.stop()
			s.watching = false
			s.paused = false
		case "pause":
			s.stop()
			s.paused = true
			_, err := fmt.Fprintln(s.app.out, "Watch paused.")
			return err
		case "resume":
			s.paused = false
		default:
			return errors.New("watch commands: pause, resume, back, quit")
		}
	} else {
		before, _, _ := cutPaths(args)
		for _, arg := range before {
			if arg == "--repo" || strings.HasPrefix(arg, "--repo=") {
				return errors.New("this session uses one repository")
			}
		}
		r, err := parseRequest(args)
		if err != nil {
			return err
		}
		r.repo = s.root
		s.stop()
		switch r.name {
		case "interactive":
			return errors.New("already in Interactive mode")
		case "watch":
			s.watching = true
			s.views = r.views
			if _, err := fmt.Fprintln(s.app.out, "Watch commands: pause, resume, back, quit"); err != nil {
				return err
			}
		case "stage", "unstage", "edit", "resolve", "version", "help":
			if err := s.app.execute(ctx, r); err != nil {
				diagnostic(s.app.diag, err)
			}
		default:
			s.view = r
		}
	}
	s.last = ""
	s.start(ctx, nil)
	return nil
}
