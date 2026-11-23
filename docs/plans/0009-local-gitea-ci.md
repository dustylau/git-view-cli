# Plan 0009: Local Gitea CI and delivery

## Specification

Implement [DEV-002](../SPECIFICATION.md#dev-002-local-ci-and-delivery) and [ADR 0005](../adr/0005-local-gitea-ci.md).

## Strategy

1. Check Docker and read existing build/test/candidate workflows. Establish executable BDD red for the missing Compose stack.
2. Drive run-result matching through a focused bootstrap test. Add the Compose services, idempotent provisioning and read-only seed, and both host launchers.
3. Push a committed fixture to the local instance and capture Actions results. Verify the downloaded deliverables, repeated bootstrap, and checkout preservation.
4. Run project gates, align documentation, review staged changes, and commit. Bring up the default stack from the resulting branch.

## Checklist

- [x] Clean named branch, specification, ADR, and plan.
- [x] Executable BDD and focused TDD red and green recorded.
- [x] Local push, container checks, artifact delivery, and repeat bootstrap verified.
- [x] Project quality gates and documentation alignment.
- [x] Scoped commit and clean branch. Git history and status record closeout.

## Validation evidence

Started from clean `7d53c8a` on `chore/local-gitea-ci`. Docker Engine 29.7.2 runs Linux/amd64 on the Windows host.

BDD red: `GIT_VIEW_LOCAL_CI_TEST=1` with `go test ./tests/acceptance -run '^TestLocalPipeline$' -count=1` reached Docker Compose and failed with `no configuration file provided`.

TDD red: `python3 -B -m unittest discover -s tools/ci` in the existing Go test image failed at the unimplemented push-run matcher.

- The first fresh instance exceeded its health check window during SQLite setup. WAL and a longer startup window let a fresh instance initialize and register the runner. The legacy v3 upload action reported success, while the REST API listed no artifact. The pinned Gitea-compatible v4 action made it available for REST download.
- The bootstrap unit test and the isolated Compose acceptance scenario passed. A push of the staged test commit triggered the Linux/amd64 job, ran the documentation gate, Go vet and static analyzers, race and acceptance tests, Python unit suites, and POSIX Ctrl+C smoke. The job verified the four packages before uploading. Bootstrap downloaded the server artifact, checked all four hashes, and returned the same run URL on a repeat push.
- Host `npm run check`, gofmt, `go mod verify`, `go vet ./...`, `go test ./... -count=1`, Staticcheck v0.8.2, govulncheck v1.10.0, both existing Docker fixture scenarios, and whitespace checks passed. Compose configuration, shell syntax, PowerShell parsing, and Actionlint v1.7.12 also passed. Actionlint used the custom runner label from Compose.

## Risks and open decisions

The first setup needs network access to the pinned images, public actions, and build dependencies. Native Windows and ARM64 validation requires the platform runners in the release plan.
