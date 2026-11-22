# Product specification

Status: proposed product contract. Application requirements are not implemented.
Current increment: DOC-001, documentation and stack comparison.

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
