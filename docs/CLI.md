# Command and mode contract

Status: CLI-001 implements plain `status`, `changes`, help, and version. Other commands and modes below remain planned. The executable name `git-view` remains provisional.

## Current slice

`git-view [--repo <directory>] [status|changes]` performs one read and exits. No command means `status`. `--repo` also works after the command; `--repo=<directory>` is accepted. Help (`--help`, `-h`, or `help`) and `--version` require neither Git nor a repository. Extra operands, repeated commands, unknown flags, and missing/empty repository values fail with exit 2.

Status prints repository root, branch (with unborn/detached distinction), upstream, local ahead/behind counts or `unavailable`, clean/dirty state, and staged/unstaged/untracked/conflict totals. A partially staged path contributes to both staged and unstaged totals; unmerged paths contribute only to conflicts. Changes prints index/worktree codes (`.` unchanged, `?` untracked, other letters from Git porcelain v2) and root-relative quoted paths; renames show `"source" -> "destination"`. Submodule records include Git's four-character state (`S<c><m><u>`). Ignored entries are omitted. A clean list prints `No changes.` and succeeds.

Repository metadata and paths use Go graphic string quoting: Unicode graphic characters remain readable, while control/format characters, invalid UTF-8 bytes, quotes, and backslashes are escaped. Output is plain on terminals and redirected streams. No pager, ANSI styling, JSON, or color option exists yet. Each Git invocation has a 10-second deadline and a 16-MiB cap for each of stdout/stderr. Exceeding either bound fails with exit 1 and no partial result. Interrupt cancellation exits 130. Dirty/conflicted repositories still exit 0.

Discovery resolves worktree root, Git directory, and common directory through Git. Bare/non-repositories and missing Git fail with exit 2; subsequent Git failures exit 1. Reads suppress optional locks and configured fsmonitor helpers. No fetch, staging, config changes, or index refresh writes are requested.

## Common options

`--repo <path>` selects a working tree; default is repository discovery from the current directory. `--help` and `--version` do not require a repository. Diagnostics go to stderr and command results go to stdout. Planned `--color auto|always|never` will control presentation; auto will disable color when redirected and respect `NO_COLOR`. A later increment must version a JSON schema before offering `--json` on data commands.

## CLI mode

CLI is the default, performs one action, and exits. No arguments is equivalent to `status`.

| Command                                   | Result                                                                              |
| ----------------------------------------- | ----------------------------------------------------------------------------------- |
| `git-view status`                         | Repository/branch summary and change counts.                                        |
| `git-view changes`                        | Paths with separate staged and unstaged status, including conflicts.                |
| `git-view files [-- <path>]`              | Directory listing with Git state; default is repository root.                       |
| `git-view edit -- <path>`                 | Launch configured editor for an existing worktree file.                             |
| `git-view log [--limit <n>] [--skip <n>]` | Paginated reachable commits from HEAD.                                              |
| `git-view history -- <path>`              | Paginated single-file history, following renames where Git supports it.             |
| `git-view show <commit>`                  | Commit metadata and diff.                                                           |
| `git-view diff [--staged] [-- <path>]`    | Working-tree versus index, or index versus HEAD.                                    |
| `git-view conflicts`                      | Unmerged paths and available stages.                                                |
| `git-view resolve --tool -- <path>`       | Explicitly launch configured Git mergetool.                                         |
| `git-view resolve --editor -- <path>`     | Open a conflicted worktree file without implicit staging.                           |
| `git-view stage -- <path>...`             | Preview and explicitly stage selected paths, including chosen deletion resolutions. |
| `git-view unstage -- <path>...`           | Explicitly remove selected changes from the index, preserving worktree content.     |

Stage and unstage require confirmation when a terminal is available; `--yes` explicitly authorizes the displayed targets in scripted use. Non-terminal execution without `--yes` fails rather than waiting for input. Unstage must work on unborn branches without relying on an existing HEAD. Tool launches run only on explicit requests. Edit and resolve require foreground terminal ownership; a missing tool fails with actionable configuration guidance.

Diff defaults to Git's patch output in a bounded viewport or plain stream. A user may explicitly select an installed enhanced pager or configured difftool. There is no new diff algorithm. Untracked file content is not silently compared as if tracked: show its state and offer explicit preview instead.

## Watch mode

```sh
git-view watch --views status,changes
git-view --repo /path/to/repo watch --views status,commits,conflicts
```

`--views` accepts a comma-separated, deduplicated list from `status`, `changes`, `files`, `commits`, and `conflicts`; default is `status,changes`. Reject unknown or empty values before starting. `files` watches the root directory listing; filtered/path-specific live views can be added in later specifications.

Refresh on debounced filesystem/Git metadata events and periodically reconcile. Display the latest successful refresh and a stale/error indication when Git is busy or unavailable. Preserve the last valid snapshot instead of pretending stale data is current. Ctrl+C exits and restores terminal state. Watch does not accept mutation actions. Redirected output produces a clear terminal-required error in the initial release rather than emitting redraw control sequences.

## Interactive mode

```sh
git-view interactive
```

Provide navigable status, changes, files, commits, history/diff, and conflicts views. A discoverable action menu exposes every CLI function using shared services. Show help in the session; exact keys will be specified with the interactive milestone.

| Transition               | Behavior                                                                        |
| ------------------------ | ------------------------------------------------------------------------------- |
| Change view              | Preserve repository, filters, and meaningful selection.                         |
| Run a one-shot action    | Show its result or launch the tool; return to the initiating view and refresh.  |
| Enter Watch presentation | Choose supported live views within the same session.                            |
| Leave Watch presentation | Return to interactive navigation with session context intact.                   |
| Pause/resume updates     | Hold the displayed snapshot with a visible paused indicator; refresh on resume. |
| Quit or cancel           | Restore terminal ownership and end/reap application-owned subprocesses.         |

Standalone CLI and Watch invocations exit when complete or cancelled. Mode switching inside Interactive uses these same behaviors without recursively starting new copies of the application.

## Editor configuration

Prefer explicit user-level application configuration with separate executable and argument-array fields. Support `VISUAL`, then `EDITOR`, then Git's editor selection only through a documented compatibility adapter; do not split command strings on spaces or pass repository-derived strings to a shell. Exact configuration schema and command-string compatibility are deferred to the tool-launch specification. Support wait arguments for GUI editors; do not assume a launched process means edits are complete.

## Exit codes

| Code | Meaning                                                                                                 |
| ---- | ------------------------------------------------------------------------------------------------------- |
| 0    | Requested action completed, including empty results and clean status.                                   |
| 1    | Git, filesystem, editor, or other runtime operation failed.                                             |
| 2    | Invalid arguments/configuration, missing prerequisite, unsupported terminal, or no usable working tree. |
| 130  | User cancellation, normalized consistently across supported platforms.                                  |

A dirty repository or unresolved conflict listing is data, not an execution failure. Do not expose Git's raw exit code as the application's public contract; include useful Git diagnostics without leaking environment secrets.
