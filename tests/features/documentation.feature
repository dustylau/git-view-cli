Feature: A reviewable documentation foundation
  As a contributor
  I need discoverable requirements, workflow, and stack options
  So that implementation starts from an agreed and testable scope

  Scenario: Discover product and command contracts
    Given I read the project README
    Then the README describes "CLI", "Watch", and "Interactive" modes
    And the README distinguishes the implemented CLI from planned modes
    And these documents are available:
      | file                     | topic                        |
      | docs/SPECIFICATION.md    | FR-001                       |
      | docs/CLI.md              | Exit codes                   |
      | docs/ARCHITECTURE.md     | Git adapter                  |
      | docs/STRUCTURE.md        | cmd/git-view                 |
      | docs/ROADMAP.md          | Acceptance gate              |

  Scenario: Discover the accepted Go decision and MVP scope
    Given I read the project README
    Then these documents are available:
      | file                                  | topic                    |
      | docs/STACK_COMPARISON.md               | Equivalent components    |
      | docs/STACK_COMPARISON.md               | Pros and cons            |
      | docs/STACK_COMPARISON.md               | Decision experiment      |
      | docs/adr/0001-application-stack.md      | Status: Accepted         |
      | docs/plans/0002-go-mvp-status.md        | Validation evidence      |
      | docs/ROADMAP.md                        | MVP delivery slices      |
    And the stack comparison records the maintainer's Go decision

  Scenario: Follow a documented quality workflow
    Given I read the project README
    Then these documents are available:
      | file                                       | topic                      |
      | CONTRIBUTING.md                            | Mandatory workflow         |
      | docs/TESTING.md                            | Red/green                  |
      | SECURITY.md                                | Trust boundaries           |
      | docs/plans/0001-documentation-foundation.md | Checklist                  |
      | docs/plans/0003-mvp-cli-compatibility.md     | Linux race detector result |
    And the contribution guide defines eleven ordered workflow steps

  Scenario: Follow local documentation links
    Given I read the project README
    Then every local Markdown file link in project documents resolves
