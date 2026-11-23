Feature: Use the core v1 workflows in a local worktree
  Scenario: Browse the worktree including ignored files
    Given a Git fixture "mixed"
    When I run git-view with "files"
    Then the exit code is 0
    And stdout contains "ignored file \"ignored.txt\""
    And stdout contains "tracked file \"tracked.txt\""
    And stdout contains "untracked file \"-new Ω file.txt\""
    And the fixture is unchanged

  Scenario: Inspect a commit page
    Given a Git fixture "upstream"
    When I run git-view with "log --limit 1 --skip 1"
    Then the exit code is 0
    And stdout contains "base"
    And stdout excludes "ahead"
    And the fixture is unchanged

  Scenario: Follow a single file
    Given a Git fixture "mixed"
    When I run git-view with "history -- tracked.txt"
    Then the exit code is 0
    And stdout contains "Fixture"
    And stdout contains "base"
    And the fixture is unchanged

  Scenario: Empty history on an unborn branch
    Given a Git fixture "unborn"
    When I run git-view with "log"
    Then the exit code is 0
    And stdout contains "No commits."

  Scenario: Inspect a commit patch
    Given a Git fixture "mixed"
    When I run git-view with "show HEAD"
    Then the exit code is 0
    And stdout contains "+base"
    And the fixture is unchanged

  Scenario: Inspect only the staged version
    Given a Git fixture "mixed"
    When I run git-view with "diff --staged -- tracked.txt"
    Then the exit code is 0
    And stdout contains "+staged"
    And stdout excludes "unstaged"
    And the fixture is unchanged

  Scenario: Inspect working changes
    Given a Git fixture "mixed"
    When I run git-view with "diff -- tracked.txt"
    Then the exit code is 0
    And stdout contains "+unstaged"
    And the fixture is unchanged

  Scenario: Inspect the available conflict stages
    Given a Git fixture "conflict"
    When I run git-view with "conflicts"
    Then the exit code is 0
    And stdout contains "\"tracked.txt\""
    And stdout contains "base"
    And stdout contains "ours"
    And stdout contains "theirs"
    And the fixture is unchanged

  Scenario: Explicitly stage just one path
    Given a Git fixture "mixed"
    When I run git-view with "stage --yes -- tracked.txt"
    Then the exit code is 0
    When I run git-view with "changes"
    Then stdout contains "M . \"tracked.txt\""
    And stdout contains ". D \"deleted.txt\""

  Scenario: Unstage while keeping working edits
    Given a Git fixture "mixed"
    When I run git-view with "unstage --yes -- tracked.txt"
    Then the exit code is 0
    When I run git-view with "changes"
    Then stdout contains ". M \"tracked.txt\""

  Scenario Outline: Live and tool sessions require a terminal
    Given a Git fixture "mixed"
    When I run git-view with "<command>"
    Then the exit code is 2
    And stderr contains "terminal"
    And the fixture is unchanged

    Examples:
      | command                     |
      | watch --views status,files  |
      | interactive                 |
      | edit -- tracked.txt         |
