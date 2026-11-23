# git-view-cli

A portable terminal companion for inspecting and managing a local Git repository.

**Implemented: read-only `status` and `changes` CLI in Go (0.1.0-dev). No release is published.**
The executable name `git-view` is provisional. Watch and Interactive are planned.

## Run the first slice

Use Go 1.27.4 and Git 2.43.0 or newer. Run these commands from the checkout.

```sh
go run ./cmd/git-view
go run ./cmd/git-view status
go run ./cmd/git-view --repo /path/to/repository changes
go run ./cmd/git-view --help
```

Build with `go build -o bin/git-view ./cmd/git-view` on Linux, or `go build -o bin/git-view.exe ./cmd/git-view` on Windows. Run the built executable directly to observe its documented exit codes. `go run` wraps nonzero program exits. The binary requires Git on PATH for data commands. Node is used for documentation checks.

Status discovers the worktree from the current or selected directory and reports branch/upstream and separate staged, unstaged, untracked, and conflict counts. Changes shows index/worktree codes and quoted root-relative paths, including renames. Reads are bounded, suppress optional index locks, and do not fetch or stage. See the [CLI contract](docs/CLI.md) for output, limits, escaping, and exit codes.

## Build and test with Docker

With Docker running Linux containers and Python 3.9 or newer, run these commands from the checkout. The first command builds the image and runs the CLI and Go tests with a fresh seeded Git volume. The second checks an existing repository through a read-only bind mount.

```powershell
py -3 tests/container/container_workflow.py
py -3 tests/container/container_workflow.py --repo "D:\path\to\checkout"
```

Use `python3` in place of `py -3` on Linux or macOS. See [container test details](docs/TESTING.md#container-build-and-test).

## Planned product capabilities

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

Bare `git-view` shows status once. The table's Watch and Interactive examples remain future commands. Planned interactive sessions will switch views, pause updates, enter a Watch presentation, and execute one-shot actions while preserving repository context.

## Stack decision

**Go is selected by the maintainer.** The [accepted ADR](docs/adr/0001-application-stack.md) supersedes the comparative experiment; the [Rust versus Go comparison](docs/STACK_COMPARISON.md) remains design context. The current CLI uses the Go standard library; Godog is a test dependency. Terminal and watcher libraries are deferred.

Git supplies repository semantics. Existing diff, pager, editor, and merge tools take priority in later slices. The MVP targets Windows and Linux on amd64 and arm64. macOS support follows in a later release. Local tests cover Windows 11 and Ubuntu 24.04 under WSL2 on amd64. See the compatibility plan for the native runner gates and exact results.

## Documentation

- [Specification](docs/SPECIFICATION.md): numbered requirements and acceptance criteria.
- [CLI contract](docs/CLI.md): planned commands, modes, output, and errors.
- [Architecture](docs/ARCHITECTURE.md): components and Git/tool boundaries.
- [Repository structure](docs/STRUCTURE.md): Go packages and test directories.
- [Roadmap](docs/ROADMAP.md): incremental delivery gates.
- [Current plan](docs/plans/0003-mvp-cli-compatibility.md): compatibility scope, checklist, and validation evidence.
- [Contributing](CONTRIBUTING.md): mandatory development workflow.
- [Testing](docs/TESTING.md): BDD/TDD, checks, and platform verification.
- [Security](SECURITY.md): local repository and process trust boundaries.

## Verify the project

Install Go 1.27.4, Git, and Node.js 22.16 or later in the Node 22 line, or Node 24 LTS, plus npm. Node is used only for documentation tooling.

```sh
npm ci
npm run check
go vet ./...
go test ./...
go run honnef.co/go/tools/cmd/staticcheck@v0.8.2 ./...
go run golang.org/x/vuln/cmd/govulncheck@v1.10.0 ./...
git diff --check
```

`gofmt -l cmd internal tests/acceptance` must produce no output. Go tests include unit and process tests and executable Godog scenarios against temporary repositories. Documentation checks cover acceptance, local links, formatting, and linting. CI configures native Windows and Linux jobs for both MVP architectures. See [testing guidance](docs/TESTING.md) for the available run evidence.

## Development rule

Every change starts on a clean, named branch with an updated specification and checklist. Define a failing BDD scenario, implement through TDD red/green steps, run all applicable checks, align documentation, and commit all changes to that branch. See the [full workflow](CONTRIBUTING.md).

## License

No license has been selected. Choose one before public distribution.
