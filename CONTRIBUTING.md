# Contributing

## Mandatory workflow

Apply every step to features, fixes, refactors, configuration, and documentation. Scale the evidence to the change; do not invent application tests for documentation-only behavior. Authorization to perform a task includes routine local verification and commits to its feature branch.

1. **Start clean on a named branch.** Check `git status --short --branch`; preserve existing user work. Create `feat/<scope>`, `fix/<scope>`, `docs/<scope>`, or `chore/<scope>` from the intended base. Never silently discard, stash, or commit unrelated work. For the initial repository, use the clean unborn `docs/project-foundation` branch.
2. **Update specifications first.** Describe the problem, observable behavior, scope, acceptance criteria, and affected requirement IDs in `docs/SPECIFICATION.md` and the relevant contract. Record architectural decisions in an ADR.
3. **Write the implementation plan.** Add a numbered plan under `docs/plans/` using the [template](docs/templates/implementation-plan.md). Include ordered steps, risks, test strategy, and a checklist.
4. **Demonstrate BDD red.** Write an executable `.feature` with real step bindings. Run it and confirm failure for the missing behavior, not a broken harness or unrelated environment problem. Record the command and failure reason before implementing behavior.
5. **Implement through TDD.** For each logical behavior, write the smallest useful failing test, implement enough to pass, then refactor while green. Repeat until the acceptance feature passes. For docs-only work, the documentation acceptance scenarios supply the red/green loop; do not add redundant unit tests for prose.
6. **Run quality tools.** Format, lint, and run language-appropriate static analysis (Clippy if Rust; vet and the selected linter if Go). The current documentation gate is `npm run format:check` and `npm run lint`. “Copped” means the equivalent static-quality gate for the selected language, not a requirement to introduce Ruby tooling.
7. **Run all tests.** Run `npm test` now, and all unit, integration, BDD, and applicable terminal/platform tests once added. Record limitations explicitly; a skipped platform is not a passed platform.
8. **Update user and project docs.** Revise README, commands, setup, architecture, examples, roadmap, and test instructions affected by the change. State what exists versus what is planned.
9. **Check alignment.** Compare documentation with actual files, commands, dependencies, test results, and implementation. Check local links and whitespace. Review the complete diff for unintended changes and secrets.
10. **Prepare a meaningful commit message.** Describe scope and outcome, such as `docs: establish project contracts and compare Rust with Go`. Add a body when it clarifies the reason or validation.
11. **Commit all scoped changes to the feature branch.** Stage deliberately, inspect `git diff --cached`, commit, and verify `git status --short --branch` is clean. Leave merging, pushing, or publishing to separately authorized work. Do not claim completion while scoped changes remain uncommitted.

## Current checks

```sh
npm ci
npm run format
npm run check
gofmt -l cmd internal tests/acceptance
go vet ./...
go test ./...
go run honnef.co/go/tools/cmd/staticcheck@v0.8.2 ./...
go run golang.org/x/vuln/cmd/govulncheck@v1.10.0 ./...
git diff --check
```

Node/npm is documentation development tooling only. Go 1.27.4 is the selected application toolchain. `gofmt -l` must produce no output (it does not itself fail for unformatted files); use `gofmt -w` to fix formatting. Follow [testing guidance](docs/TESTING.md) for executable BDD, race tests, dependencies, and platform limits.

## Definition of done

- Scope, acceptance criteria, plan, and implementation agree.
- Red/green evidence exists for changed observable behavior.
- Applicable quality gates and the full current suite pass.
- User documentation and architecture decisions are current.
- Known limitations are recorded accurately.
- Changes are committed on the named feature branch and its worktree is clean.

## Review and dependency discipline

Keep increments small and reviewable. Prefer standard libraries and established components. Review maintenance, license, native dependencies, compatibility, and advisories before introducing a dependency. Commit application lockfiles as appropriate for the selected language and pin CI tools. Do not change the Git engine, add a custom diff algorithm, or replace external tools without a specification and an ADR explaining the unmet requirement.
