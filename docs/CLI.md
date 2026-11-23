# Command and mode contract

The 1.0.0 command contract extends the MVP with files, history, patches, conflict actions and live sessions.

## Status and changes

`git-view [--repo <directory>] [status|changes]` performs one read and exits. No command means `status`. `--repo` also works after the command; `--repo=<directory>` is accepted. Help (`--help`, `-h`, or `help`) and `--version` require neither Git nor a repository. Extra operands, repeated commands, unknown flags, and missing/empty repository values fail with exit 2.

Status prints repository root, branch (with unborn/detached distinction), upstream, local ahead/behind counts or `unavailable`, clean/dirty state, and staged/unstaged/untracked/conflict totals. A partially staged path contributes to both staged and unstaged totals; unmerged paths contribute only to conflicts. Changes prints index/worktree codes (`.` unchanged, `?` untracked, other letters from Git porcelain v2) and root-relative quoted paths; renames show `"source" -> "destination"`. Submodule records include Git's four-character state (`S<c><m><u>`). Ignored entries are omitted. A clean list prints `No changes.` and succeeds.

Repository metadata and paths use Go graphic string quoting: Unicode graphic characters remain readable, while control/format characters, invalid UTF-8 bytes, quotes, and backslashes are escaped. Output is plain on terminals and redirected streams. No pager, ANSI styling, JSON, or color option exists yet. Each background Git invocation has a 10-second deadline and a 16-MiB cap for each of stdout/stderr. Status and changes fail with exit 1 on overflow or timeout. Interrupt cancellation exits 130. Dirty/conflicted repositories still exit 0.

Discovery resolves worktree root, Git directory, and common directory through Git. Bare/non-repositories and missing Git fail with exit 2; subsequent Git failures exit 1. Reads suppress optional locks and configured fsmonitor helpers. No fetch, staging, config changes, or index refresh writes are requested.

The MVP compatibility floor is Git 2.43.0. Status asks Git for detailed local ahead/behind counts even when `status.aheadBehind` is disabled. A branch named `(detached)` is displayed as a named branch. Git transports are disabled for reads. In a partial clone, inspection that requires a missing object fails with exit 1 and leaves stdout empty.

## Common options

`--repo <path>` selects a working tree and defaults to discovery from the current directory. Help and version work outside a repository. Diagnostics go to stderr and command results go to stdout. Output uses plain text. See the [output decision](adr/0002-mvp-output-contract.md) for JSON schema versioning and path byte preservation.

## CLI mode

CLI is the default, performs one action, and exits. No arguments is equivalent to `status`.

| Command                                   | Result                                                                           |
| ----------------------------------------- | -------------------------------------------------------------------------------- |
| `git-view status`                         | Repository/branch summary and change counts.                                     |
| `git-view changes`                        | Paths with separate staged and unstaged status, including conflicts.             |
| `git-view files [-- <path>]`              | Directory listing with Git state; default is repository root.                    |
| `git-view edit -- <path>`                 | Launch configured editor for an existing worktree file.                          |
| `git-view log [--limit <n>] [--skip <n>]` | Paginated reachable commits from HEAD.                                           |
| `git-view history -- <path>`              | Paginated single-file history, following renames where Git supports it.          |
| `git-view show <commit>`                  | Commit metadata and diff.                                                        |
| `git-view diff [--staged] [-- <path>]`    | Working-tree versus index, or index versus HEAD.                                 |
| `git-view conflicts`                      | Unmerged paths and available stages.                                             |
| `git-view resolve --tool -- <path>`       | Explicitly launch configured Git mergetool.                                      |
| `git-view resolve --editor -- <path>`     | Open a conflicted worktree file without implicit staging.                        |
| `git-view stage -- <path>...`             | Preview and explicitly stage selected paths, using `--remove` for missing files. |
| `git-view unstage -- <path>...`           | Explicitly remove selected changes from the index, preserving worktree content.  |

Stage and unstage require confirmation when a terminal is available; `--yes` explicitly authorizes the displayed targets in scripted use. Non-terminal execution without `--yes` fails rather than waiting for input. Unstage must work on unborn branches without relying on an existing HEAD. Tool launches run only on explicit requests. Edit and resolve require foreground terminal ownership; a missing tool fails with actionable configuration guidance.

Diff displays Git patches with escaped control characters. Binary files use Git’s summary. Patches over 256 KiB produce a size notice. History defaults to 20 commits per page and accepts `--limit` from 1 to 100 and a nonnegative `--skip`. File listings cover one directory at a time, up to 10,000 entries. Use root-relative paths and `--` before file operands. For an untracked path the command displays its state.

## Watch mode

```sh
git-view watch --views status,changes
git-view --repo /path/to/repo watch --views status,commits,conflicts
```

`--views` accepts a comma-separated, deduplicated list from `status`, `changes`, `files`, `commits`, and `conflicts`; default is `status,changes`. Reject unknown or empty values before starting. `files` watches the root directory listing; filtered/path-specific live views can be added in later specifications.

Reconcile selected views every two seconds. Display the latest successful refresh and a stale/error indication when Git is busy or unavailable. Preserve the last valid snapshot instead of pretending stale data is current. Ctrl+C exits and restores terminal state. Watch does not accept mutation actions. Redirected output produces a clear terminal-required error in the initial release rather than emitting redraw control sequences.

## Interactive mode

```sh
git-view interactive
```

The session provides status, changes, files, commits, history/diff, and conflicts views. Type `help` to list commands. Enter a command and press Return. Use double-quoted Go strings for spaces or escaped filename bytes. Enter `back` inside Watch to restore the initiating view. Enter `pause` or `resume` inside session Watch.

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

Set `GIT_VIEW_EDITOR` to a JSON array of executable and arguments, for example `["code","--wait"]`. The selected absolute file path is appended as one argument. Alternatively set `VISUAL` or `EDITOR` to a single executable name or path. A configured Git editor executable is the final fallback. Use the JSON array for multi-argument commands.

## Exit codes

| Code | Meaning                                                                                                 |
| ---- | ------------------------------------------------------------------------------------------------------- |
| 0    | Requested action completed, including empty results and clean status.                                   |
| 1    | Git, filesystem, editor, or other runtime operation failed.                                             |
| 2    | Invalid arguments/configuration, missing prerequisite, unsupported terminal, or no usable working tree. |
| 130  | User cancellation, normalized consistently across supported platforms.                                  |

A dirty repository or unresolved conflict listing is data, not an execution failure. Do not expose Git's raw exit code as the application's public contract; include useful Git diagnostics without leaking environment secrets.
