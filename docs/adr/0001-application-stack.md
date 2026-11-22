# ADR 0001: application language and terminal stack

Status: Proposed

Date: 2026-11-22

## Context

We need a portable Git terminal application with CLI, Watch, and Interactive modes. We will delegate Git semantics and optional editor/diff/merge workflows to established tools. The project requires executable BDD, TDD, and strict quality gates.

## Candidates

- Go: Cobra, Bubble Tea, Bubbles/Lip Gloss, fsnotify, os/exec, and Godog.
- Rust: clap, Ratatui/Crossterm, notify, std::process, and cucumber-rs.

The [comparison](../STACK_COMPARISON.md) records equivalent components, advantages, limitations, and primary references. Both are viable. Go's simpler initial integration and Rust's ownership/resource control are tradeoffs, not proof of a better application.

## Proposed decision process

Keep the stack decision open. Execute the same bounded decision experiment in each candidate during M1. Establish correctness, OS support, process/terminal cleanup, build reproducibility, maintenance effort, and resource budgets before choosing. Accept this ADR only after recording actual evidence and the chosen stack.

## Decisions that apply to both

- Use installed Git CLI as the source of truth; do not introduce a second Git engine.
- Prefer existing widgets, Git patch output, and optional external diff/merge tools.
- Keep domain/use cases separate from mode-specific presentation.
- Use explicit tool launches and index mutations, with safe terminal handoff.
- Test natively on Windows, macOS, and Linux.

## Consequences

No Cargo or Go application manifest is added now. Source directories are conceptual placeholders. Node/npm tooling runs documentation acceptance and lint/format checks only. M1 will select/pin a compiler and dependencies and adapt the directory structure to the chosen language's conventions.

## Evidence required for acceptance

Attach the M1 plan/results, platform test evidence, dependency/license review, measured comparison, and supported target/toolchain matrix. Record why the alternative was not selected and what evidence would justify revisiting the choice.
