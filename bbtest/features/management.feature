Feature: System control

  Scenario: check units presence
    Then  systemctl contains following active units
      | name              | type    |
      | bondster-bco      | service |
      | bondster-bco-rest | service |
  
  Scenario: onboard
    Given tenant lorem is onboarded
    And   tenant ipsum is onboarded
    Then  systemctl contains following active units
      | name                      | type    |
      | bondster-bco-import@lorem | service |
      | bondster-bco-import@ipsum | service |
    And unit "bondster-bco-import@lorem.service" is running
    And unit "bondster-bco-import@ipsum.service" is running

  Scenario: stop
    When restart unit "bondster-bco.service"
    Then unit "bondster-bco-import@lorem.service" is not running
    And  unit "bondster-bco-import@ipsum.service" is not running

  Scenario: start
    When restart unit "bondster-bco.service"
    Then unit "bondster-bco-import@lorem.service" is running
    And  unit "bondster-bco-import@ipsum.service" is running

  Scenario: restart
    When restart unit "bondster-bco.service"
    Then unit "bondster-bco-import@lorem.service" is running
    And  unit "bondster-bco-import@ipsum.service" is running

  Scenario: offboard
    Given tenant lorem is offboarded
    And   tenant ipsum is offboarded
    Then  systemctl does not contain following active units
      | name                      | type    |
      | bondster-bco-import@lorem | service |
      | bondster-bco-import@ipsum | service |
    And systemctl contains following active units
      | name                      | type    |
      | bondster-bco              | service |
      | bondster-bco-rest         | service |
