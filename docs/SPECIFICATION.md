# Product specification

Status: product contract; implementation proceeds in the slices below.
Current increment: CLI-002, compatibility and release validation for the read-only CLI.

## Purpose

Provide one consistent terminal interface over one local Git working tree per session. Git remains responsible for repository semantics, revision traversal, index updates, diffs, and merges. No remote hosting account is required.

## Functional requirements

| ID     | Requirement                 | Acceptance criteria                                                                                                                                           |
| ------ | --------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| FR-001 | Monitor a local repository. | Worktree, index, HEAD, and ref changes refresh live views; periodic reconciliation recovers missed events.                                                    |
| FR-002 | Display current status.     | Show root, branch or detached/unborn HEAD, upstream, locally known ahead/behind counts, and clean/dirty state.                                                |
| FR-003 | Browse files and folders.   | Navigate the worktree, select files, distinguish tracked/untracked/ignored entries, and avoid traversing symlinks outside the root.                           |
| FR-004 | Launch editors.             | Open a selected existing file using the configured editor; suspend and restore terminal ownership and refresh afterward.                                      |
| FR-005 | List changes.               | Distinguish untracked, added, modified, deleted, renamed, staged, unstaged, and unmerged entries. A file can have staged and unstaged changes simultaneously. |
| FR-006 | Browse commits.             | Paginate reachable commits with object ID, subject, author, and timestamp. Unborn branches show empty history.                                                |
| FR-007 | Inspect history and diffs.  | Show file history, commit details, staged diffs, and working diffs. Bound large output and summarize binary content.                                          |
| FR-008 | Assist conflict resolution. | Show unmerged paths and available base/ours/theirs stages; launch editor or Git mergetool and re-query index state afterward.                                 |
| FR-009 | CLI mode.                   | Perform one action and exit. No arguments means status.                                                                                                       |
| FR-010 | Watch mode.                 | Arguments choose views; refresh without input and restore terminal state on cancellation.                                                                     |
| FR-011 | Interactive mode.           | Reach all functions and views, run one-shot actions, and switch between live Watch presentation and normal session navigation without losing context.         |

## Quality requirements

- NFR-001: Support Windows, macOS, and Linux. Target amd64 and arm64 binaries. Validate on actual OS runners; cross-compiling is insufficient evidence of runtime compatibility.
- NFR-002: Do not implicitly fetch, stage, commit, continue/abort merges, or modify Git configuration. Use optional-lock suppression for background reads.
- NFR-003: Preserve path identity and support spaces, Unicode, tabs, and newlines where the OS permits. Escape terminal control sequences in displayed repository content.
- NFR-004: Bound output, paginate history, debounce events, and support cancellation. Git subprocesses must not block terminal input.
- NFR-005: Support plain CLI output and redirected streams. Interactive and Watch presentations require a suitable terminal and explain unsupported environments.
- NFR-006: Follow the eleven-step [development workflow](../CONTRIBUTING.md) for every increment.

## Edge cases

Support linked worktrees by discovering the Git directory and common directory instead of assuming `.git` is a folder. Handle detached/unborn HEAD, missing upstream, rename pairs, staged deletions, partial staging, external Git operations, and add/delete conflicts. Represent submodules as entries with state; opening a submodule is an explicit repository change. Initially, bare repositories produce a clear working-tree-required error. Do not change safe-directory configuration automatically.

## Conflict safety

Marker removal and a zero tool exit code do not prove index resolution. A configured Git mergetool may stage files itself; always report the resulting state. Opening an editor does not stage. Explicit stage/unstage actions preview targets and preserve unrelated worktree edits. Deletion resolutions require an explicit removal action. Never auto-continue a merge or rebase. Explain that ours/theirs refer to operation-specific Git stages, not permanent branch identities.

## Initial non-goals

No replacement Git engine, custom diff algorithm, built-in three-way merge editor, hosting API, background daemon, multi-repository dashboard, automatic resolution, history rewriting, push/pull interface, or commit authoring.

## DOC-001 acceptance

Create README, specification, CLI contract, architecture, directory guide, Rust/Go comparison, proposed ADR, roadmap, workflow, testing guide, security guidance, and a current implementation plan. Create language-neutral source placeholders and test directories. Execute a real documentation BDD feature, first red for missing documents and then green. Add reproducible lint/format/link checks and CI. Do not select Rust or Go, add application dependencies, or claim application tests pass. Commit the complete increment to `docs/project-foundation`.

