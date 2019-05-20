Feature: Metrics test

  Scenario: metrics have expected keys
    And   tenant M2 is onbdoarded
    And   bondster-bco is reconfigured with
    """
      METRICS_REFRESHRATE=1s
    """

    Then metrics file /reports/metrics.M2.json should have following keys:
    """
      createdTokens
      deletedTokens
      importedTransactions
      importedTransfers
      listTransactionLatency
      searchTransactionLatency
    """