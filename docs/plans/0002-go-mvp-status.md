# Plan 0002: Go MVP and read-only status/changes

## Specification

Implement CLI-001 in the [specification](../SPECIFICATION.md): FR-002, FR-005, FR-009 and applicable NFR-002 through NFR-005 safeguards. MVP 0.1.0 is a read-only CLI; this slice supplies status/changes and leaves release hardening to the next slice. The maintainer's Go decision supersedes the comparative experiment in accepted ADR 0001. No remote push or release is authorized by this task.

## Strategy

1. Start on clean `feat/go-mvp-status`, inspect foundation and installed tools, update the specification/CLI/ADR/roadmap, and record this plan.
2. Introduce pinned Go/module and Godog test infrastructure. Bind a real Gherkin feature to a built executable and isolated temporary Git repositories. Run against an explicit unimplemented CLI stub and record missing-behavior BDD red.
3. TDD the porcelain-v2 parser: branch headers, ordinary/rename/unmerged/untracked/submodule records, partial staging, malformed/truncated input, and literal unusual path bytes. Refactor while green.
4. TDD the bounded/cancellable direct Git process boundary and discovery/status acquisition. Suppress optional locks and fsmonitor helpers; use explicit status flags and isolate fixture Git configuration.
5. TDD argument handling, plain safe rendering, error codes, and stdout/stderr separation; wire the entry point with interrupt cancellation. Rerun executable BDD to green with real clean/dirty/unborn/detached/worktree/rename/conflict/upstream fixtures.
6. Add native OS CI, pinned Staticcheck and vulnerability scan, build instructions, and accurate documentation acceptance. Run quality gates and full suites; review complete staged diff and commit all scoped work cleanly.

Data flows from `cmd/git-view` to `internal/cli`, then `internal/git`, returning a snapshot consumed by plain rendering. Git retains repository semantics; Go does not reimplement diff or merge. No terminal ownership changes occur in this slice. Standard-library CLI parsing avoids adding Cobra for two commands. Godog reuses maintained Gherkin bindings rather than a custom feature parser. Failures return diagnostics without partial snapshots. Changes are locally reversible through Git history.

## Checklist

- [x] Clean named branch and reviewed starting state.
- [x] Specifications and relevant ADRs updated before behavior.
- [x] Ordered implementation steps and acceptance strategy recorded.
- [x] Executable BDD fails for missing behavior, with command/reason recorded.
- [x] Parser TDD red/green and refactor.
- [x] Process/discovery TDD red/green and refactor.
- [x] CLI/rendering TDD red/green and executable BDD green.
- [x] Formatting, vet, selected linter, and all current tests pass.
- [x] User/project documentation aligned and dependencies reviewed.
- [x] Complete diff and actual platform evidence reviewed.
- [x] Meaningful commit created; scoped work committed; branch clean (closeout Git history/status establish this item).

## Validation evidence

Starting state: clean `docs/project-foundation` at `1620835`; created `feat/go-mvp-status`. Installed tools: Go 1.27.4 windows/amd64, Git 2.49.0.windows.1, Node 22.16.0, npm 11.6.4. No remote configured.

