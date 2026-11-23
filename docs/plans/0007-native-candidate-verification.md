# Plan 0007: Native candidate verification

## Specification

Implement [REL-001 native candidate verification](../SPECIFICATION.md#rel-001-native-candidate-verification) for NFR-001 and NFR-002. Feed one versioned artifact set to each runner. Keep source-level native tests and release approval as separate gates.

## Strategy

1. Extend the bound release feature to use an existing candidate and inspect the extracted host executable on the existing mixed-change Git fixture. Establish red with an existing candidate directory.
2. Enable prebuilt candidate selection, verify that missing and damaged candidates fail, and drive the installed executable checks through the same acceptance scenario.
3. Add a dispatch-only workflow that builds once, passes the artifact by its upload ID, and runs verification on all four native runner targets. Check local artifact verification on Windows and Linux and lint the workflow.
4. Check the full repository, align documentation, review and commit.

## Checklist

- [x] Clean named branch, specification, ADR, and ordered plan.
- [x] Executable BDD red and focused red/green results recorded.
- [x] Prebuilt candidate verified on available Windows and Linux environments.
- [x] Native verification workflow defined and linted.
- [x] Quality gates, documentation, and staged diff reviewed.
- [x] Scoped commit on a clean branch. Git history and status record closeout.

## Validation evidence

Started from clean `b51bd0c` on `chore/native-release-candidate`.

Red on Windows/amd64: `go test ./tests/acceptance -run '^TestReleasePackaging$' -count=1 -v` with `GIT_VIEW_RELEASE_TEST=1` and `GIT_VIEW_RELEASE_DIR` set to the retained `0.1.0-rc.1` candidate reached the prebuilt selection step and failed at its unimplemented diagnostic. With no supplied directory, the run passed packaging, archive, checksum, and version steps, then failed at the unimplemented mixed-worktree check.

- Both modes passed after implementation. A missing supplied directory was left absent, a directory for a different requested version was rejected, and changing ZIP metadata without updating the manifest failed at the checksum gate. The failure cases stopped before invoking the executable.
- One `0.1.0-rc.2` candidate was built in a network-disabled Linux/amd64 Docker container and reused on Windows/amd64, Linux/amd64 Docker, and emulated Linux/arm64 Docker. All five release acceptance steps passed on each. The Linux/amd64 run also passed `go test -race -mod=readonly ./... -count=1` with candidate verification enabled. Windows/arm64 was checked at the archive and executable-header level.
- `sha256sum -c SHA256SUMS` verified all four archives. `go version -m -json` on the four extracted executables reported Go 1.27.4, cgo disabled, and no linked external Go modules. Git remains a runtime prerequisite.
- `npm ci`, `npm run check` (five documentation scenarios), gofmt, `go mod verify`, `go vet ./...`, `go test ./... -count=1`, Staticcheck v0.8.2, govulncheck v1.10.0, the Python unit suites, Windows terminal smoke, both opt-in Docker acceptance scenarios, and `git diff --check` passed.
- Actionlint v1.7.12 passed. Upload and download action pins matched their upstream tags. The Windows verification step passed when run locally with the same PowerShell commands and environment variables as the workflow.

## Risks and open decisions

The checkout has no configured remote. Hosted job execution and native ARM64 checks require runner access. Licensing, final distribution identity, and publication remain open.
