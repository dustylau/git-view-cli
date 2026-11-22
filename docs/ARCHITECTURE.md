# Architecture

Status: proposed, language-neutral architecture. Rust and Go remain under comparison; see the [ADR](adr/0001-application-stack.md).

## Shared application core

CLI, Watch, and Interactive call the same use cases. Model repository identity, branch state, change entries, commits, diffs, conflicts, and snapshots separately from terminal widgets. Keep data acquisition independent of rendering. Inject process, filesystem, clock, watcher, and tool-launch boundaries for deterministic tests.

```text
CLI / Watch / Interactive
          |
Shared use cases and session state
          |
Git adapter / Monitor / External tools
          |
Installed Git CLI and operating system
```

## Git adapter

Invoke Git directly with argument arrays and a known working directory, never shell-built repository commands. Discover worktree root, Git directory, and common directory through `git rev-parse`; account for linked worktrees and submodules. Keep Git protocol parsing and command construction in one adapter.

Use machine-readable output where available. Git documents stable porcelain status and NUL-separated records; prefer `git status --porcelain=v2 --branch -z` and parse bytes, including rename records and unmerged stages. Suppress optional locks on background reads. See [Git status documentation](https://git-scm.com/docs/git-status).

Use Git for file listings (`ls-files` with NUL termination), bounded commit/history queries (`log`), patches (`diff`/`show`), and unmerged entries (`ls-files --unmerged -z`). Exact invocations and parser fixtures belong in each feature specification. Do not parse decorated human-readable tables, assume fixed object-ID lengths, or assume line-delimited filenames. Resolve user revision arguments before commands and use `--` for path arguments; use literal pathspec handling where the command expects a literal filename.

Do not execute repository-configured external diff/textconv helpers during automatic read views. Explicit tool launches are a separate trust boundary. Do not use Git libraries to duplicate CLI state or introduce a second Git engine.

## Monitor

Watch worktree paths and relevant Git/common-directory metadata; explicitly account for external index/ref operations and newly created directories. Events are invalidation hints, not repository truth. Debounce bursts, coalesce duplicate work, and compute one shared snapshot for all selected views. Native watcher support differs by platform and filesystem, so periodic reconciliation and a polling fallback are required.

Use bounded queues and at most one active refresh per repository. If another invalidation arrives, schedule one follow-up refresh. Reject obsolete results using generation identifiers. Git can change during acquisition; retry a detected inconsistent snapshot and show a stale indicator when a consistent result cannot be obtained. Avoid triggering refresh loops through application logs or index-refresh side effects.

## Terminal and session lifecycle

Keep terminal rendering on the session's event loop and run Git/filesystem work off that loop. Handle resize, narrow terminals, Unicode widths, empty lists, cancellation, and plain output. Centralize raw-mode/alternate-screen cleanup across normal exits, errors, and panic/unwind paths. Do not assume Windows key events or process cancellation match Unix behavior.

Before launching an editor, pager, difftool, or mergetool, suspend the application renderer and release terminal ownership. On return, restore terminal mode and invalidate the snapshot. Do not render concurrently with a foreground child process. A GUI editor needs configured wait semantics.

## Diff and merge integration

Git generates diffs. Render plain patches first using existing viewport components. Evaluate [delta](https://github.com/dandavison/delta) as an optional external renderer; it is usable from either language and is not an assumed embeddable widget. Delegate explicitly requested comparison and merge sessions to [git difftool](https://git-scm.com/docs/git-difftool) and [git mergetool](https://git-scm.com/docs/git-mergetool).

After a merge tool returns, query unmerged index entries again. Tool exit success alone does not establish resolution. Surface tool side effects, including Git mergetool staging/backups, and never automatically commit or continue the Git operation.

## Process and resource boundaries

Apply timeouts to background reads, bound output and memory, cancel superseded work, drain pipes without deadlock, and reap children. Foreground editors have explicit cancellation/wait policy rather than background read timeouts. Test process-tree cleanup on each OS; killing the immediate process may leave helpers alive. Log diagnostics to an opt-in file outside the renderer and avoid logging repository content by default.

## Delivery

Pin the selected language toolchain and dependencies after the decision experiment. Use native OS test jobs and release artifacts for each supported target. Verify external Git availability at startup. Do not equate a single application executable with a guarantee that OS libraries, terminal behavior, or external Git/editor dependencies disappear.
