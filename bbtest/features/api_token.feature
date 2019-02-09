Feature: Token API test

  Scenario: Token API - get tokens when application is from scratch
    Given tenant API is onbdoarded
    And bondster-bco is reconfigured with
    """
      LOG_LEVEL=DEBUG
      HTTP_PORT=443
    """

    When I request curl GET https://localhost/token/API
    Then curl responds with 200
    """
      []
    """

  Scenario: Token API - create non existant token
    Given tenant API is onbdoarded
    And bondster-bco is reconfigured with
    """
      LOG_LEVEL=DEBUG
      HTTP_PORT=443
    """

    When I request curl POST https://localhost/token/API
    """
      {
        "username": "X",
        "password": "Y"
      }
    """
    Then curl responds with 200

  Scenario: Token API - get tokens
    Given tenant API is onbdoarded
    And bondster-bco is reconfigured with
    """
      LOG_LEVEL=DEBUG
      HTTP_PORT=443
    """

    When I request curl POST https://localhost/token/API
    """
      {
        "username": "X",
        "password": "Y"
      }
    """
    Then curl responds with 200

    When I request curl GET https://localhost/token/API
    Then curl responds with 200
