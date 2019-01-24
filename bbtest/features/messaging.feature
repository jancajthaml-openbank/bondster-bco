Feature: Messaging behaviour

  Scenario: create new token
    Given tenant MSG is onbdoarded
    When tenant MSG receives "token req_id_1 NT token_1 X Y"
    Then tenant MSG responds with "token req_id_1 TN"
    And no other messages were received

  Scenario: do not create already existing token
    Given tenant MSG is onbdoarded
    When tenant MSG receives "token req_id_2 NT token_2 X Y"
    And tenant MSG receives "token req_id_2 NT token_2 X Y"
    Then tenant MSG responds with "token req_id_2 TN"
    And tenant MSG responds with "token req_id_2 EE"
    And no other messages were received

  Scenario: delete existing token
    Given tenant MSG is onbdoarded
    When tenant MSG receives "token req_id_3 NT token_3 X Y"
    And tenant MSG receives "token req_id_3 DT token_3"
    Then tenant MSG responds with "token req_id_3 TN"
    And tenant MSG responds with "token req_id_3 TD"
    And no other messages were received

  Scenario: do not delete non existing token
    Given tenant MSG is onbdoarded
    When tenant MSG receives "token req_id_4 DT token_4"
    Then tenant MSG responds with "token req_id_4 EE"
    And no other messages were received
