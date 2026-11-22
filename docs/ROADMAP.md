# Delivery roadmap

Every milestone follows the [mandatory workflow](../CONTRIBUTING.md). Requirements are defined in the [specification](SPECIFICATION.md). Milestones describe intended scope, not completed application behavior.

| Milestone                   | Scope                                                                                       | Acceptance gate                                                                                                           |
| --------------------------- | ------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------- |
| M0: Foundation              | DOC-001: docs, directory structure, language comparison, reproducible documentation checks. | Documentation BDD first red then green; docs lint/format/link checks pass; committed clean branch.                        |
| M1: Stack decision          | Equivalent Rust and Go experiments, native terminal/process checks, accepted ADR.           | Same bounded scenarios measured on all three OS families; choose and document stack/toolchain and dependency versions.    |
| M2: Read-only CLI           | FR-002, FR-005, FR-009: repository discovery, status, changes, help/errors.                 | Real temporary-repo BDD covers clean/dirty, staged/unstaged, unborn/detached HEAD, worktrees, and unusual paths.          |
| M3: Files and history       | FR-003, FR-006, FR-007: files, commits, history, Git diffs.                                 | Paginated queries and binary/large-file fallbacks pass tests; no implicit repository mutations.                           |
| M4: Watch                   | FR-001, FR-010: selectable live views, metadata invalidation, reconciliation.               | External Git/editor changes refresh correctly; overflow/fallback/cancellation tests pass.                                 |
| M5: Interactive and editors | FR-004, FR-011: navigation, actions, mode transitions, editor handoff.                      | Terminal restored after child tools/errors; all existing functions accessible; state survives transitions.                |
| M6: Conflict assistance     | FR-008: conflict stages, explicit mergetool/editor launch, stage/unstage.                   | Content/add/delete conflict fixtures pass; report index state and preserve unrelated edits; no implicit continuation.     |
| M7: Release hardening       | NFR-001 through NFR-006; distribution, compatibility matrix, licensing.                     | Native test evidence, supported architecture artifacts, checksums, dependency review, release docs, and license decision. |

Ship vertical slices rather than implementing all parsing, then all UI, then all tests. Within each milestone, update the contract and write one failing executable behavior before implementing that behavior. Do not check off future requirements merely because they have documentation.

## Outstanding decisions

- Rust versus Go, based on the [decision experiment](STACK_COMPARISON.md).
- Exact supported OS, terminal, Git, and compiler versions.
- Configuration schema, editor-command compatibility, and JSON schema/versioning.
- Poll/reconciliation defaults and measured responsiveness/resource budgets.
- License, executable naming/distribution availability, and release signing approach.
