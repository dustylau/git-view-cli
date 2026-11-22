# ADR 0002: Plain output for MVP 0.1.0

Status: Accepted for CLI-002

## Context

The CLI serves terminal users and supports redirected output. Git can report path bytes that are invalid UTF-8 on some filesystems. The current renderer escapes those bytes in plain text. A JSON API would also need conventions for byte preservation, schema evolution, and representing unknown upstream counts.

## Decision

Ship plain output for `status` and `changes` in 0.1.0. Keep the Git snapshot separate from rendering so a future JSON command can define its own fields and versioning. Revisit when an automation use case can drive the format and round-trip tests for paths.

## Consequences

Redirected output follows the [CLI display contract](../CLI.md). Consumers that already need Git's machine protocol can use Git porcelain directly. A future JSON format will have its own published schema and compatibility tests.
