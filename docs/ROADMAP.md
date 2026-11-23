# Delivery roadmap

The [specification](SPECIFICATION.md) defines the requirements. Every milestone uses the [contribution workflow](../CONTRIBUTING.md).

## 1.0.0 delivery sequence

| Milestone | Included in v1                                                           | Acceptance gate                                                                      |
| --------- | ------------------------------------------------------------------------ | ------------------------------------------------------------------------------------ |
| M3        | Files, paginated log/history, commit details, staged and working patches | Real listings, binary and size boundaries, literal paths, empty histories            |
| M4        | Live status, changes, files, commits and conflicts                       | Reconciliation observes external updates, handles stale results and cancels promptly |
| M5        | Interactive command navigation, editor handoff and return to view        | Terminal-required checks, quoted selection, pause/resume and mode transitions        |
| M6        | Conflict stages, mergetool and confirmed stage/unstage                   | Index re-query, explicit deletions and unrelated worktree edits preserved            |
| M7        | Public MIT release on four Windows/Linux targets                         | Quality gates, native execution, versioned ZIPs and SHA-256 manifest                 |

[Plan 0011](plans/0011-v1-release.md) tracks the build and validation for the release. After v1, prioritize macOS native runners, event-assisted refresh, a full-screen selector and enhanced diff presentation. JSON needs its own versioned schema.

## MVP delivery slices

The foundation and [Go decision](adr/0001-application-stack.md) established the stack. The MVP delivered status and changes in [plan 0002](plans/0002-go-mvp-status.md), followed by [compatibility and cancellation](plans/0003-mvp-cli-compatibility.md), [container testing](plans/0004-container-test-workflow.md) and [platform scope](plans/0005-mvp-platform-scope.md). The release process builds on [packaging](plans/0006-mvp-candidate-packaging.md), [native verification](plans/0007-native-candidate-verification.md), [Go notices](plans/0008-mvp-go-notices.md) and [local Gitea CI](plans/0009-local-gitea-ci.md).
