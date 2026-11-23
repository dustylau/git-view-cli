# Plan 0011: Git View 1.0.0

## Specification

Implement [V1-001](../SPECIFICATION.md#v1-001-public-100-release), FR-001 through FR-011 and the applicable quality requirements. Use the [command contract](../CLI.md) and [terminal decision](../adr/0006-v1-terminal.md). The maintainer authorized a public MIT release.

## Strategy

1. Start a clean feature worktree at main. Define the v1 contract and milestone gates.
2. Bind executable scenarios for listing files, history and patches, conflict stages, exact-path mutations and terminal requirements. Capture their red result.
3. Drive command parsing and safe Git queries with focused tests. Reuse the existing process budget, repository identity and safe rendering.
4. Exercise index operations and editor/mergetool handoff against temporary repositories, including unborn, partial staging, deletion and literal path cases.
5. Drive Watch and Interactive navigation, stale results, terminal detection, pause/resume and cancellation through focused tests and a real terminal fixture.
6. Align docs, packaging and version. Verify licenses, local quality gates and candidate installation, then commit all scoped work.
7. Run native source and candidate workflows on the release revision. Verify archives and publish v1.0.0 with checksums and release notes.

## Checklist

- [x] Clean named feature worktree, specification, ADR and plan.
- [x] Bound Gherkin fails on missing commands.
- [x] Files, commits and diffs pass red/green tests.
- [x] Index operations and foreground tools pass red/green tests.
- [x] Watch and Interactive pass red/green tests.
- [x] Documentation, MIT packaging, quality checks and diff review.
- [x] Commit scoped work to the feature branch. Git history and status establish closeout.

The [v1.0.0 release record](https://github.com/dustylau/git-view-cli/releases/tag/v1.0.0) carries the external signoff, workflow links and verified assets.

## Validation evidence

Started from a1579b9 on feat/v1-release in a new worktree. The original checkout contains a pre-existing untracked nested repository.

BDD red: `go test ./tests/acceptance -run '^TestCLI$' -count=1` failed all 13 new v1 scenarios on unknown commands after successful fixture setup. Parser, browse, action and live-session tests first failed at their missing implementation boundaries. After implementation, the Go suite passed on Windows/amd64. Regression tests also caught a staged deletion masked by an untracked replacement and staging a child implicitly replacing an indexed parent. Both cases now pass.

Documentation BDD first failed on the missing version and installation guide. Packaging acceptance first failed on the missing MIT license. The four candidate archives subsequently passed checksum, format, license and native Windows executable checks. A local CI regression test verifies both MVP and v1 artifact names against the selected revision.

- Windows 11 amd64, Go 1.27.4, Git 2.49.0.windows.1. `npm ci`, `npm run check`, gofmt, vet, module verification, Staticcheck v0.8.2, govulncheck v1.10.0, the Go suite, the Python container/release/CI unit suites and whitespace checks passed. The vulnerability scan reported no findings.
- The native Windows console test exercised external worktree changes, pause/resume, return to the active view, foreground input, stage confirmation, console input mode preservation and Ctrl+C child cleanup.
- Both seeded and host-mounted Docker acceptance scenarios passed, including the Linux/amd64 race suite. Real Git fixtures cover editor/mergetool index re-query, content/add/delete conflicts and preservation of merge state.
- The Windows benchmark recorded small-fixture p95 of 359.5 ms for status and 362.2 ms for changes. Medium-fixture p95 was 289.4 ms and 290.1 ms respectively. Both fixture budgets passed.

## Risks and open decisions

Watch reconciles on a two-second timer. Foreground GUI editors need an explicit wait argument. Public distribution targets Windows and Linux on amd64 and arm64. macOS distribution, event-driven refresh and advanced presentation follow v1.
