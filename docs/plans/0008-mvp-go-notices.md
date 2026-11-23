# Plan 0008: Go notices in MVP archives

## Specification

Include the Go toolchain notices with each [REL-001 local candidate](../SPECIFICATION.md#rel-001-local-candidate-packaging). The project license still requires the maintainer's decision. See [ADR 0004](../adr/0004-mvp-packaging.md).

## Strategy

1. Extend release BDD to compare the notices in all four archives with the installed toolchain. Establish red against an existing candidate.
2. Drive missing and empty notice handling with focused packager tests. Load both notices before creating output and carry them through deterministic archive writing.
3. Build candidate archives on Linux and verify them on the available platforms. Repeat packaging to check archive reproducibility.
4. Run the repository gates, align documentation, review the staged change, and commit it.

## Checklist

- [x] Clean named branch, specification, ADR, and plan.
- [x] BDD and focused packager red and green verified.
- [x] Four archives and repeat build verified.
- [x] Quality gates, documentation, and staged diff reviewed.
- [x] Scoped commit and clean branch. Git history and status record closeout.

## Validation evidence

Started from clean `ba42c9b` on `chore/mvp-go-notices`.

Red on Windows/amd64: the opt-in `TestReleasePackaging` against `0.1.0-rc.2` failed because the first archive lacked `licenses/go/PATENTS`. `py -3 -B -m unittest discover -s tools/release` failed both new tests because the packager reached the build with a missing `LICENSE` or blank `PATENTS`.

- All five packager unit tests passed after the change. The release scenario passed with a fresh build on Windows/amd64. A test copy with an altered license and regenerated checksums failed the byte comparison.
- Two independent `0.1.0-rc.3` builds in the Linux/amd64 Docker environment produced byte-identical ZIPs and manifests. `sha256sum -c SHA256SUMS` passed for all four archives. The retained candidate passed verification on Windows/amd64, Linux/amd64 Docker, and emulated Linux/arm64 Docker. Windows/arm64 was checked at the archive and executable-header level.
- `npm ci`, `npm run check`, gofmt, `go mod verify`, `go vet ./...`, `go test ./... -count=1` with release acceptance enabled, Staticcheck v0.8.2, govulncheck v1.10.0, Windows Ctrl+C smoke, both Python unit suites, both real Docker fixture scenarios, Actionlint v1.7.12, and `git diff --check` passed.

## Risks and open decisions

Hosted validation still requires authorized runner access. The remaining release decisions include the git-view license, release name, and publication approval.
