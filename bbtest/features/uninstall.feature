Feature: Uninstall package

  Scenario: uninstall
    Given lake is not running
    And   package bondster-bco is uninstalled
    Then  systemctl does not contain following active units
      | name                 | type    |
      | bondster-bco         | service |
      | bondster-bco-rest    | service |
      | bondster-bco-watcher | path    |
      | bondster-bco-watcher | service |