Feature: Messaging behaviour

  Scenario: token
    Given tenant MSG1 is onboarded

    When lake recieves "BondsterImport/MSG1 BondsterRest token_1 req_id_1 NT X Y"
    Then lake responds with "BondsterRest BondsterImport/MSG1 req_id_1 token_1 TN"

    When lake recieves "BondsterImport/MSG1 BondsterRest token_2 req_id_2 DT"
    Then lake responds with "BondsterRest BondsterImport/MSG1 req_id_2 token_2 EM"

    When lake recieves "BondsterImport/MSG1 BondsterRest token_3 req_id_3 NT X Y"
    Then lake responds with "BondsterRest BondsterImport/MSG1 req_id_3 token_3 TN"

    When lake recieves "BondsterImport/MSG1 BondsterRest token_3 req_id_3 DT"
    Then lake responds with "BondsterRest BondsterImport/MSG1 req_id_3 token_3 TD"

