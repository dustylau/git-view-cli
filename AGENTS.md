# Repository working instructions

Read [CONTRIBUTING.md](CONTRIBUTING.md) before changing this repository. Its eleven-step workflow is mandatory: clean named branch, specification, detailed plan/checklist, demonstrated BDD red, iterative TDD red/green, formatting/static analysis, all tests, documentation updates/alignment, meaningful commit, and all scoped changes committed to the feature branch.

Read the [current specification](docs/SPECIFICATION.md), [roadmap](docs/ROADMAP.md), and relevant plan before implementing. Go is selected by the maintainer in accepted ADR 0001; the comparative experiment is superseded. Use Git CLI and established terminal/diff/merge utilities rather than reimplementing them. Do not introduce application behavior in a documentation-only task.

Current verification: `npm ci`, `npm run check`, gofmt verification, `go vet ./...`, `go test ./...`, pinned Staticcheck/govulncheck commands in [TESTING.md](docs/TESTING.md), and `git diff --check`. Node is documentation tooling only. Report actual checks and limitations; do not claim application or platform coverage that has not run. Preserve unrelated work and leave the feature branch clean after committing authorized changes.
