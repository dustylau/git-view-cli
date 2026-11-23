# git-view-cli

Git View 1.0.0 brings repository status, file browsing, commit history, patches and conflict resolution into one terminal. The CLI works with installed Git 2.43.0 or newer. Download the ZIP for Windows or Linux on amd64 or arm64 from [GitHub Releases](https://github.com/dustylau/git-view-cli/releases) and follow the [installation guide](docs/INSTALL.md).

## Use Git View

Run these commands inside a worktree, or select one with `--repo <directory>`.

```sh
git-view status
git-view changes
git-view files -- src
git-view log --limit 10
git-view history -- README.md
git-view show HEAD
git-view diff --staged -- README.md
git-view watch --views status,changes
git-view interactive
```

Bare `git-view` prints status. Watch refreshes selected views every two seconds. In Interactive, type a command and press Return. For example, enter `files`, then `edit -- "name with spaces.txt"`. Enter `watch` for live updates, `back` to return to the active view, or `quit` to leave.

Use `conflicts` to see the available Git stages. Open an unmerged file with `resolve --editor -- FILE` or launch your configured Git mergetool with `resolve --tool -- FILE`. Git View reports the resulting index state. Stage or unstage individual files with confirmation. Use `stage --remove -- FILE` for an intentional deletion and `--yes` for scripted confirmation. The [CLI guide](docs/CLI.md) covers editor settings, quoted paths and exit codes.

## Build from source

Use Go 1.27.4. Build with `go build -o bin/git-view ./cmd/git-view` on Linux or `go build -o bin/git-view.exe ./cmd/git-view` on Windows. Running `go run ./cmd/git-view --help` prints the command list. Node is used for documentation checks.

## Build and test with Docker

With Docker running Linux containers and Python 3.9 or newer, run these commands from the checkout. The first command builds the image and runs the CLI and Go tests with a fresh seeded Git volume. The second checks an existing repository through a read-only bind mount.

```powershell
py -3 tests/container/container_workflow.py
py -3 tests/container/container_workflow.py --repo "D:\path\to\checkout"
```

Use `python3` in place of `py -3` on Linux or macOS. See [container test details](docs/TESTING.md#container-build-and-test).

## Run local CI and delivery

With Docker running Linux containers, seed the current committed branch into Gitea and wait for the tested ZIPs and checksums.

```powershell
.\tools\ci\bootstrap.ps1 --wait
```

On Linux or WSL, run `sh tools/ci/bootstrap.sh --wait`. The local server starts at `http://127.0.0.1:3080`. See [local CI setup and downloads](docs/LOCAL_CI.md).

## Package a local candidate

With Go 1.27.4 and Python 3.9 or newer, run the packager from the checkout.

```powershell
py -3 tools/release/package_release.py --version 1.0.0
```

On Linux, use `python3` in place of `py -3`. The command writes four ZIP archives and `SHA256SUMS` to `dist/git-view-1.0.0/`. Each archive includes installation instructions, the MIT license and the Go notices. Use `--out` to choose another parent directory. See [installation and upgrade instructions](docs/INSTALL.md) and the [candidate packaging plan](docs/plans/0006-mvp-candidate-packaging.md) for validation and release gates.

The [candidate verification guide](docs/TESTING.md#release-candidate-packaging) describes checking an existing set locally or running the release-candidate workflow on the four native runners.

## Implementation

Git supplies repository status, revision traversal, patches, index transactions and merge tools. The Go executable uses normal terminal line input. See the [architecture](docs/ARCHITECTURE.md) and the [v1 roadmap](docs/ROADMAP.md) for design decisions and later features.

## Documentation

- [Specification](docs/SPECIFICATION.md): numbered requirements and acceptance criteria.
- [CLI contract](docs/CLI.md): commands, modes, output, and errors.
- [Architecture](docs/ARCHITECTURE.md): components and Git/tool boundaries.
- [Repository structure](docs/STRUCTURE.md): Go packages and test directories.
- [Roadmap](docs/ROADMAP.md): incremental delivery gates.
- [Current plan](docs/plans/0011-v1-release.md): v1 scope, checklist, and validation evidence.
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

`gofmt -l cmd internal tests/acceptance` must produce no output. Go tests include unit and process tests and executable Godog scenarios against temporary repositories. Documentation checks cover acceptance, local links, formatting, and linting. CI configures native Windows and Linux jobs for both architectures. See [testing guidance](docs/TESTING.md) for the available run evidence.

## Development rule

Every change starts on a clean, named branch with an updated specification and checklist. Define a failing BDD scenario, implement through TDD red/green steps, run all applicable checks, align documentation, and commit all changes to that branch. See the [full workflow](CONTRIBUTING.md).

## License

Git View is distributed under the [MIT License](LICENSE).
