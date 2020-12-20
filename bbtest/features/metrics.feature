Feature: Metrics test

  Scenario: metrics measures expected stats
    Given tenant M2 is onboarded

    Then metrics reports:
      | key                                           | type  | value |
      | openbank.bco.bondster.M2.token.created        | count |     0 |
      | openbank.bco.bondster.M2.token.deleted        | count |     0 |
      | openbank.bco.bondster.M2.transaction.imported | count |       |
      | openbank.bco.bondster.M2.transfer.imported    | count |       |

    When token M2/A is created

    Then metrics reports:
      | key                                           | type  | value |
      | openbank.bco.bondster.M2.token.created        | count |     1 |
      | openbank.bco.bondster.M2.token.deleted        | count |     0 |
      | openbank.bco.bondster.M2.transaction.imported | count |       |
      | openbank.bco.bondster.M2.transfer.imported    | count |       |
