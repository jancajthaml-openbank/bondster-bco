Feature: Properly behaving units

  Scenario: onboard
    Given tenant lorem is onbdoarded
    And tenant ipsum is onbdoarded
    Then systemctl contains following
    """
      bondster-bco.service
      bondster-bco@lorem.service
      bondster-bco@ipsum.service
    """

    When stop unit "bondster-bco.service"
    Then unit "bondster-bco.service" is not running

    When start unit "bondster-bco.service"
    Then unit "bondster-bco.service" is running

    When restart unit "bondster-bco.service"
    Then unit "bondster-bco.service" is running

    When stop unit "bondster-bco@lorem.service"
    Then unit "bondster-bco@lorem.service" is not running

    When start unit "bondster-bco@lorem.service"
    Then unit "bondster-bco@lorem.service" is running

    When restart unit "bondster-bco@ipsum.service"
    Then unit "bondster-bco@ipsum.service" is running

  Scenario: offboard
    Given tenant lorem is offboarded
    And tenant ipsum is offboarded

    Then systemctl does not contains following
    """
      bondster-bco@lorem.service
      bondster-bco@ipsum.service
    """
    And systemctl contains following
    """
      bondster-bco.service
    """
