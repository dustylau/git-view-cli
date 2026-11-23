# Repository structure

## Current layout

Go is selected in [ADR 0001](adr/0001-application-stack.md). CLI-001 replaces the language-neutral `src/` placeholders with Go packages:

```text
git-view-cli/
  cmd/git-view/main.go
  internal/
    cli/                 # arguments, plain rendering, exit codes; unit tests
    git/                 # snapshot model, porcelain parser, Git boundary; tests
  tests/
    acceptance/
      cli_test.go
      release_test.go
      terminal_smoke.py
      measure_cli.py
      testdata/slowgit/main.go
    container/
      container_workflow.py
      seed.sh
      check.sh
      test_workflow.py
    features/
      cli.feature
      container.feature
      release.feature
      documentation.feature
    steps/documentation.mjs
  docs/
    SPECIFICATION.md
    CLI.md
    INSTALL.md
    ARCHITECTURE.md
    STACK_COMPARISON.md
    STRUCTURE.md
    ROADMAP.md
    TESTING.md
    adr/
    plans/
    templates/
  .github/workflows/
    documentation.yml
    go.yml
  .gitea/workflows/local-ci.yml
  tools/ci/              # Docker job image, local provisioning and bootstrap
  tools/release/
    package_release.py
    test_package_release.py
  go.mod
  go.sum
  Dockerfile
  compose.yaml
  .dockerignore
  package.json
  package-lock.json
```

| Area                       | Responsibility                                                                                  |
| -------------------------- | ----------------------------------------------------------------------------------------------- |
| `cmd/git-view/`            | Executable entry point, interrupt context, and process exit.                                    |
| `internal/cli/`            | Standard-library argument parser, injected snapshot reader, safe plain output, public errors.   |
| `internal/git/`            | Repository identity/status model, direct bounded Git calls, NUL-delimited porcelain-v2 parsing. |
| `tests/acceptance/`        | Godog bindings, temporary Git fixtures, terminal smoke test, and performance measurement.       |
| `tests/features/`          | Executable Gherkin specifications for the CLI and project documentation.                        |
| `tests/container/`         | Docker build runner, mount lifecycle tests, and scripts for the mounted Git checks.             |
| `tools/release/`           | Build, package, and checksum the four MVP executables.                                          |
| `tests/steps/`             | Node-based documentation bindings only.                                                         |
| `tests/fixtures/`          | Reserved for future captured samples; current repository fixtures use temporary paths.          |
| `docs/templates/`          | Implementation-plan template for the mandatory workflow.                                        |
| `.github/workflows/go.yml` | Configured native Go test/static checks; not proof that CI has run.                             |

No monitor, terminal widget, or tool-launch packages exist yet. Add `internal/core`, `monitor`, `terminal`, or `tools` only when an implementing slice needs them. Keep tests beside Go packages and avoid speculative empty package hierarchies.

`go.mod` uses the provisional local module name `git-view-cli` because no hosting identity is configured; finalize it before release installation instructions use a remote module path. `go.sum` locks the test dependency graph. Production builds use only standard-library packages. `package.json` and its lockfile are documentation development tooling. Generated `bin/`, `dist/`, coverage, and report files are ignored.
