# Architecture

The [Go decision](adr/0001-application-stack.md) and [v1 terminal decision](adr/0006-v1-terminal.md) describe the implementation choices.

## Git adapter

`internal/git` discovers the worktree, Git directory and common directory through Git. The porcelain-v2 status parser preserves NUL-delimited path bytes, rename pairs and separate index/worktree states. Local branch counts use explicit ahead/behind reporting. A symbolic-ref query disambiguates a branch named `(detached)`.

The process adapter executes Git argument arrays in the selected directory with optional locks suppressed, fsmonitor disabled, repository routing variables cleared and transports disabled. A background call has a ten-second deadline, a 16-MiB limit per stream and a bounded pipe-drain interval. Bare and non-repositories produce a working-tree-required error.

File browsing joins one level of rooted filesystem entries with Git index and ignore records. Directory queries cap each level at 10,000 entries. Path checks reject metadata, directory symlinks and nested checkouts. Git log supplies bounded pages with NUL-separated metadata. Git resolves commit references before show. Plain patch views disable external diff and textconv, preserve Git's binary summaries and replace output over 256 KiB with a size notice.

## Terminal sessions

`cmd/git-view` owns interrupt cancellation. `internal/cli` shares command parsing, escaped rendering and operations among CLI, Watch and Interactive. Live modes require terminal input and output. The session loop reads complete lines, owns rendering and retains the current request including its file filter and pagination.

Watch periodically reconciles all selected views. One cancellable worker shares a status snapshot across the views and publishes their results together. Unchanged snapshots produce no extra output. A failed refresh leaves the previous snapshot visible with a stale diagnostic. The next successful refresh redraws the view.

## External tools and mutations

The application waits while an explicitly requested editor or Git mergetool owns the terminal, then reads status again. Git mergetool controls tool execution, staging and backups. The editor receives one absolute path at the end of its argument array.

Stage and unstage validate individual root-relative targets, show a preview, and confirm before issuing one Git index operation. Unstage restores from HEAD or removes an unborn index entry with `rm --cached`. Deletion staging also uses `rm --cached`. Git retains responsibility for index locking and merge state.

## Delivery

The [package tool](../tools/release/package_release.py) builds four ZIPs with a SHA-256 manifest. The [candidate workflow](../.github/workflows/release-candidate.yml) builds once and executes the same archives on each native target. See [testing](TESTING.md) for command and platform evidence.
