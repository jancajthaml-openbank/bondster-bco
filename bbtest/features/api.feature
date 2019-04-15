Feature: REST

  Scenario: Tenant API test
    Given bondster-bco is reconfigured with
    """
      LOG_LEVEL=DEBUG
      HTTP_PORT=443
    """

    When I request curl GET https://127.0.0.1/tenant
    Then curl responds with 200
    """
      []
    """

    When I request curl POST https://127.0.0.1/tenant/APITESTA
    Then curl responds with 200
    """
      {}
    """

    When I request curl POST https://127.0.0.1/tenant/APITESTB
    Then curl responds with 200
    """
      {}
    """

    When I request curl GET https://127.0.0.1/tenant
    Then curl responds with 200
    """
      [
        "APITESTB"
      ]
    """

    When I request curl POST https://127.0.0.1/tenant/APITESTC
    Then curl responds with 200
    """
      {}
    """

    When I request curl DELETE https://127.0.0.1/tenant/APITESTC
    Then curl responds with 200
    """
      {}
    """

  Scenario: Token API
    Given tenant API is onbdoarded
    And bondster-bco is reconfigured with
    """
      LOG_LEVEL=DEBUG
      HTTP_PORT=443
    """

    When I request curl GET https://127.0.0.1/token/API
    Then curl responds with 200
    """
      []
    """

    When I request curl POST https://127.0.0.1/token/API
    """
      {
        "username": "X",
        "password": "Y"
      }
    """
    Then curl responds with 200

    When I request curl POST https://127.0.0.1/token/API
    """
      {
        "username": "X",
        "password": "Y"
      }
    """
    Then curl responds with 200

    When I request curl GET https://127.0.0.1/token/API
    Then curl responds with 200

