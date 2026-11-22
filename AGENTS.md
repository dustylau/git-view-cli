# Repository working instructions

Read [CONTRIBUTING.md](CONTRIBUTING.md) before changing this repository. Its eleven-step workflow is mandatory: clean named branch, specification, detailed plan/checklist, demonstrated BDD red, iterative TDD red/green, formatting/static analysis, all tests, documentation updates/alignment, meaningful commit, and all scoped changes committed to the feature branch.

Read the [current specification](docs/SPECIFICATION.md), [roadmap](docs/ROADMAP.md), and relevant plan before implementing. Keep Rust/Go selection open until the decision experiment and accepted ADR establish the stack. Use Git CLI and established terminal/diff/merge utilities rather than reimplementing them. Do not introduce application behavior in a documentation-only task.

Current verification: `npm ci`, `npm run check`, and `git diff --check`. Node is documentation tooling only. Report actual checks and limitations; do not claim application or platform coverage that has not run. Preserve unrelated work and leave the feature branch clean after committing authorized changes.
