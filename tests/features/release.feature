Feature: Package a local MVP release candidate

  Scenario: Build a versioned candidate for all supported targets
    When I package version "0.1.0-rc.1"
    Then the candidate archives cover exactly:
      | os      | arch  |
      | linux   | amd64 |
      | linux   | arm64 |
      | windows | amd64 |
      | windows | arm64 |
    And the checksum manifest verifies every archive
    And the host executable reports version "0.1.0-rc.1"
