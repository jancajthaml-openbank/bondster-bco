Feature: Properly behaving units

  Scenario: onboard
    Given tenant lorem is onbdoarded
    And tenant ipsum is onbdoarded
    Then systemctl contains following
    """
      bondster-bco-import@lorem.service
      bondster-bco-import@ipsum.service
      bondster-bco-rest.service
      bondster-bco.service
    """

    When stop unit "bondster-bco-rest.service"
    Then unit "bondster-bco-rest.service" is not running

    When start unit "bondster-bco-rest.service"
    Then unit "bondster-bco-rest.service" is running

    When restart unit "bondster-bco-rest.service"
    Then unit "bondster-bco-rest.service" is running

    When stop unit "bondster-bco-import@lorem.service"
    Then unit "bondster-bco-import@lorem.service" is not running

    When start unit "bondster-bco-import@lorem.service"
    Then unit "bondster-bco-import@lorem.service" is running

    When restart unit "bondster-bco-import@ipsum.service"
    Then unit "bondster-bco-import@ipsum.service" is running

  Scenario: offboard
    Given tenant lorem is offboarded
    And tenant ipsum is offboarded

    Then systemctl does not contains following
    """
      bondster-bco-import@lorem.service
      bondster-bco-import@ipsum.service
    """
    And systemctl contains following
    """
      bondster-bco-rest.service
      bondster-bco.service
    """
