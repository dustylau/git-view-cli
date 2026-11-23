package cli

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	gitview "git-view-cli/internal/git"
)

func (a *app) index(ctx context.Context, w gitview.Workspace, s gitview.Snapshot, r request) error {
	unstage := r.name == "unstage"
	if err := w.CheckTargets(ctx, s.Branch, r.paths, unstage, r.remove); err != nil {
		return err
	}
	verb := r.name
	if r.remove {
		verb = "stage deletion"
	}
	for _, p := range r.paths {
		if _, err := fmt.Fprintln(a.out, verb, strconv.QuoteToGraphic(p)); err != nil {
			return err
		}
	}
	if !r.yes {
		if !a.terminal {
			return fmt.Errorf("%w: confirmation requires a terminal or --yes", errUsage)
		}
		if _, err := fmt.Fprint(a.out, "Proceed? [y/N] "); err != nil {
			return err
		}
		answer, err := a.line(ctx)
		if err != nil {
			return err
		}
		if !strings.EqualFold(answer, "y") && !strings.EqualFold(answer, "yes") {
			_, err := fmt.Fprintln(a.out, "Cancelled.")
			return err
		}
		current, err := a.read(ctx, w.Root)
		if err != nil {
			return err
		}
		if current.Branch.OID != s.Branch.OID || current.Branch.Name != s.Branch.Name || current.Branch.Detached != s.Branch.Detached {
			return errors.New("HEAD changed during confirmation; retry the command")
		}
	}
	if err := w.UpdateIndex(ctx, s.Branch, r.paths, unstage, r.remove); err != nil {
		return err
	}
	return a.report(ctx, w.Root)
}

func (a *app) report(ctx context.Context, root string) error {
	s, err := a.read(ctx, root)
	if err != nil {
		return fmt.Errorf("refresh repository state: %w", err)
	}
	_, err = io.WriteString(a.out, renderChanges(s))
	return err
}

func editorArgs(ctx context.Context, w gitview.Workspace) ([]string, error) {
	if value, ok := os.LookupEnv("GIT_VIEW_EDITOR"); ok {
		var args []string
		if err := json.Unmarshal([]byte(value), &args); err != nil || len(args) == 0 || args[0] == "" {
			return nil, fmt.Errorf("%w: GIT_VIEW_EDITOR requires a JSON array of executable and arguments", errUsage)
		}
		return args, nil
	}
	for _, key := range []string{"VISUAL", "EDITOR"} {
		if name := os.Getenv(key); name != "" {
			return []string{name}, nil
		}
	}
	name, err := w.GitEditor(ctx)
	if err != nil {
		return nil, fmt.Errorf("set GIT_VIEW_EDITOR to choose an editor: %w", err)
	}
	return []string{name}, nil
}

func (a *app) foreground(ctx context.Context, w gitview.Workspace, s gitview.Snapshot, r request) error {
	p := r.paths[0]
	if r.name == "resolve" {
		found := false
		for _, c := range s.Changes {
			if c.Conflict && p == c.Path {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("path is not unmerged: %q", p)
		}
	}
	st, err := w.PathInfo(p, r.tool)
	if err != nil {
		return err
	}
	var args []string
	if r.tool {
		if st != nil && !st.Mode().IsRegular() {
			return errors.New("mergetool requires a regular or deleted file")
		}
		args = []string{"git", "--literal-pathspecs", "mergetool", "--no-prompt", "--", p}
		if _, err := fmt.Fprintln(a.out, "Git mergetool may stage the result and leave a backup."); err != nil {
			return err
		}
	} else {
		if !st.Mode().IsRegular() {
			return errors.New("editor requires a regular file")
		}
		args, err = editorArgs(ctx, w)
		if err != nil {
			return err
		}
		args = append(args, filepath.Join(w.Root, filepath.FromSlash(p)))
	}
	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	cmd.Dir, cmd.Env = w.Root, gitview.LocalEnv()
	cmd.Stdin, cmd.Stdout, cmd.Stderr = a.in, a.out, a.diag
	launchErr := cmd.Run()
	if ctx.Err() != nil {
		return ctx.Err()
	}
	if launchErr != nil {
		if r.tool {
			launchErr = fmt.Errorf("git mergetool failed; check merge.tool: %w", launchErr)
		} else {
			launchErr = fmt.Errorf("editor failed; check GIT_VIEW_EDITOR: %w", launchErr)
		}
	}
	return errors.Join(launchErr, a.report(ctx, w.Root))
}

// Read only while the application owns the terminal. Never prefetch another
// command while an editor or mergetool is taking input.
type lineResult struct {
	text string
	err  error
}

func (a *app) readLine() <-chan lineResult {
	ch := make(chan lineResult, 1)
	go func() {
		var b strings.Builder
		var one [1]byte
		for b.Len() < 65536 {
			n, err := a.in.Read(one[:])
			if n > 0 {
				if one[0] == '\n' {
					break
				}
				b.WriteByte(one[0])
			}
			if err != nil {
				ch <- lineResult{b.String(), err}
				return
			}
		}
		if b.Len() == 65536 {
			ch <- lineResult{"", bufio.ErrTooLong}
			return
		}
		ch <- lineResult{strings.TrimSuffix(b.String(), "\r"), nil}
	}()
	return ch
}
func (a *app) line(ctx context.Context) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case result := <-a.readLine():
		return result.text, result.err
	}
}
