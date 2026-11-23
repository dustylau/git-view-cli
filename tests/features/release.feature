Feature: Verify an MVP release candidate

  Scenario: Check a versioned candidate for all supported targets
    When I select the release candidate
    Then the candidate archives include the Go notices and cover exactly:
      | os      | arch  |
      | linux   | amd64 |
      | linux   | arm64 |
      | windows | amd64 |
      | windows | arm64 |
    And the checksum manifest verifies every archive
    And the host executable reports the requested version
    And the host executable inspects a mixed-change worktree without changing it
