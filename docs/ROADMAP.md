# Delivery roadmap

Every milestone follows the [mandatory workflow](../CONTRIBUTING.md). Requirements are defined in the [specification](SPECIFICATION.md). MVP 0.1.0 targets a read-only status and changes CLI for Windows and Linux on amd64 and arm64. Broader terminal modes follow in later increments. macOS support follows after 0.1.0.

| Milestone                   | Scope                                                                                       | Acceptance gate                                                                                                           |
| --------------------------- | ------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------- |
| M0: Foundation              | DOC-001: docs, directory structure, language comparison, reproducible documentation checks. | Documentation BDD first red then green; docs lint/format/link checks pass; committed clean branch.                        |
| M1: Stack decision          | Go selected by maintainer; ADR accepted, comparative experiment superseded.                 | Pin toolchain/dependencies in CLI-001; retain native/platform and terminal gates for relevant increments.                 |
| M2: Read-only CLI           | FR-002, FR-005, FR-009: repository discovery, status, changes, help/errors.                 | Real temporary-repo BDD covers clean/dirty, staged/unstaged, unborn/detached HEAD, worktrees, and unusual paths.          |
| M3: Files and history       | FR-003, FR-006, FR-007: files, commits, history, Git diffs.                                 | Paginated queries and binary/large-file fallbacks pass tests; no implicit repository mutations.                           |
| M4: Watch                   | FR-001, FR-010: selectable live views, metadata invalidation, reconciliation.               | External Git/editor changes refresh correctly; overflow/fallback/cancellation tests pass.                                 |
| M5: Interactive and editors | FR-004, FR-011: navigation, actions, mode transitions, editor handoff.                      | Terminal restored after child tools/errors; all existing functions accessible; state survives transitions.                |
| M6: Conflict assistance     | FR-008: conflict stages, explicit mergetool/editor launch, stage/unstage.                   | Content/add/delete conflict fixtures pass; report index state and preserve unrelated edits; no implicit continuation.     |
| M7: Release hardening       | NFR-001 through NFR-006; distribution, compatibility matrix, licensing.                     | Native test evidence, supported architecture artifacts, checksums, dependency review, release docs, and license decision. |

Ship vertical slices rather than implementing all parsing, then all UI, then all tests. Within each milestone, update the contract and write one failing executable behavior before implementing that behavior. Do not check off future requirements merely because they have documentation.

## Outstanding decisions

- Go chosen in [ADR 0001](adr/0001-application-stack.md); terminal/watch dependencies remain deferred.
- Exact supported OS/terminal/Git versions and the compiler support policy beyond pinned Go 1.27.4.
- Configuration schema, editor-command compatibility, and JSON schema/versioning.
- Poll/reconciliation defaults and measured responsiveness/resource budgets.
- License, executable naming/distribution availability, and release signing approach.

## MVP delivery slices

1. **CLI-001:** Go foundation, discovery, status, changes, safe plain output, bounded Git reads, executable BDD/TDD, and initial CI configuration. See [plan 0002](plans/0002-go-mvp-status.md).
2. **CLI-002 (current):** Git 2.43.0 compatibility floor, local read safeguards, branch identity, terminal cancellation, and performance measurements. Keep the output plain for 0.1.0. Native runner results remain a release gate. See [plan 0003](plans/0003-mvp-cli-compatibility.md).
3. **REL-001:** choose a license and final distribution identity, produce and checksum the four MVP artifacts, review dependencies, document installation and upgrades, and obtain release approval. Publish 0.1.0 after the native runtime gates pass.

M0 and the Go decision are complete. M2 tracks native validation through CLI-002. M3 through M6 follow the MVP.

DEV-001 provides a repeatable Linux container build and test workflow with seeded and host fixtures. See [plan 0004](plans/0004-container-test-workflow.md).
