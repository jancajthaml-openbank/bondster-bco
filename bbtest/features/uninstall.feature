Feature: Uninstall package

  Scenario: uninstall
    Given package bondster-bco is uninstalled
    Then  systemctl does not contain following active units
      | name              | type    |
      | bondster-bco-rest | service |
      | bondster-bco      | service |
      | bondster-bco      | path    |
