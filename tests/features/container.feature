Feature: Build and test against a mounted Git repository

  Scenario: Create and remove a seeded fixture volume
    Given the container workflow uses a seeded fixture
    When I run the container build and test workflow
    Then the container workflow succeeds
    And the container output contains "Branch: \"main\""
    And the container output contains "M M \"tracked.txt\""
    And the container output contains "? ? \"new note.txt\""
    And its seeded volume has been removed

  Scenario: Inspect an existing host checkout
    Given the container workflow uses a host fixture
    When I run the container build and test workflow
    Then the container workflow succeeds
    And the container output contains "M M \"tracked.txt\""
    And the container output contains "? ? \"-new Ω file.txt\""
    And the host fixture is unchanged
