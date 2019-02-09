Feature: Bondster Marketplace import

  Scenario: import from gateway token
    Given bondster gateway contains following statements
    """
      []
    """
    Given tenant IMPORT is onbdoarded
    And bondster-bco is reconfigured with
    """
      SYNC_RATE=1s
      HTTP_PORT=443
    """
    And token IMPORT/importToken is created
    And I sleep for 5 seconds
