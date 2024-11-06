Feature: List Namespaces

  Scenario: user get namespace
    Given user has access to a namespace
    Then  the user can retrieve the namespace
