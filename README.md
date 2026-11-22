# git-view-cli

A portable terminal companion for inspecting and managing a local Git repository.

**Status: documentation foundation. No application, installable binary, or release exists yet.**
The executable name `git-view` is provisional. Application commands below describe planned behavior.

## Capabilities

- Monitor a local repository and display branch, upstream, and working state.
- Browse repository files and folders and open selected files in an editor.
- List new, updated, deleted, renamed, staged, and conflicted files.
- Browse commits, inspect file history, and view diffs.
- Assist merge conflict resolution through existing Git tools and explicit staging.

## Three modes

| Mode          | Behavior                                                   | Planned example                         |
| ------------- | ---------------------------------------------------------- | --------------------------------------- |
| CLI (default) | Perform one action, display the result, and exit.          | `git-view status`                       |
| Watch         | Continuously refresh views selected through arguments.     | `git-view watch --views status,changes` |
| Interactive   | Maintain a session with access to all functions and views. | `git-view interactive`                  |

Bare `git-view` will show status once. Interactive sessions will switch views, pause updates, enter a Watch presentation, and execute one-shot actions while preserving repository context.

## Stack decision

**Rust and Go are both candidates; no application stack is selected.** Read the [Rust versus Go comparison](docs/STACK_COMPARISON.md) for equivalent libraries, pros and cons, and a decision experiment. The [proposed ADR](docs/adr/0001-application-stack.md) records the decision gate.

Both candidates will delegate repository operations to the installed Git CLI. Existing diff, pager, editor, and merge tools take priority over custom implementations. Runtime targets are Windows, macOS, and Linux; the initial architecture targets are amd64 and arm64. Exact minimum versions must be established by testing before release.

## Documentation

- [Specification](docs/SPECIFICATION.md): numbered requirements and acceptance criteria.
- [CLI contract](docs/CLI.md): planned commands, modes, output, and errors.
- [Architecture](docs/ARCHITECTURE.md): components and Git/tool boundaries.
- [Repository structure](docs/STRUCTURE.md): current directories and future language layouts.
- [Roadmap](docs/ROADMAP.md): incremental delivery gates.
- [Current plan](docs/plans/0001-documentation-foundation.md): scope, checklist, and validation evidence.
- [Contributing](CONTRIBUTING.md): mandatory development workflow.
- [Testing](docs/TESTING.md): BDD/TDD, checks, and platform verification.
- [Security](SECURITY.md): local repository and process trust boundaries.

## Verify the foundation

Install Node.js 22.16 or later in the Node 22 line, or Node 24 LTS, plus npm and Git. Node is used only for documentation tooling; it does not select the application runtime.

```sh
npm ci
npm run check
git diff --check
```

Current checks cover documentation acceptance, local Markdown file links, formatting, and linting. There are no application tests yet. CI is configured to run these checks on Windows, macOS, and Linux; local execution alone does not demonstrate that CI has passed.

## Development rule

Every change starts on a clean, named branch with an updated specification and checklist. Define a failing BDD scenario, implement through TDD red/green steps, run all applicable checks, align documentation, and commit all changes to that branch. See the [full workflow](CONTRIBUTING.md).

## License

No license has been selected. Choose one before public distribution.
