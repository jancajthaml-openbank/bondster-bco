@install
Feature: Install package

  Scenario: install
    Given package "bondster-bco.deb" is installed
    Then  systemctl contains following
    """
      bondster-bco.service
      bondster-bco.path
      bondster-bco-rest.service
    """
