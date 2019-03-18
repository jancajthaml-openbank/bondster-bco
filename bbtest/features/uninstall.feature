@uninstall
Feature: Uninstall package

  Scenario: uninstall
    Given package "bondster-bco" is uninstalled
    Then  systemctl does not contains following
    """
      bondster-bco.service
      bondster-bco.path
      bondster-bco-rest.service
    """
