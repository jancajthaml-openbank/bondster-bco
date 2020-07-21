Feature: Messaging behaviour

  Scenario: token
    Given tenant MSG1 is onboarded

    When lake recieves "BondsterImport/MSG1 BondsterRest token_1 req_id_1 NT X Y"
    Then lake responds with "BondsterRest BondsterImport/MSG1 req_id_1 token_1 TN"

    When lake recieves "BondsterImport/MSG1 BondsterRest token_2 req_id_2 NT X Y"
    And  lake recieves "BondsterImport/MSG1 BondsterRest token_2 req_id_2 NT X Y"
    Then lake responds with "BondsterRest BondsterImport/MSG1 req_id_2 token_2 TN"
    And  lake responds with "BondsterRest BondsterImport/MSG1 req_id_2 token_2 EE"

    When lake recieves "BondsterImport/MSG1 BondsterRest token_3 req_id_3 DT"
    Then lake responds with "BondsterRest BondsterImport/MSG1 req_id_3 token_3 EE"

    When lake recieves "BondsterImport/MSG1 BondsterRest token_4 req_id_4 NT X Y"
    Then lake responds with "BondsterRest BondsterImport/MSG1 req_id_4 token_4 TN"
