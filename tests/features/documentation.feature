Feature: A reviewable documentation foundation
  As a contributor
  I need discoverable requirements, workflow, and stack options
  So that implementation starts from an agreed and testable scope

  Scenario: Discover product and command contracts
    Given I read the project README
    Then the README describes "CLI", "Watch", and "Interactive" modes
    And the README states that the application is not implemented
    And these documents are available:
      | file                     | topic                        |
      | docs/SPECIFICATION.md    | FR-001                       |
      | docs/CLI.md              | Exit codes                   |
      | docs/ARCHITECTURE.md     | Git adapter                  |
      | docs/STRUCTURE.md        | Language-neutral             |
      | docs/ROADMAP.md          | Acceptance gate              |

  Scenario: Compare candidates without prematurely selecting a stack
    Given I read the project README
    Then these documents are available:
      | file                                  | topic                    |
      | docs/STACK_COMPARISON.md               | Equivalent components    |
      | docs/STACK_COMPARISON.md               | Pros and cons            |
      | docs/STACK_COMPARISON.md               | Decision experiment      |
      | docs/adr/0001-application-stack.md      | Status: Proposed         |
    And the stack comparison keeps the application decision open

  Scenario: Follow a documented quality workflow
    Given I read the project README
    Then these documents are available:
      | file                                          | topic              |
      | CONTRIBUTING.md                               | Mandatory workflow |
      | docs/TESTING.md                               | Red/green          |
      | SECURITY.md                                   | Trust boundaries   |
      | docs/plans/0001-documentation-foundation.md     | Checklist          |
    And the contribution guide defines eleven ordered workflow steps

  Scenario: Follow local documentation links
    Given I read the project README
    Then every local Markdown file link in project documents resolves
