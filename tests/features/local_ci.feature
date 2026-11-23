Feature: Local container CI and delivery

  Scenario: Deliver a candidate from a push to the local Git server
    Given the local Compose pipeline is configured
    When I bootstrap a committed checkout and wait for its Actions run
    Then the pushed commit succeeds and delivers a candidate artifact
    And repeating the bootstrap verifies the same run
