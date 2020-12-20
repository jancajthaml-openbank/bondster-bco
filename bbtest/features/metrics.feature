Feature: Metrics test

  Scenario: metrics measures expected stats
    Given tenant M2 is onboarded

    Then metrics reports:
      | key                                        | type  |      tags | value |
      | openbank.bco.bondster.token.created        | count | tenant:M2 |     0 |
      | openbank.bco.bondster.token.deleted        | count | tenant:M2 |     0 |
      | openbank.bco.bondster.transaction.imported | count | tenant:M2 |       |
      | openbank.bco.bondster.transfer.imported    | count | tenant:M2 |       |

    When token M2/A is created

    Then metrics reports:
      | key                                 | type  |      tags | value |
      | openbank.bco.bondster.token.created | count | tenant:M2 |     1 |