## MVP 0.1.0

Deliver a read-only CLI for FR-002, FR-005, and FR-009, with NFR-002 through NFR-005 safeguards applicable to plain output. Release requires native Windows/macOS/Linux test evidence, an amd64/arm64 build matrix, documented prerequisites, checksums, license selection, and no unresolved release-blocking defects. CLI-001 begins this MVP; it does not publish a release or claim all-platform support. Files/history, Watch, Interactive, editor/merge handoff, and index mutations remain later increments.

## CLI-001 acceptance

- Use Go, as explicitly selected by the maintainer on 2026-11-22; accept ADR 0001 with the comparative experiment superseded and untested portability recorded.
- `git-view` and `git-view status` show discovered worktree root, named/unborn/detached HEAD, upstream, locally known ahead/behind counts when available, clean/dirty state, and separate staged/unstaged/untracked/conflict counts.
- `git-view changes` shows root-relative paths with separate index/worktree codes, rename source and destination, unmerged pairs, and submodule state. A partially staged path counts in both staged and unstaged totals; conflicts count only as conflicts. Ignored files are omitted. Clean results succeed.
- Discover from the current directory or `--repo <directory>` (before or after the subcommand), including nested directories and linked worktrees. Reject bare/non-repositories clearly. Git ownership errors retain useful diagnostics and do not alter configuration.
- Use Git porcelain v2 with NUL records, explicit rename detection, all untracked paths, submodule scanning, optional-lock suppression, and direct argument arrays. Do not fetch or mutate worktree/index/config. Bound each Git stream to 16 MiB and each invocation to 10 seconds; fail explicitly rather than present truncated data.
- Preserve spaces, Unicode, leading dashes, tabs, newlines, and non-UTF-8 path bytes in parsing where the OS permits. Quote displayed paths/metadata using Go graphic string escaping so controls cannot change terminal state. Plain output contains no ANSI sequences or pager use.
- `--help`/`-h`, `help`, and `--version` work outside repositories and without Git. Reject unknown commands/options, extra operands, and missing/empty `--repo` values with exit 2. Runtime failures exit 1; no usable worktree or missing Git exits 2; cancellation exits 130. Results use stdout and diagnostics use stderr.
- Demonstrate real temporary-repository BDD red/green and focused TDD, run formatting/vet/selected Go linter/all current tests, update native CI, and commit the scoped work cleanly. JSON and color options require later contracts and are rejected in this slice.

## CLI-002 acceptance

- Preserve FR-002, FR-005, and FR-009 across Git configurations. Request detailed ahead/behind counts even if Git's status preference suppresses them. Report a branch literally named `(detached)` as a branch.
- For NFR-002, prohibit Git transports during reads, including when older Git retries a missing promisor object. If required objects are absent, fail clearly with exit 1 and empty stdout. A local partial-clone fixture must demonstrate that inspection cannot add objects or change Git metadata.
- Use Git 2.43.0 as the MVP compatibility floor. Record actual Git, OS, and architecture evidence, with native CI smoke tests for the six target pairs. Keep unverified targets visible as release gates under NFR-001.
- Verify NFR-004 with Ctrl+C during a pending Git read in a POSIX pseudo-terminal and an isolated Windows console. Expect exit 130, no status snapshot, and the immediate Git subprocess to be gone.
- Measure warm one-shot `status` and `changes` from a built binary against a Git-only baseline on small and medium temporary repositories. The p95 budgets are 400 ms and 1 s respectively. Keep the fixture and measurement command reproducible.
- Keep 0.1.0 output plain. Defer a JSON schema until it can contract path byte encoding and field evolution independently of the human-readable output.

## DEV-001 container build and test

- Build the CLI and prepare the Go test suite in a pinned Linux image. Keep repository data, local Git metadata, and build output out of the build context.
- By default, each run creates its own Docker volume. Seed an initial commit and staged, unstaged, and untracked changes. Inspect that volume read-only with the built CLI and verify the expected changes. Remove the volume after the run, including when seeding or tests fail.
- Accept an existing host checkout as the fixture through an explicit read-only bind mount. Run status and changes on that checkout and surface Git or startup failures. Apply the ownership exception for the mounted path only and keep it inside the test container.
- Both modes run the Go test suite with the race detector in a network-disabled Linux container. The different container filesystem and Git build provide Linux evidence under NFR-001 and NFR-002. Native macOS, Windows, and architecture validation follow the release gates.
