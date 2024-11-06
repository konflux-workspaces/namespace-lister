Feature: List Namespaces

  Scenario: user list namespace
    Given user has access to a namespace
    Then  the user can retrieve the namespace
