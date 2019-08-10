Feature: Install package

  Scenario: install
    Given package bondster-bco is installed
    Then  systemctl contains following active units
      | name              | type    |
      | bondster-bco-rest | service |
      | bondster-bco      | service |
      | bondster-bco      | path    |
