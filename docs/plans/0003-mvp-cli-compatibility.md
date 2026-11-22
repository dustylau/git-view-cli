# Plan 0003: MVP CLI compatibility

## Specification

CLI-002 hardens `status` and `changes` under FR-002, FR-005, FR-009, and NFR-001 through NFR-006. The [specification](../SPECIFICATION.md) defines Git compatibility, partial-clone preservation, terminal cancellation, and performance acceptance. Windows, Linux, and macOS on amd64 and arm64 are release targets. MVP 0.1.0 uses plain output.

## Strategy

1. Start on `feat/mvp-cli-compatibility` from the clean CLI-001 commit. Update the contract and record the local platform inventory.
2. Extend the executable Gherkin feature with an upstream configuration override, branch name ambiguity, and a filtered local clone missing a blob. Demonstrate BDD red against the existing binary.
3. Drive the Git invocation, transport policy, and branch detection with focused tests, then run the full fixture suite. Preserve the existing read boundary and use Git plumbing for branch identity.
4. Send Ctrl+C in a POSIX pseudo-terminal and isolated Windows console using a controlled Git child. Verify exit 130, no snapshot, and immediate child cleanup. Run local environments and configure native CI for all six targets.
5. Measure small (100 tracked, 10 untracked) and medium (5,000 tracked, 1,000 untracked) repositories. Include tracked edits, exclude fixture/build time, use three warmups and 20 samples per command, and record median and p95. Compare with Git porcelain status using the same read options. Target p95 at or below 400 ms and 1 s respectively for both CLI commands.
6. Run the project gates, update docs with measured results, review the diff, and commit the branch cleanly.

## Checklist

- [x] Clean named branch and reviewed starting state.
- [x] Specifications and implementation plan updated.
- [x] Demonstrate executable BDD red.
- [x] Git compatibility TDD red and green, followed by BDD green.
- [x] Terminal smoke tests and performance measurements on the available Windows and WSL2 environments.
- [x] Formatting, static analysis, and project tests pass.
- [x] Project and user docs match implementation and platform evidence.
- [x] Review diff and commit scoped work on a clean branch. Git history and status record closeout.
- [ ] Record native hosted runner results, including macOS and arm64.

## Validation evidence

Started from clean `8e37b74`. Available environments are Windows/amd64 with Git 2.49.0.windows.1 and Ubuntu 24.04 under WSL2 with Git 2.43.0 on amd64.

- BDD red on Windows, `go test ./tests/acceptance -run TestCLI -count=1`. Both literal `(detached)` branch scenarios reached the executable and failed their branch summary assertions. The 19 other scenarios passed, including the upstream configuration and filtered-clone fixtures.
- BDD red repeated on WSL2 using a temporary export of `8e37b74` and the new fixtures. The same two scenarios failed on Git 2.43.0.
- TDD red, `go test ./internal/git -count=1`. The process inherited the caller's transport allowlist, status omitted the explicit counts flag, and both named branch variants still appeared detached.
- TDD green after setting the empty transport allowlist, requesting `--ahead-behind`, and resolving ambiguous HEAD through Git's symbolic ref. `go test ./... -count=1` passed on Windows 11 amd64 and Ubuntu 24.04 WSL2 amd64 with Go 1.27.4. All 21 Gherkin scenarios passed on each. The POSIX path test passed on Linux and skipped on Windows. The partial-clone fixture checked that the source blob was absent and that inspection preserved the fixture's Git and worktree contents.
- `terminal_smoke.py` passed with Python 3.12 on Windows and Python 3.12 under WSL2. Ctrl+C during a pending read produced exit 130 with a cancellation diagnostic and no snapshot. The Windows console and Linux pseudo-terminal both confirmed immediate child exit.
- `npm ci`, `npm run check`, gofmt, `go vet ./...`, Staticcheck v0.8.2, `go mod verify`, and `git diff --check` passed. npm reported zero audit findings and govulncheck v1.10.0 found none. Actionlint v1.7.12 passed on the Go workflow. The Go module graph is unchanged.
- Compiled Windows, macOS, and Linux for amd64 and arm64. No remote is configured. Native hosted jobs and macOS/arm64 runtime evidence are pending. The local WSL2 race attempt stopped before execution because cgo was disabled and the distro has no C compiler. Linux CI configures the race run.

### Warm command timing

`measure_cli.py` ran on the same Ryzen 9 5950X host on Windows 11 build 26100 and Ubuntu 24.04 WSL2. The fixtures had 10% unstaged tracked edits plus the listed untracked files. The table reports median / p95 in milliseconds for 20 measured invocations after three warmups. Both CLI commands passed the 400 ms small and 1 s medium p95 budgets.

| Environment                   | Fixture                |         Git |      `status` |     `changes` |
| ----------------------------- | ---------------------- | ----------: | ------------: | ------------: |
| Windows, Git 2.49.0.windows.1 | Small (100 / 10)       | 38.4 / 41.3 | 207.5 / 215.6 | 205.5 / 215.0 |
| Windows, Git 2.49.0.windows.1 | Medium (5,000 / 1,000) | 85.6 / 91.1 | 257.7 / 264.2 | 252.5 / 268.8 |
| WSL2, Git 2.43.0              | Small (100 / 10)       |   3.9 / 4.1 |   17.9 / 18.3 |   18.0 / 18.7 |
| WSL2, Git 2.43.0              | Medium (5,000 / 1,000) | 11.5 / 11.9 |   26.1 / 27.5 |   26.3 / 28.8 |

## Risks and open decisions

- Git releases differ in how early they suppress lazy fetch. The transport guard must work even when Git attempts a fetch.
- A host CLI timing depends on its Git build, filesystem, cache and antivirus state. The repeatable warm fixture gives a baseline for later comparison.
- Native macOS and arm64 results require external runners. Configuration and cross-compilation establish separate evidence.
