package git

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

var ErrMissingGit = errors.New("git executable not found; install Git and add it to PATH")
var ErrNoWorktree = errors.New("working tree required")
var ErrOutputLimit = errors.New("git output exceeds the configured limit")

type Process struct {
	Executable string
	Timeout    time.Duration
	Limit      int
}

func (p Process) Run(ctx context.Context, dir string, args ...string) ([]byte, error) {
	if p.Executable == "" {
		p.Executable = "git"
	}
	if p.Timeout <= 0 {
		p.Timeout = 10 * time.Second
	}
	if p.Limit <= 0 {
		p.Limit = 16 * 1024 * 1024
	}
	ctx, cancel := context.WithTimeout(ctx, p.Timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, p.Executable, args...)
	if cmd.Err != nil {
		return nil, fmt.Errorf("%w: %v", ErrMissingGit, cmd.Err)
	}
	cmd.Dir = dir
	cmd.Env = LocalEnv()
	stdout := &boundedBuffer{limit: p.Limit, cancel: cancel}
	stderr := &boundedBuffer{limit: p.Limit, cancel: cancel}
	cmd.Stdout, cmd.Stderr = stdout, stderr
	cmd.WaitDelay = 250 * time.Millisecond
	err := cmd.Run()
	if stdout.exceeded || stderr.exceeded {
		return nil, ErrOutputLimit
	}
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	if err != nil {
		return nil, fmt.Errorf("git failed: %w: %s", err, stderr.buffer.String())
	}
	return stdout.buffer.Bytes(), nil
}

type boundedBuffer struct {
	buffer   bytes.Buffer
	limit    int
	exceeded bool
	cancel   context.CancelFunc
}

func (b *boundedBuffer) Write(data []byte) (int, error) {
	size := len(data)
	remaining := b.limit - b.buffer.Len()
	if size > remaining {
		data = data[:remaining]
		b.exceeded = true
		b.cancel()
	}
	_, _ = b.buffer.Write(data)
	return size, nil
}

type Runner interface {
	Run(context.Context, string, ...string) ([]byte, error)
}

func safeArgs(args ...string) []string {
	return append([]string{"--no-optional-locks", "-c", "core.fsmonitor=false", "-c", "core.untrackedCache=false"}, args...)
}

func Read(ctx context.Context, runner Runner, directory string) (Snapshot, error) {
	if directory == "" {
		directory = "."
	}
	read := func(dir string, args ...string) ([]byte, error) {
		return runner.Run(ctx, dir, safeArgs(args...)...)
	}
	inside, err := read(directory, "rev-parse", "--is-inside-work-tree")
	if err != nil {
		if errors.Is(err, ErrMissingGit) || errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) || errors.Is(err, ErrOutputLimit) {
			return Snapshot{}, err
		}
		return Snapshot{}, fmt.Errorf("%w: %v", ErrNoWorktree, err)
	}
	if string(inside) != "true\n" {
		return Snapshot{}, ErrNoWorktree
	}
	var identity Repository
	for _, query := range []struct {
		arg    string
		target *string
	}{
		{"--show-toplevel", &identity.Root}, {"--absolute-git-dir", &identity.GitDir}, {"--git-common-dir", &identity.CommonDir},
	} {
		out, err := read(directory, "rev-parse", "--path-format=absolute", query.arg)
		if err != nil {
			return Snapshot{}, err
		}
		*query.target = strings.TrimSuffix(string(out), "\n")
		if *query.target == "" {
			return Snapshot{}, errors.New("git returned an empty repository path")
		}
	}
	data, err := read(identity.Root, "status", "--porcelain=v2", "--branch", "--ahead-behind", "-z", "--untracked-files=all", "--renames", "--ignore-submodules=none", "--")
	if err != nil {
		return Snapshot{}, err
	}
	snapshot, err := ParseStatus(data)
	if err != nil {
		return Snapshot{}, err
	}
	if snapshot.Branch.Detached {
		// Porcelain uses the same marker for a detached HEAD and a branch with this name.
		ref, err := read(identity.Root, "symbolic-ref", "--quiet", "HEAD")
		if err != nil {
			var exit interface{ ExitCode() int }
			if snapshot.Branch.Unborn || !errors.As(err, &exit) || exit.ExitCode() != 1 {
				return Snapshot{}, err
			}
		} else {
			if string(ref) != "refs/heads/(detached)\n" {
				return Snapshot{}, errors.New("HEAD changed during status; retry the command")
			}
			snapshot.Branch.Name = "(detached)"
			snapshot.Branch.Detached = false
		}
	}
	snapshot.Repository = identity
	return snapshot, nil
}

// LocalEnv selects the requested worktree and disables transports.
func LocalEnv() []string {
	var env []string

	// Explicit directory selection takes precedence over inherited Git routing.
	for _, item := range os.Environ() {
		key, _, _ := strings.Cut(item, "=")
		switch strings.ToUpper(key) {
		case "GIT_DIR", "GIT_WORK_TREE", "GIT_COMMON_DIR", "GIT_INDEX_FILE", "GIT_PREFIX", "GIT_LITERAL_PATHSPECS", "GIT_GLOB_PATHSPECS", "GIT_NOGLOB_PATHSPECS", "GIT_ICASE_PATHSPECS", "GIT_NO_LAZY_FETCH", "GIT_TERMINAL_PROMPT", "GIT_ALLOW_PROTOCOL":
			continue
		}
		env = append(env, item)
	}
	// The empty transport allowlist also covers Git versions without lazy-fetch suppression.
	env = append(env, "GIT_NO_LAZY_FETCH=1", "GIT_TERMINAL_PROMPT=0", "GIT_ALLOW_PROTOCOL=")
	return env
}
