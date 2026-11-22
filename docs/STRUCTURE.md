# Repository structure

## Language-neutral foundation

```text
git-view-cli/
  README.md
  CONTRIBUTING.md
  SECURITY.md
  docs/
    SPECIFICATION.md
    CLI.md
    ARCHITECTURE.md
    STACK_COMPARISON.md
    STRUCTURE.md
    ROADMAP.md
    TESTING.md
    adr/
    plans/
    templates/
  src/
    README.md
    core/
    git/
    monitor/
    terminal/
    tools/
  tests/
    features/
    steps/
    fixtures/
  .github/
    workflows/
  package.json
  package-lock.json
```

The source and fixture directories contain `.gitkeep` placeholders only. Their boundaries express responsibilities; they do not impose an application package structure before choosing a language. Node dependencies are documentation development tools only. No application manifest or executable exists.

| Area              | Responsibility                                                           |
| ----------------- | ------------------------------------------------------------------------ |
| `src/core/`       | Domain models, use cases, shared session state.                          |
| `src/git/`        | Git command adapter and machine-output parsers.                          |
| `src/monitor/`    | Invalidation, debounce, reconciliation, polling.                         |
| `src/terminal/`   | CLI, Watch, Interactive, terminal lifecycle and presentation.            |
| `src/tools/`      | Explicit editor/pager/diff/merge integration.                            |
| `tests/features/` | Executable Gherkin acceptance specifications.                            |
| `tests/steps/`    | Documentation feature bindings; product bindings follow stack selection. |
| `tests/fixtures/` | Future parser samples and small, synthetic repository fixtures.          |
| `docs/templates/` | Scope plan template for the mandatory workflow.                          |

## After stack selection

If Rust is selected, add one Cargo binary crate initially: `Cargo.toml`, committed `Cargo.lock`, pinned `rust-toolchain.toml`, and modules under `src/`. Add integration tests under `tests/`; do not introduce a multi-crate workspace until boundaries justify it.

If Go is selected, migrate conceptual source directories to `cmd/git-view/` and `internal/{core,git,monitor,terminal,tools}/`, add `go.mod`/`go.sum`, and place unit tests alongside packages. Do not force Go into the provisional `src/` tree.

Either migration requires an updated plan, executable failing BDD scenario, tests, CI, and this directory guide. Remove obsolete placeholders rather than retaining duplicate layouts.
