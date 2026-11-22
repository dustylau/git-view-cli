# Plan 0001: documentation and stack comparison

## Scope

Deliver DOC-001 in the [specification](../SPECIFICATION.md). The earlier attempt initialized an empty Git branch but wrote no files. This increment resumes on that clean branch. Application implementation and final stack selection are outside this increment.

## Strategy

Write the README first. Keep the source structure language-neutral. Research Rust equivalents for each proposed Go component using primary documentation. Compare project-specific tradeoffs without claiming unmeasured performance advantages. Use Cucumber.js for documentation BDD and Node only as development tooling so neither application candidate is selected by the test harness.

## Checklist

- [x] Verify a clean starting tree on `docs/project-foundation` and inspect available tools.
- [x] Write README first; record DOC-001, implementation strategy, and checklist.
- [x] Define and run a failing BDD documentation feature before creating supporting docs.
- [x] Add product, CLI, architecture, and structure docs; rerun BDD.
- [x] Add researched stack comparison and proposed ADR; rerun BDD.
- [x] Add workflow, test strategy, security, roadmap, and source placeholders; rerun BDD to green.
- [x] Run formatting, linting/static checks, all current tests, and whitespace checks.
- [x] Verify documentation matches implementation and record evidence/limitations.
- [x] Create a meaningful commit and verify the branch is clean (Git history and final status provide closeout evidence).

## Evidence

Executed locally on Windows with Node 22.16.0, npm 11.6.4, and PowerShell 7. The final Git commit and clean status establish the last checklist item's completion; this document does not embed its own commit hash.

| Check                                | Actual outcome                                                                                                                                                                                                                     |
| ------------------------------------ | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Starting Git status                  | Clean unborn `docs/project-foundation`; no pre-existing project files.                                                                                                                                                             |
| BDD red: `npm test`                  | Corrected an initial step-binding arity error, then reran: four scenarios failed for missing supporting docs/links, with six steps passed. Harness failure was not treated as acceptance red.                                      |
| First document group: `npm test`     | Product/command scenario green; one scenario passed, three still failed.                                                                                                                                                           |
| Stack comparison and ADR: `npm test` | Two scenarios passed, two still failed for remaining docs/links.                                                                                                                                                                   |
| Remaining docs: `npm test`           | Four scenarios and twelve steps passed.                                                                                                                                                                                            |
| `npm run format`                     | Applied consistent Markdown, JSON, YAML, and JavaScript formatting.                                                                                                                                                                |
| `npm ci`                             | Reinstalled the exact lockfile successfully.                                                                                                                                                                                       |
| `npm run check`                      | Formatting passed, fifteen Markdown files linted with zero issues, step JavaScript syntax valid, all four BDD scenarios/twelve steps passed.                                                                                       |
| `git diff --check`                   | Passed; staged whitespace verification is repeated at closeout.                                                                                                                                                                    |
| Dependency review                    | Initial audit reported four low-severity package findings from one transitive KaTeX advisory. Scoped override to patched KaTeX 0.18.2 applied; renderer API smoke check passed; final install/audit reported zero vulnerabilities. |

## Review and limitations

Reviewed requirements against README, planned commands, architecture, roadmap, source placeholders, dependency manifests, and executable checks. The source folders contain no implementation. Stack choice remains open and ADR 0001 remains proposed. No application manifests, install claims, prototype benchmarks, or application test results are presented as complete.

The dependency override crosses the math extension's original version range. Its used renderer API and current lint suite passed; revisit the override when the upstream range incorporates the patch. See [testing guidance](../TESTING.md).

No Go/Rust compiler is installed locally. Native macOS/Linux tests, ARM64 behavior, and CI jobs have not run. The documentation CI matrix is configured but there is no remote configured for execution. Node 22 emits an experimental glob warning during Cucumber execution; scenarios still run and pass. No production runtime dependencies exist.
