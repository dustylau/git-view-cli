Feature: Read-only local Git inspection
  As a developer
  I want a plain status and changes CLI
  So I can inspect my repository without changing it

  Scenario: Default status on an unborn branch
    Given a Git fixture "unborn"
    When I run git-view with "default"
    Then the exit code is 0
    And stdout contains "Branch: \"main\" (unborn)"
    And stdout contains "State: clean"
    And stderr is empty
    And the fixture is unchanged

  Scenario: Discover from a nested directory
    Given a Git fixture "nested"
    When I run git-view with "status"
    Then the exit code is 0
    And stdout contains "Branch: \"main\""
    And stdout contains "Upstream: none"
    And stdout contains "State: clean"
    And the fixture is unchanged

  Scenario: Count partial staging separately
    Given a Git fixture "mixed"
    When I run git-view with "status"
    Then the exit code is 0
    And stdout contains "State: dirty"
    And stdout contains "Changes: staged 1, unstaged 2, untracked 1, conflicts 0"
    And the fixture is unchanged

  Scenario: Show status columns and literal unusual paths
    Given a Git fixture "mixed"
    When I run git-view with "changes"
    Then the exit code is 0
    And stdout contains "M M \"tracked.txt\""
    And stdout contains ". D \"deleted.txt\""
    And stdout contains "? ? \"-new Ω file.txt\""
    And stdout excludes "ignored.txt"
    And stderr is empty
    And the fixture is unchanged

  Scenario: Preserve both rename paths
    Given a Git fixture "rename"
    When I run git-view with "changes"
    Then the exit code is 0
    And stdout contains "R . \"tracked.txt\" -> \"renamed Ω file.txt\""
    And the fixture is unchanged

  Scenario: Show staged additions and deletions
    Given a Git fixture "staged"
    When I run git-view with "changes"
    Then the exit code is 0
    And stdout contains "A . \"added.txt\""
    And stdout contains "D . \"deleted.txt\""
    And the fixture is unchanged

  Scenario: Report dirty submodules even when configured to ignore them
    Given a Git fixture "submodule"
    When I run git-view with "changes"
    Then the exit code is 0
    And stdout contains ". M \"child\" [submodule S.MU]"
    And the fixture is unchanged

  Scenario: Detached HEAD
    Given a Git fixture "detached"
    When I run git-view with "status"
    Then the exit code is 0
    And stdout contains "Branch: detached ("
    And stdout contains "State: clean"
    And the fixture is unchanged

  Scenario: A branch can be named like Git's detached marker
    Given a Git fixture "detached name"
    When I run git-view with "status"
    Then the exit code is 0
    And stdout contains "Branch: \"(detached)\""
    And the fixture is unchanged

  Scenario: An unborn branch can be named like Git's detached marker
    Given a Git fixture "unborn detached name"
    When I run git-view with "status"
    Then the exit code is 0
    And stdout contains "Branch: \"(detached)\" (unborn)"
    And the fixture is unchanged

  Scenario: Linked worktree selected after command
    Given a Git fixture "worktree"
    When I run git-view with "selected changes"
    Then the exit code is 0
    And stdout contains "No changes."
    And the fixture is unchanged

  Scenario: Local upstream counts even when Git's status preference turns them off
    Given a Git fixture "upstream"
    When I run git-view with "selected status"
    Then the exit code is 0
    And stdout contains "Upstream: \"origin/main\""
    And stdout contains "Ahead/behind: +1 -0 (local refs)"
    And the fixture is unchanged

  Scenario: A missing promisor blob cannot be fetched during rename inspection
    Given a Git fixture "partial clone"
    When I run git-view with "changes"
    Then the exit code is 1
    And stdout is empty
    And stderr contains "git failed"
    And the fixture is unchanged

  Scenario: Report conflicts as successful data
    Given a Git fixture "conflict"
    When I run git-view with "changes"
    Then the exit code is 0
    And stdout contains "U U \"tracked.txt\""
    And the fixture is unchanged

  Scenario: Count a conflict only once
    Given a Git fixture "conflict"
    When I run git-view with "status"
    Then the exit code is 0
    And stdout contains "Changes: staged 0, unstaged 0, untracked 0, conflicts 1"
    And the fixture is unchanged

  Scenario: Reject a bare repository
    Given a Git fixture "bare"
    When I run git-view with "status"
    Then the exit code is 2
    And stderr contains "working tree required"
    And stdout is empty
    And the fixture is unchanged

  Scenario: Explain a directory outside Git
    Given a Git fixture "outside"
    When I run git-view with "default"
    Then the exit code is 2
    And stderr contains "working tree required"
    And stdout is empty

  Scenario: Help without a repository or Git
    Given a Git fixture "no git"
    When I run git-view with "help"
    Then the exit code is 0
    And stdout contains "Usage: git-view"
    And stderr is empty

  Scenario: Version without a repository or Git
    Given a Git fixture "no git"
    When I run git-view with "version"
    Then the exit code is 0
    And stdout contains "git-view 0.1.0-dev"
    And stderr is empty

  Scenario: Missing Git is actionable
    Given a Git fixture "no git"
    When I run git-view with "default"
    Then the exit code is 2
    And stderr contains "git executable not found"
    And stdout is empty

  Scenario: Invalid options do not inspect a repository
    Given a Git fixture "outside"
    When I run git-view with "invalid"
    Then the exit code is 2
    And stderr contains "unknown option"
    And stdout is empty
