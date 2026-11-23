# Testing and verification

## Current executable suite

Use Go 1.27.4, Git 2.43.0 or newer, and Node 22.16+ in the Node 22 line or Node 24 LTS with npm. Node serves the documentation suite. Application BDD uses Godog v0.16.0. Python 3.9 or newer runs the terminal and performance checks. See [plan 0003](plans/0003-mvp-cli-compatibility.md) for results and platform limits.

```sh
npm ci
npm run check
gofmt -l cmd internal tests/acceptance
go vet ./...
go test ./...
go run honnef.co/go/tools/cmd/staticcheck@v0.8.2 ./...
go run golang.org/x/vuln/cmd/govulncheck@v1.10.0 ./...
git diff --check
```

`gofmt -l` must produce no output; use `gofmt -w` to fix formatting. It does not itself return failure for listed files, so CI enforces the empty result explicitly. `npm run check` combines Prettier, Markdown lint, step-binding JavaScript syntax, and four documentation BDD scenarios. `npm run format` formats documentation/tooling files. Go uses gofmt independently.

`go test ./...` includes parser, process-boundary, and CLI unit tests plus Godog acceptance. To focus on executable BDD, run `go test ./tests/acceptance -run TestCLI -count=1 -v`. It builds a fresh binary, binds real Gherkin, and invokes the binary directly to assert exit codes and stream separation. Undefined/pending steps fail under strict mode. Fixtures isolate Git system/global configuration and identity, use synthetic data and local refs, and require no network. Read-only scenarios hash fixture files, including Git metadata, before and after invocation. This verifies content preservation, not the absence of every filesystem event.

The documentation link scenario checks inline relative file links in project Markdown, skipping external URLs and fenced examples. It verifies file existence, not remote availability, fragment anchors, reference-style links, or semantic correctness. Human review remains necessary. Original DOC-001 evidence remains in [plan 0001](plans/0001-documentation-foundation.md).

## Red/green

Before implementing behavior, run the new executable feature and confirm failure for the intended missing behavior. Fix harness errors first. For CLI-001, an intentional CLI stub established BDD red; parser fixtures and process/discovery/CLI expectations then drove separate unit red/green loops. Documentation scenarios were updated to require the accepted Go decision and distinguish implemented commands from planned modes, and demonstrated red before updating the prose.

A `.feature` file without executable bindings is not coverage. Skipped scenarios cannot establish acceptance. Do not write future-product scenarios into the passing suite and call those features implemented.

## Coverage and remaining test layers

| Layer                | Current evidence or future gate                                                                                                                                        |
| -------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Unit/process         | Porcelain records, malformed/truncated input, literal unusual path bytes, safe rendering, argument errors, stream caps, deadlines, cancellation.                       |
| CLI acceptance       | Real clean/dirty, unborn/detached, staged/unstaged, deleted/renamed/untracked/ignored, conflict, upstream, submodule, worktree, help/version/error fixtures.           |
| Platform paths       | Spaces, Unicode, leading dashes on Windows; parser/render tests cover tabs/newlines/control/non-UTF-8 bytes. `TestPOSIXPaths` requires native POSIX and skips Windows. |
| Race detector        | Run `go test -race ./...` where cgo/a C compiler are available. CI configures this on Linux. Report a missing local compiler explicitly.                               |
| Watch                | Future: external changes, atomic saves, metadata events, overflow, polling, stale results, Ctrl+C, process-tree cleanup.                                               |
| Interactive/terminal | Future: navigation, resize, native key events, editor handoff, cancellation, mode transitions, terminal restoration.                                                   |
| Conflict assistance  | Current listing/counts only; future tool launches and index actions require preservation, stage-state, explicit-resolution tests.                                      |

Use bounded eventual assertions for real OS events rather than sleeps. Keep pure unit tests deterministic through injected boundaries. Each later slice must add tests for its behavior; CLI tests cannot prove terminal ownership or watcher behavior.

## CI and platform claims