- `npm ci`: 171 packages installed, zero audit findings.
- BDD red: `go test ./tests/acceptance -run TestCLI -count=1`: all 16 scenarios failed at the exit-code assertion against the intentional unimplemented executable stub; fixture creation and invocation passed (32 passed steps). No undefined steps or fixture failures established this red.
- Parser red: `go test ./internal/git -run TestParse -count=1`: eight valid-input cases failed with `porcelain v2 parser not implemented`. Parser implementation then passed in the next package run, while the newly added process/discovery tests failed for their explicit unimplemented boundaries.
- Process/discovery red: `go test ./internal/git -count=1`: literal argument, output cap, cancellation/deadline, diagnostic, missing executable, repository identity, and error-mapping expectations failed for unimplemented process/discovery behavior.
- Process/discovery green: corrected the helper's flag separator (a harness defect, not acceptance evidence). Stream-cap tests then caught `bytes.Buffer`'s promoted `ReadFrom` bypassing the capped `Write`; replaced embedding with a private buffer. All process/parser/discovery tests passed, including independent stdout/stderr caps and own/parent deadlines.
- CLI red: `go test ./internal/cli -count=1`: argument, rendering, diagnostic, and exit-code expectations failed against the unimplemented `Run`. Implemented the CLI, corrected the submodule protocol annotation, and refactored command selection while green.
- BDD green: initial 16 scenarios passed after wiring the CLI. Added real staged-add/delete and dirty-submodule fixtures; `go test ./tests/acceptance -count=1 -v` passed all 18 scenarios and 103 steps. Read-only fixture file hashes remained equal. `TestPOSIXPaths` explicitly skipped on Windows; real POSIX control/non-UTF-8/root-whitespace cases remain unverified locally.
- Environment TDD: `go test ./internal/git -run TestProcess/read-only -count=1` failed with inherited lazy-fetch/prompt values `0|1`; forcing `GIT_NO_LAZY_FETCH=1`/`GIT_TERMINAL_PROMPT=0` passed, alongside stripping inherited repository-routing variables. This asserts process configuration, not a real partial-clone network experiment.
- Documentation red: updated documentation acceptance first, then `npm test` failed two scenarios for the stale README/open decision. Documentation alignment made all four scenarios/twelve steps pass.
- Quality green: `npm run format`, `npm run check` (15 Markdown files, zero lint issues, four scenarios), empty `gofmt -l cmd internal tests/acceptance`, `go vet ./...`, Staticcheck v0.8.2, `go mod verify`, `go test ./... -count=1`, and `git diff --check`. Staticcheck initially flagged capitalized error strings; normalized them and reran successfully.
- `go run golang.org/x/vuln/cmd/govulncheck@v1.10.0 ./...`: no vulnerabilities found for this graph/toolchain at execution time. `npm ci` audit also reported zero findings. Production builds have no non-standard-library Go dependencies; Godog v0.16.0 is test-only (MIT), with no added cgo requirement. Reviewed module license files: Cucumber modules MIT, HashiCorp modules MPL-2.0, UUID/pflag/difflib BSD-style, go-spew ISC, testify MIT, yaml.v3 MIT/Apache. Some transitive modules are older; pin/check the graph and revisit during dependency upgrades. No app license is selected by adopting test tooling.
- Built `bin/git-view.exe` and smoke-tested version, help, and status on this checkout. Compiled Windows/Linux/macOS for amd64 and arm64 into ignored `bin/` artifacts; these are compilation checks, not native runtime evidence or release packages.
- `go test -race ./...` could not start: `-race requires cgo`; local `CGO_ENABLED=0`. Linux CI is configured for the race detector. No interactive terminal signal smoke test, native macOS/Linux run, ARM64 runtime, or remote CI run occurred.

Reviewed the complete staged diff against actual commands, dependencies, tests, paths, and documented scope. No unrelated starting work existed. The closeout commit and final clean status provide commit evidence without embedding this file's own commit hash. No push or release was performed.

## Risks and open decisions

- Native macOS/Linux and ARM64 runtime evidence cannot be inferred from Windows tests or cross-builds. CI configuration is not a passing CI run.
- Read-only Git is bounded to 10 seconds and 16 MiB per output stream; larger repositories may fail explicitly. Optimize only with measured evidence.
- Repository changes between discovery and status are not atomic. Live reconciliation belongs to Watch; one-shot failures remain explicit.
- No process-tree cleanup guarantee for future foreground tools is established here. Automatic status disables configured fsmonitor helpers.
- Git minimum versions and supported OS versions need release validation. A versioned JSON contract, color, license, and distribution/signing remain future work.
