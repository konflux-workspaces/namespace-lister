Feature: List Namespaces

  Scenario: user list namespace
    Given user has access to "10" namespaces
    Then  the user can retrieve only the namespaces they have access to
