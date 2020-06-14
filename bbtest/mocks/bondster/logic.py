import json
import datetime


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

  def prolong_token(self):
    return {
      "jwtToken": {
        "value": "jwt",
        "expirationDate": (now + datetime.timedelta(hours=1)).strftime("%Y-%m-%dT%H:%M%S.%.3fZ")
      }
    }

  def validate_login_step(self, data):
    code = data['scenarioCode']

    if code != 'USR_PWD':
      return False

    username = [item['value'] for item in data['authProcessStepValues'] if item['authDetailType'] == 'USERNAME'][0]
    password = [item['value'] for item in data['authProcessStepValues'] if item['authDetailType'] == 'PWD'][0]

    if not (username == 'X' and password == 'Y'):
      return False

    now = datetime.datetime.now()

    return {
      "result": "FINISH",
      "idAuthProcess": "id",
      "nextSMSAfter": 0,
      "jwt": {
        "value": "jwt",
        "expirationDate": (now + datetime.timedelta(seconds=30)).strftime("%Y-%m-%dT%H:%M%S.%.3fZ")
      },
      "ssid": {
        "value": "ssid",
        "expirationDate": (now + datetime.timedelta(hours=1)).strftime("%Y-%m-%dT%H:%M%S.%.3fZ")
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

