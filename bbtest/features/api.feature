Feature: REST

  Scenario: Tenant API test
    Given unit "bondster-bco-rest.service" is running

    When I request HTTP https://127.0.0.1/tenant
      | key    | value |
      | method | GET   |
    Then HTTP response is
      | key    | value |
      | status | 200   |
      """
      []
      """

    When I request HTTP https://127.0.0.1/tenant/APITESTA
      | key    | value |
      | method | POST  |
    Then HTTP response is
      | key    | value |
      | status | 200   |

    When I request HTTP https://127.0.0.1/tenant/APITESTB
      | key    | value |
      | method |  POST |
    Then HTTP response is
      | key    | value |
      | status | 200   |

    When I request HTTP https://127.0.0.1/tenant
      | key    | value |
      | method | GET   |
    Then HTTP response is
      | key    | value |
      | status | 200   |
      """
      [
        "APITESTB"
      ]
      """

    When I request HTTP https://127.0.0.1/tenant/APITESTC
      | key    | value |
      | method | POST  |
    Then HTTP response is
      | key    | value |
      | status | 200   |

    When I request HTTP https://127.0.0.1/tenant/APITESTC
      | key    | value  |
      | method | DELETE |
    Then HTTP response is
      | key    | value  |
      | status | 200    |


  Scenario: Token API
    Given unit "bondster-bco-rest.service" is running
    And tenant API is onboarded

    When I request HTTP https://127.0.0.1/token/API
      | key    | value |
      | method | GET   |
    Then HTTP response is
      | key    | value |
      | status | 200   |
      """
      []
      """

    When I request HTTP https://127.0.0.1/token/API
      | key    | value |
      | method | POST  |
      """
      {
        "username": "X",
        "password": "Y"
      }
      """
    Then HTTP response is
      | key    | value |
      | status | 200   |


    When I request HTTP https://127.0.0.1/token/API
      | key    | value |
      | method | POST  |
      """
      {
        "username": "X",
        "password": "Y"
      }
      """
    Then HTTP response is
      | key    | value |
      | status | 200   |

    When I request HTTP https://127.0.0.1/token/API
      | key    | value |
      | method | POST  |
      """
      {
        "username": "",
        "password": ""
      }
      """
    Then HTTP response is
      | key    | value |
      | status | 400   |

    When I request HTTP https://127.0.0.1/token/API
      | key    | value |
      | method | GET   |
    Then HTTP response is
      | key    | value |
      | status | 200   |


  Scenario: Health API
    Given unit "bondster-bco-rest.service" is running

    When I request HTTP https://127.0.0.1/health
      | key    | value |
      | method | GET   |
    Then HTTP response is
      | key    | value |
      | status | 200   |
