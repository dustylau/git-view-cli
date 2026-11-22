# Testing and verification

## Current executable suite

Use Go 1.27.4, installed Git, and Node 22.16+ in the Node 22 line or Node 24 LTS with npm. Node remains documentation tooling only; application BDD uses Godog v0.16.0. Current native evidence is Windows/amd64 with Git 2.49.0.windows.1. See [plan 0002](plans/0002-go-mvp-status.md) for actual red/green results and limits.

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

The [documentation workflow](../.github/workflows/documentation.yml) and [Go workflow](../.github/workflows/go.yml) configure native Windows/macOS/Linux jobs, read-only permissions, and pinned actions/tools. Go CI runs formatting, vet, Staticcheck, module verification, all tests, a Linux race run, and a vulnerability scan. No remote is configured locally, so these files are configuration rather than passing remote evidence. No macOS/Linux native or ARM64 runtime test has run in this session. Cross-builds only prove compilation.

## Dependency review

Production builds use only Go's standard library and installed Git. Godog v0.16.0 is a test dependency (MIT, no cgo); its transitive graph is recorded in `go.mod`/`go.sum`. Staticcheck v0.8.2 and govulncheck v1.10.0 run at pinned versions without becoming application dependencies. Run `go mod tidy`, `go mod verify`, and the pinned vulnerability command when changing the Go graph. Advisory results are time-specific, not permanent guarantees.

Use `npm audit` when changing documentation tooling. The foundation's scoped override pins Markdown lint's transitive KaTeX to `0.18.2` for the [KaTeX advisory](https://github.com/advisories/GHSA-238p-pmpm-9mq7). This crosses the math extension's original range; its initial renderer smoke test and lint suite passed. Revisit when upstream includes the fix. Node packages do not ship in the CLI.
