# Plan 0004: Container build and test workflow

## Specification

Implement [DEV-001](../SPECIFICATION.md) for FR-002, FR-005, FR-009, NFR-001, NFR-002, and NFR-006. A developer can build and test the CLI on Linux with a new seeded Git volume or a read-only host checkout. The [fixture decision](../adr/0003-container-test-fixtures.md) defines the image and data boundary.

## Strategy

1. Branch from clean CLI-002. Document the contract and verify Docker and the pinned Go image.
2. Add Godog scenarios that launch the workflow with both fixture choices. Demonstrate red before implementing.
3. Use focused runner tests to drive mount construction, volume cleanup after failure, and host directory selection.
4. Build the image and run both scenarios against the Docker engine. Verify staged, unstaged, and untracked output and host fixture integrity.
5. Document prerequisites and commands. Run repository checks, review the diff, and commit.

## Checklist

- [x] Clean named branch and reviewed starting state.
- [x] Specification, fixture decision, and plan written.
- [x] Executable BDD red and focused TDD red/green.
- [x] Both Docker fixture modes verified.
- [x] Project checks and documentation alignment.
- [x] Scoped commit and clean branch. Git history and status record closeout.

## Validation evidence

Started from clean `021efac` on Windows/amd64. Docker Engine availability and container runtime results are tracked below.

BDD red: with `GIT_VIEW_CONTAINER_TEST=1`, `go test ./tests/acceptance -run '^TestContainerWorkflow$' -count=1` reached the Python runner in both scenarios. Both failed with the intentional unimplemented workflow diagnostic.

- TDD red and green via `py -3 -B -m unittest discover -s tests/container`. Four tests initially failed on the unimplemented runner. All four passed after implementation. They exercise mount quoting, missing-path handling, the seeded lifecycle, and cleanup when either seeding or testing raises an error.
- BDD green via `go test ./tests/acceptance -run '^TestContainerWorkflow$' -count=1 -v` with the same environment flag. Both scenarios and all 13 steps passed from Windows PowerShell and Ubuntu 24.04 WSL2 against Docker Engine 29.7.2 on Linux/amd64. Both modes ran the offline Go race suite. Host fixture hashes matched, and the runner's volume label had no remaining volumes.
- The built image reported Go 1.27.4, Git 2.47.3, and `CGO_ENABLED=1`. A direct container check confirmed `/src` excludes `.git`, `node_modules`, and `bin`. The pinned Go image digest is a multi-platform image index. This run exercised Linux/amd64.
- `npm ci`, `npm run check`, gofmt, `go mod verify`, `go vet ./...`, `go test ./... -count=1`, Staticcheck v0.8.2, govulncheck v1.10.0, and `git diff --check` passed. The Go dependency graph is unchanged.
- The existing Ctrl+C smoke tests passed on Windows and WSL2. Both container scripts passed `sh -n` under WSL2.

## Risks and open decisions

- Host fixtures need Git metadata reachable inside the mounted working tree. Docker Desktop must be able to share the chosen path.
- The first image build downloads the base image and Go modules. Runtime fixture checks use an isolated network namespace.
- Container architecture follows the Docker Engine's platform. The hosted native release matrix remains separate.
