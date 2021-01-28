Feature: Bondster Marketplace import

  Scenario: import from gateway token
    Given bondster gateway contains following statements
      | key       | value |
    And bondster-bco is configured with
      | property  | value |
      | SYNC_RATE |    1s |
    And tenant IMPORT is onboarded
    And token IMPORT/importToken is created
