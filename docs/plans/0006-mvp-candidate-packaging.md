# Plan 0006: MVP candidate packaging

## Specification

Prepare local candidate packages for [REL-001](../SPECIFICATION.md#rel-001-local-candidate-packaging) and NFR-001, NFR-005, and NFR-006. The [packaging decision](../adr/0004-mvp-packaging.md) specifies ZIP archives, checksums, and version stamping. Native validation, licensing, and publication remain release gates.

## Strategy

1. Start on a clean branch from the narrowed MVP platform scope. Record the archive contract.
2. Add executable acceptance for a four-target package build and version smoke test. Run against the unimplemented command.
3. Drive version validation, failure cleanup, and preservation of an existing candidate through focused runner tests. Build and inspect all four binaries through the real packaging command.
4. Check repeated builds for matching archive hashes. Run host-native and emulated Linux candidate smoke tests where available.
5. Document packaging and installation, run project quality gates, review, and commit.

## Checklist

- [x] Clean named branch and specification, ADR, and plan.
- [x] Executable BDD red and TDD red/green.
- [x] Verified four architecture headers, checksums, and candidate version.
- [x] Reproducibility and accessible runtime smoke checks.
- [x] User docs and quality gates.
- [x] Scoped commit on a clean branch. Git history and status record closeout.

## Validation evidence

Started from clean `e29a317` on `chore/mvp-release-packaging`. The checkout has no configured Git remote. The [compatibility plan](0003-mvp-cli-compatibility.md) tracks native validation.

BDD red: with `GIT_VIEW_RELEASE_TEST=1`, `go test ./tests/acceptance -run '^TestReleasePackaging$' -count=1` reached the packaging command and failed with the intentional unimplemented diagnostic.

TDD red and green: `py -3 -B -m unittest discover -s tools/release` initially failed all three tests against the stub. The implemented command passed input validation, preservation of existing output, and cleanup after a simulated build failure. An intermediate BDD run built four correctly labeled archives with matching checksums, then failed because the constant version stayed at `0.1.0-dev` in the extracted executable.

- After the version became linker-settable, the BDD scenario and all four steps passed on Windows/amd64. It inspected both PE headers and both ELF headers, Linux executable modes, archive contents, all four hashes, and the installed executable's `--version`.
- `py -3 tools/release/package_release.py --version 0.1.0-rc.1` generated the candidate in `dist/git-view-0.1.0-rc.1/`. A second build into an independent temporary location matched every archive hash. `sha256sum -c SHA256SUMS` verified all four archives from the retained candidate.
- The extracted Windows/amd64 executable passed `--version` and `status` on a temporary checkout. The Linux/amd64 executable passed the same checks in Docker on amd64. Linux/arm64 passed those checks under Docker emulation. Windows/arm64 evidence is limited to the executable header and archive checks.
- `npm ci`, `npm run check`, gofmt, `go mod verify`, `go vet ./...`, `go test ./... -count=1`, Staticcheck v0.8.2, govulncheck v1.10.0, Python unit suites, Windows Ctrl+C smoke, and `git diff --check` passed.
- Both real Docker scenarios passed with `GIT_VIEW_CONTAINER_TEST=1`. They exercised the Go race suite with the seeded volume and host checkout.

## Risks and open decisions

- Artifact inspection and emulated smoke tests have limited platform scope. Native checks remain a release requirement.
- Packaging produces local files. The maintainer still needs to settle the license, final distribution identity, and publication approval.
