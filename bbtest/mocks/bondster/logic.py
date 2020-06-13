import json


class BussinessLogic(object):

  def __init__(self):
    pass

  def get_login_scenario(self):
    return {
      "scenarios": [
        {
          "code": "USR_PWD",
          "steps": [
            {
              "details": [
                {
                  "code": "USERNAME"
                },
                {
                  "code": "PWD"
                }
              ]
            }
          ]
        }
      ]
    }

  def validate_login_step(self, data):
    code = data['scenarioCode']

    if code != 'USR_PWD':
      return False

    username = [item['value'] for item in data['authProcessStepValues'] if item['authDetailType'] == 'USERNAME'][0]
    password = [item['value'] for item in data['authProcessStepValues'] if item['authDetailType'] == 'PWD'][0]

    if not (username == 'X' and password == 'Y'):
      return False

    return {
      "result": "FINISH",
      "idAuthProcess": "rIw5_8ARbFY2YrL8TG_UQ7vIPof3KuoiRt6YT27S75Y_fPx-iWXEwl36vHTerr8JSqnlGkMpkfMvLPWhFAlskw==", # fixme generate
      "nextSMSAfter": 0, # fixme generate
      "jwt": {
        "value": "rM9aAc3ccNFg8A37gE3TN24rPb34zViEyg5YnmUB1ygj5YxSWZ86zAIMt9O2ONQIZ2XclAqaxp9CC2YSGTggumzv+i5cWj+ntJWqHC4/cuxvM70NOES+50JhVBitJC92dBeSjRo7Xg9M+5kcCHeeU5eP7JiMmzlEKptdHQW2sY3G+m2acfiG4BR1VV6hLkoL00Zl5nZixtGEm+Sx/E4yz6wqhq0O9ykB8Wg18w/ZuJAT4ZvYjDbuJisKaTgk5rIB7/V3GdRLjJzHwRjeG9dnltWyVcE6wdOB3nWc9pX6x+0azpQTcAar9GVfb0aM1V/NGK4goqNXALljg5DQBj6FWAUW11DfN+a3K9rr0G2RkR8dY2jVRXAylVv9KW7d6y5TYYTYNekxjGzTafrDAxwslKYWPJh9VCjUfUZJCee/ip1uijmJw5EoxbojApQB/FzZAVu6+qdx5cta/LCrxmPuTI0GyxcJEWSOxilxMtf5fyOPePmm00ZAU8Iu+qKQdwPgo0XVAnNZS6gQm6VO+jSfzJjKv/vrr54GX9HXbIbsqeloDDoo8WbJAZlK3CEwmMix4BB6pne2FXe9RRv7ltBr1r3WXOBf3zQcmF8DPbYbLF36BpLHFT5YQBbvTD0jRag4BY0tJoqyFXJhIq1ybGmut/xYVKE5/X3kP7nyY0UwvZOgwInDLVG9yot37rEUbf6GDFnqWmKdk7iFegtoKYMr+2Yx8uIRqYRn80hTzGIYMcE=",  # fixme generate
        "expirationDate": "2019-01-17T18:32:23.173Z" # fixme generate
      },
      "ssid": {
        "value": "vWfbuAdFXdzpPK6JtKAy4QI7qVU=", # fixme generate
        "expirationDate": "2019-01-18T00:22:23.152Z" # fixme generate
      }
    }

  def get_contain_information(self):
    return {
      "status": "VERIFIED",
      "marketVerifiedExternalAccount": {
        "status": "APPROVED",
        "currencyToAccountMap": {
          "EUR": [
            {
              "portfolioCurrency": "EUR",
              "bankCode": "1111",
              "accountNumber": "2222",
              "status": "APPROVED",
              "accountNumberFormat": "IBAN"
            }
          ],
          "CZK": [
            {
              "portfolioCurrency": "CZK",
              "bankCode": "1111",
              "accountNumber": "2222",
              "status": "APPROVED",
              "accountNumberFormat": "LOCALCZ"
            }
          ]
        }
      }
    }

  def transaction_search(self, currency, valueDateFrom, valueDateTo):
    # fixme filter transactions based on valueDateFrom and valueDateTo

    return {
      "transferIdList": [
        "a",
        "b"
      ],
      "summary": {
        "startingBalance": 0,
        "finalBalance": 0,
        "currencyCode": currency,
        "principalInstallmentSum": 0,
        "principalOtherSum": 0,
        "interestSum": 0,
        "sanctionSum": 0,
        "investorDepositSum": 0,
        "investorWithdrawalSum": 0
      }
    }

  def transaction_list(self, transactions):
    # fixme generate transactions based on ids (pick or filter by ids)

    return [
      {
        "idTransaction": "x",
        "idTransfer": "y",
        "direction": "CREDIT",
        "valueDate": "2018-05-28T08:56:15.683Z",
        "transactionType": "type",
        "loanNumber": "0",
        "internalId": "1",
        "amount": {
          "amount": 10.0,
          "currencyCode": "CZK"
        },
        "originator": {
          "idOriginator": "123",
          "originatorName": "xzy"
        },
        "storno": False
      }
    ]

