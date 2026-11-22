# Testing and verification

## Current executable suite

The repository currently contains documentation, conceptual source placeholders, and Node-based documentation development tools. No Git application exists. The current suite checks discoverability of the product/mode contracts, open Rust/Go decision, required workflow docs, and local Markdown file links.

```sh
npm ci
npm test
npm run format:check
npm run lint
npm run check
git diff --check
```

`npm run check` combines format, Markdown lint, JavaScript syntax checking for the step bindings, and BDD. `npm run format` applies formatting. Versions are exact in `package.json` and resolved in `package-lock.json`. The Cucumber runner uses real Gherkin and step bindings, not a bespoke feature-file parser.

The local-link scenario checks inline relative file links in project Markdown, skipping external URLs and fenced examples. It verifies file existence, not remote availability, fragment anchors, reference-style links, or semantic correctness. Human review remains necessary. It also checks newly created docs before the initial commit.

## Red/green evidence

Before implementation, run the new feature and verify it fails for the intended missing behavior. Fix any harness errors before recording a valid red. For this increment, missing supporting documentation supplies the failure. Add documents in logical groups and rerun until all scenarios pass. Record actual results in the [current plan](plans/0001-documentation-foundation.md).

Future product features must assert observable application results against temporary repositories. A `.feature` file without executable bindings is not acceptance coverage. Undefined, pending, or skipped scenarios cannot establish acceptance. Do not leave future-product prose scenarios in the active passing suite and call the product implemented.

## Application test layers after selection

| Layer                | Purpose and examples                                                                                                                                      |
| -------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Unit                 | Status parser records, rename/NUL handling, command argument construction, state transitions, refresh coalescing, cancellation.                           |
| Git integration      | Real temporary repositories: staged/unstaged combinations, unborn/detached HEAD, worktrees, ignored files, submodules, conflicts, binary/large files.     |
| CLI acceptance       | Exit codes, stdout/stderr separation, no-argument status, help outside a repo, path and revision validation, plain output.                                |
| Watch acceptance     | External changes, atomic saves, newly created directories, metadata events, overflow, polling, stale snapshots, Ctrl+C.                                   |
| Interactive/terminal | Navigation, resize, Windows key event behavior, tool handoff, cancellation, mode transitions, terminal restoration after failures.                        |
| Conflict acceptance  | Content/add/delete conflicts, absent stages, tool failure/success, actual unmerged index state, explicit staging and unstaging without unrelated changes. |

Use synthetic data and per-test temporary directories. Set test-only Git author identity and configuration in those repositories; never depend on a contributor's global Git config, network, pager, editor, or credentials. Clean up child processes before removing fixtures. Include spaces, Unicode, leading dashes, and control characters where supported. Assert read-only commands leave worktree/index/config unchanged.

Inject clock and event sources for deterministic debounce/reconciliation unit tests. Use bounded eventual assertions for real OS events rather than fixed sleeps. Keep expensive terminal smoke tests separate from deterministic rendering/state tests, and run both when terminal code changes.

## Candidate quality gates

These commands are planned and will become executable only after selecting a stack and adding manifests/toolchains:

| Rust                                                | Go                                                                            |
| --------------------------------------------------- | ----------------------------------------------------------------------------- |
| `cargo fmt --check`                                 | `gofmt` verification                                                          |
| `cargo clippy --all-targets -- -D warnings`         | `go vet ./...` and the selected pinned linter                                 |
| `cargo test --locked` plus cucumber-rs target       | `go test ./...` plus Godog integration                                        |
| Selected dependency/license/advisory tooling        | Selected dependency/license/advisory tooling                                  |
| Explicit child-process and terminal lifecycle tests | Race testing where supported, plus child-process and terminal lifecycle tests |

Pin the Rust BDD target/harness and Go BDD invocation in the application bootstrap. Include all supported dependency features/build configurations in the matrix rather than assuming the defaults establish every build.

## CI and platform claims

The [documentation workflow](../.github/workflows/documentation.yml) is configured for Windows, macOS, and Linux with read-only repository permissions, pinned action revisions, `npm ci`, and the current checks. It is configuration until executed by CI; local Windows results do not prove Linux/macOS results. Product CI and release target checks will be added after stack selection. Do not claim ARM64 runtime coverage from amd64 cross-builds.

## Dependency review

Use `npm audit` when changing documentation tooling and record outstanding findings rather than applying breaking `--force` updates blindly. Production dependencies do not exist yet. Dependency advisories, their reachability, and any overrides must be recorded in the current plan and revisited with dependency updates.

The initial dependency review found a low-severity [KaTeX advisory](https://github.com/advisories/GHSA-238p-pmpm-9mq7) through Markdown lint's math extension. A scoped npm override pins that extension's KaTeX dependency to patched `0.18.2`. This crosses the extension's original dependency range; the current lint suite and a renderer API smoke check passed. Revisit/remove the override when the upstream range includes the fix. This is development tooling, not an application runtime dependency.