The [Go workflow](../.github/workflows/go.yml) checks the Git floor and native Go toolchain on four MVP runner targets: Windows and Linux, each on amd64 and arm64. It configures the Go suite and terminal smoke test on each target, plus the race detector and advisory scan on Linux. See [GitHub's runner labels](https://docs.github.com/en/actions/reference/runners/github-hosted-runners). Local results cover Windows/amd64 and Linux/amd64 under WSL2. The Linux race suite passed under WSL2 with cgo and GCC. Hosted jobs and native arm64 results are pending. macOS support is scheduled for a later release.

Run the terminal check on Windows with `py -3 tests/acceptance/terminal_smoke.py` and on Linux with `python3 tests/acceptance/terminal_smoke.py`. The script builds the CLI and a controlled Git child in a temporary directory. It sends Ctrl+C after the child signals readiness, checks exit 130 and the cancellation diagnostic, and verifies that the child has exited. Windows uses an isolated console with captured streams. Linux uses a pseudo-terminal.

Run `py -3 tests/acceptance/measure_cli.py` on Windows or `python3 tests/acceptance/measure_cli.py` on Linux. It times 20 subprocess invocations after three warmups per command on each fixture, and includes the equivalent Git query as a baseline. It reports median and p95 and checks the 400 ms small-fixture and 1 s medium-fixture p95 budgets. CI keeps timing separate from the correctness gates because runner load varies.

## Container build and test

Use Docker Engine 25 or newer with Linux containers and Python 3.9 or newer on the host. Docker Desktop must have access to any host path you choose to mount. Run from the project checkout.

```sh
python3 tests/container/container_workflow.py
python3 tests/container/container_workflow.py --repo "/path/to/existing checkout"
```

On Windows, use `py -3` instead of `python3` and a Windows path such as `D:\Development\git-view-cli`. Use paths visible to the shell running the command and its local Docker Engine. From WSL, use Linux paths such as `/mnt/d/Development/git-view-cli`.

Every run builds or reuses cached layers for `git-view:test`. The build copies Go source and test files into a pinned Go 1.27.4 image. The first build downloads the base image and Go modules. The mounted repository is supplied at runtime. Both fixture modes run with the network disabled, print `status` and `changes`, and run `go test -race -mod=readonly ./... -count=1`.

- **Seeded volume.** The runner creates a new named volume, seeds a commit and mixed changes, then mounts the resulting checkout read-only at `/repo`. It asserts the expected output and removes the volume after success or an ordinary test failure. The image tag and build cache remain available.
- **Host checkout.** Supply a worktree root with Git metadata available within that directory. The runner mounts it read-only at `/repo` and permits that path through Git's ownership check for the test process only. Checkout layouts that rely on Git directories or nested mounts outside the selected root require the seeded mode or a standalone checkout.

Run the host-side runner tests with `python3 -B -m unittest discover -s tests/container`. To run the two real Docker acceptance scenarios, set `GIT_VIEW_CONTAINER_TEST=1` and run `go test ./tests/acceptance -run '^TestContainerWorkflow$' -count=1 -v`. The tests check seeded cleanup and compare the host fixture's worktree and Git metadata before and after. This workflow supplies Linux container evidence. Native platform gates are tracked in [plan 0003](plans/0003-mvp-cli-compatibility.md).

## Release candidate packaging

Run the packaging unit tests with `python3 -B -m unittest discover -s tools/release`. To build and inspect all four candidate archives in an acceptance test, set `GIT_VIEW_RELEASE_TEST=1` and run `go test ./tests/acceptance -run '^TestReleasePackaging$' -count=1 -v` on a Windows or Linux host. The scenario checks each PE or ELF architecture, Linux executable permissions, archive contents, and checksums. It compares the bundled Go notices byte for byte with the pinned toolchain files. It runs the extracted host executable with `--version`, `status`, and `changes` and compares the worktree and Git metadata before and after each repository read.

To check a prebuilt candidate from the project root, set `GIT_VIEW_RELEASE_DIR` to the absolute path of the directory containing the four ZIPs and `SHA256SUMS`. Set `GIT_VIEW_RELEASE_VERSION` to the expected version. For example, in PowerShell:

```powershell
$env:GIT_VIEW_RELEASE_TEST = '1'
$env:GIT_VIEW_RELEASE_VERSION = '0.1.0-rc.3'
$env:GIT_VIEW_RELEASE_DIR = (Resolve-Path 'dist/git-view-0.1.0-rc.3').Path
go test ./tests/acceptance -run '^TestReleasePackaging$' -count=1 -v
```

On Linux, export the same variables with `GIT_VIEW_RELEASE_DIR="$PWD/dist/git-view-0.1.0-rc.3"`. With `GIT_VIEW_RELEASE_DIR` unset, the test builds into a temporary directory. `GIT_VIEW_RELEASE_VERSION` defaults to `0.1.0-rc.1`.

After the [Release candidate workflow](../.github/workflows/release-candidate.yml) is available in the GitHub repository's default branch, dispatch it from the Actions tab with a branch and version. The build job uploads one artifact. All four verification jobs download it by ID and validate the same archive bytes. Use the workflow run and its revision to record candidate evidence. See [GitHub's manual dispatch requirements](https://docs.github.com/en/actions/how-tos/manage-workflow-runs/manually-run-a-workflow).

Use the [packaging guide](../README.md#package-a-local-mvp-candidate) to generate files for local review and the [installation guide](INSTALL.md) to verify and run them. The candidate check complements the source-level compatibility checks in [plan 0003](plans/0003-mvp-cli-compatibility.md). See [plan 0007](plans/0007-native-candidate-verification.md) for actual candidate verification results.

## Local pipeline acceptance

Use the [local CI guide](LOCAL_CI.md) to run the pipeline from an ordinary checkout with Docker Compose as the host prerequisite. For the end-to-end regression, set `GIT_VIEW_LOCAL_CI_TEST=1` and run `go test ./tests/acceptance -run '^TestLocalPipeline$' -count=1 -timeout=25m -v`. This starts an isolated Compose project on a free loopback port, pushes a committed revision, verifies the delivered artifact over HTTP, and removes the test project's containers and volumes. The test selects `HEAD` by default. Set `GIT_VIEW_LOCAL_CI_REF` to verify a different committed revision that contains the workflow.

## Dependency review

Production builds use only Go's standard library and installed Git. Godog v0.16.0 is a test dependency (MIT, no cgo); its transitive graph is recorded in `go.mod`/`go.sum`. Staticcheck v0.8.2 and govulncheck v1.10.0 run at pinned versions without becoming application dependencies. Run `go mod tidy`, `go mod verify`, and the pinned vulnerability command when changing the Go graph. Advisory results are time-specific, not permanent guarantees.

Use `npm audit` when changing documentation tooling. The foundation's scoped override pins Markdown lint's transitive KaTeX to `0.18.2` for the [KaTeX advisory](https://github.com/advisories/GHSA-238p-pmpm-9mq7). This crosses the math extension's original range; its initial renderer smoke test and lint suite passed. Revisit when upstream includes the fix. Node packages do not ship in the CLI.
