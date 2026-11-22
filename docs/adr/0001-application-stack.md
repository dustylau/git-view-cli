# ADR 0001: application language and terminal stack

Status: Accepted

Date: 2026-11-22

## Context

We need a portable Git terminal application with CLI, Watch, and Interactive modes. We will delegate Git semantics and optional editor/diff/merge workflows to established tools. The project requires executable BDD, TDD, and strict quality gates.

## Candidates

- Go: Cobra, Bubble Tea, Bubbles/Lip Gloss, fsnotify, os/exec, and Godog.
- Rust: clap, Ratatui/Crossterm, notify, std::process, and cucumber-rs.

The [comparison](../STACK_COMPARISON.md) records equivalent components, advantages, limitations, and primary references. Both are viable. Go's simpler initial integration and Rust's ownership/resource control are tradeoffs, not proof of a better application.

## Decision

Use Go. On 2026-11-22 the maintainer explicitly chose Go and requested implementation, superseding the planned Rust/Go experiment. Both candidates remain technically viable; this is a maintainer preference for delivery and maintainability, not a measured performance or portability result. The experiment is not an acceptance prerequisite after this decision.

Start with the standard library for the CLI, Git subprocesses, parsing, and rendering. Use Godog for application BDD, alongside the existing documentation suite. Use gofmt, go vet, and Staticcheck. Pin the installed Go 1.27.4 toolchain and exact test/tool versions. Defer Cobra, terminal frameworks, and watchers until their slices justify dependencies.

## Decisions that apply to both

- Use installed Git CLI as the source of truth; do not introduce a second Git engine.
- Prefer existing widgets, Git patch output, and optional external diff/merge tools.
- Keep domain/use cases separate from mode-specific presentation.
- Use explicit tool launches and index mutations, with safe terminal handoff.
- Test natively on Windows, macOS, and Linux.

## Consequences

Adopt `cmd/git-view/` and `internal/` packages; remove the conceptual `src/` placeholders. Commit `go.mod` and `go.sum`. Node/npm remains documentation tooling only. The first slice is read-only, plain CLI output; terminal/process-tree lifecycle for interactive tools and watchers remains unproved and must be validated when introduced.

## Evidence and revisit criteria

The [CLI-001 plan](../plans/0002-go-mvp-status.md) records installed-toolchain evidence, dependency review, tests, and limits. No Rust prototype, comparative benchmark, native macOS/Linux run, or terminal smoke test establishes this decision. Revisit if native compatibility, required terminal lifecycle, or measured resource budgets cannot be met with maintained Go components. Release gates still require actual native evidence.
