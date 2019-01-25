require 'json'
require 'thread'
require_relative '../shims/harden_webrick'
require_relative './bondster_mock'

class BondsterGetCurrenciesLimitHandler < WEBrick::HTTPServlet::AbstractServlet

  def do_POST(request, response)
    status, content_type, body = process(request)

    response.status = status
    response['Content-Type'] = content_type
    response.body = body
  end

  def process(request)
    return 500, {} unless request.header.key?("channeluuid")
    return 500, {} unless request.header.key?("device")
    return 401, {} unless request.header.key?("authorization")

    return 200, "application/json", {
      "EUR": {
        "minInvestment": 0.01,
        "maxInvestment": 10000000,
        "maxInvestmentPercentage": 100,
        "defaultInvestment": 5
      },
      "CZK": {
        "minInvestment": 0.01,
        "maxInvestment": 10000000,
        "maxInvestmentPercentage": 100,
        "defaultInvestment": 100
      }
    }.to_json
  end
end

class BondsterListTransactionHandler < WEBrick::HTTPServlet::AbstractServlet

  def do_POST(request, response)
    status, content_type, body = process(request)

    response.status = status
    response['Content-Type'] = content_type
    response.body = body
  end

  def process(request)
    return 500, {} unless request.header.key?("channeluuid")
    return 500, {} unless request.header.key?("device")
    return 401, {} unless request.header.key?("authorization")
    return 401, {} unless request.header.key?("x-account-context")

    return 200, "application/json", [
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
        "storno": false
      }
    ].to_json
  end
end

class BondsterSearchTransactionHandler < WEBrick::HTTPServlet::AbstractServlet

  def do_POST(request, response)
    status, content_type, body = process(request)

    response.status = status
    response['Content-Type'] = content_type
    response.body = body
  end

  def process(request)
    return 500, "application/json", {} unless request.header.key?("channeluuid")
    return 500, "application/json", {} unless request.header.key?("device")
    return 401, "application/json", {} unless request.header.key?("authorization")
    return 401, "application/json", {} unless request.header.key?("x-account-context")

    body = Hash.new

    begin
      body = JSON.parse(request.body)
      raise "" unless body.key?("valueDateFrom") and body.key?("valueDateTo")
    rescue Exception
      return 400, "application/json", {}
    end

    currency = request.header["x-account-context"]

    return 200, "application/json", {
      "transferIdList": [
        "a",
        "b",
        "c",
        "d",
        "e"
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
    }.to_json
  end
end

class BondsterGetLoginScenarioHandler < WEBrick::HTTPServlet::AbstractServlet

  def do_POST(request, response)
    status, content_type, body = process(request)

    response.status = status
    response['Content-Type'] = content_type
    response.body = body
  end

  def process(request)

    return 500, "application/json", {} unless request.header.key?("channeluuid")
    return 500, "application/json", {} unless request.header.key?("device")

    #puts request.header["authorization"]

    #puts "headers: #{request.header}"

    # fixme validate headers

    return 200, "application/json", {
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
    }.to_json
  end
end

class BondsterValidateLoginStepHandler < WEBrick::HTTPServlet::AbstractServlet

  def do_POST(request, response)
    status, content_type, body = process(request)

    response.status = status
    response['Content-Type'] = content_type
    response.body = body
  end

  def process(request)

    return 500, "application/json", {} unless request.header.key?("channeluuid")
    return 500, "application/json", {} unless request.header.key?("device")

    return 200, "application/json", {
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
    }.to_json

  end
end


module BondsterHelper

  def self.start
    self.server = nil

    begin
      self.server = WEBrick::HTTPServer.new(
        Port: 4000,
        Logger: WEBrick::Log.new("/dev/null"),
        AccessLog: [],
        SSLEnable: true
      )

    rescue Exception => err
      raise err
      raise "Failed to allocate server binding! #{err}"
    end

    self.server.mount "/router/api/public/authentication/getLoginScenario", BondsterGetLoginScenarioHandler
    self.server.mount "/router/api/public/authentication/validateLoginStep", BondsterValidateLoginStepHandler
    self.server.mount "/mktinvestor/api/private/transaction/search", BondsterSearchTransactionHandler
    self.server.mount "/mktinvestor/api/private/transaction/list", BondsterListTransactionHandler
    self.server.mount "/mktinvestor/api/private/investor/limits", BondsterGetCurrenciesLimitHandler



    self.server_daemon = Thread.new do
      self.server.start()
    end
  end

  def self.stop
    self.server.shutdown() unless self.server.nil?
    begin
      self.server_daemon.join() unless self.server_daemon.nil?
    rescue
    ensure
      self.server_daemon = nil
      self.server = nil
    end
  end

  class << self
    attr_accessor :server_daemon, :server
  end

end
