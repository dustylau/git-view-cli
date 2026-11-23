# Plan 0005: MVP platform scope

## Specification

Apply the maintainer's 2026-11-23 decision to defer macOS from the initial MVP. The [MVP specification](../SPECIFICATION.md#mvp-010) retains Windows and Linux on amd64 and arm64. NFR-001 still requires native runtime evidence for each release target.

## Strategy

1. Start from the clean container workflow commit. Update the release specification and link the platform decision from the stack ADR.
2. Add executable documentation acceptance for the revised support scope and demonstrate red while the README and testing guide still describe the old release matrix.
3. Align the Go and documentation workflows, README, roadmap, and CLI-002 validation checklist with the revised release scope. Keep supplemental container evidence separate from native runner results.
4. Validate the workflow and project checks, review the diff, and commit.

## Checklist

- [x] Clean named branch and reviewed starting state.
- [x] Specification, ADR, and plan updated.
- [x] Executable documentation acceptance red and green.
- [x] Four Go CI targets and public support docs agree.
- [x] Project checks pass and scoped work is committed. Git history and status record closeout.

## Validation evidence

Started from clean `f495462` on `chore/mvp-platform-scope`. The checkout has no Git remote. Native runner results remain an open item in [plan 0003](0003-mvp-cli-compatibility.md).

BDD red: `npm test` ran five scenarios. The new platform scope scenario failed because the README still listed the original targets. The other four scenarios passed.

- BDD green: `npm run check` passed all five documentation scenarios and 14 steps with formatting, lint, and local link checks.
- Actionlint v1.7.12 passed on both workflow files. The Go matrix contains `linux/amd64`, `linux/arm64`, `windows/amd64`, and `windows/arm64`. The documentation workflow runs on Windows and Linux.
- `npm ci`, gofmt, `go mod verify`, `go vet ./...`, Staticcheck v0.8.2, govulncheck v1.10.0, `go test ./... -count=1`, and `git diff --check` passed locally on Windows/amd64.
- The CLI smoke checks and regular Go suite passed under Linux arm64 emulation. See [plan 0003](0003-mvp-cli-compatibility.md#emulated-linux-arm64) for the race detector limitation and remaining native checks.

## Risks and open decisions

- The revised target list removes macOS from 0.1.0. Native Linux and Windows evidence, including arm64, is still required for release.
- Establish minimum supported OS versions from actual results. A configured job or an emulated container run has a separate evidence scope.
