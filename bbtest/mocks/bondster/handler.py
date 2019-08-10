from http.server import BaseHTTPRequestHandler
import json


class RequestHandler(BaseHTTPRequestHandler):

  def log_message(self, format, *args):
    pass

  def __get_login_scenarion(self):
    device = self.headers.get('device')
    channel = self.headers.get('channeluuid')

    # fixme validate channel and device

    response = self.server.logic.get_login_scenario()

    self.__respond(200, response)

  def validate_device(self):
    device = self.headers.get('device')
    return True

  def validate_channel(self):
    channel = self.headers.get('channeluuid')
    return True

  def validate_authorisation(self):
    pass

  def validate_currency_context(self):
    pass

  def __validate_login_step(self):
    if not self.validate_device():
      return self.__respond(500)

    if not self.validate_channel():
      return self.__respond(500)

    request = json.loads(self.rfile.read(int(self.headers['Content-Length'])).decode('utf-8'))

    response = self.server.logic.validate_login_step(request)
    if not response:
      return self.__respond(400)

    self.__respond(200, response)

  def __get_contain_information(self):
    if not self.validate_device():
      return self.__respond(500)

    if not self.validate_channel():
      return self.__respond(500)

    authorization = self.headers.get('authorization')

    # fixme validate authorisation

    response = self.server.logic.get_contain_information()

    self.__respond(200, response)

  def __transaction_search(self):
    if not self.validate_device():
      return self.__respond(500)

    if not self.validate_channel():
      return self.__respond(500)

    authorization = self.headers.get('authorization')

    # fixme validate authorisation

    currency = self.headers.get('x-account-context')

    # fixme validate currency

    request = json.loads(self.rfile.read(int(self.headers['Content-Length'])).decode('utf-8'))

    valueDateFrom = request.get('valueDateFrom')
    valueDateTo = request.get('valueDateTo')

    response = self.server.logic.transaction_search(currency, valueDateFrom, valueDateTo)

    self.__respond(200, response)

  def __transaction_list(self):
    if not self.validate_device():
      return self.__respond(500)

    if not self.validate_channel():
      return self.__respond(500)

    authorization = self.headers.get('authorization')

    # fixme validate authorisation

    currency = self.headers.get('x-account-context')

    # fixme validate currency

    request = json.loads(self.rfile.read(int(self.headers['Content-Length'])).decode('utf-8'))

    transactions = request.get('transactionIds', [])
    response = self.server.logic.transaction_list(transactions)

    self.__respond(200, response)

  def do_POST(self):
    handler = {
      '/router/api/public/authentication/getLoginScenario': self.__get_login_scenarion,
      '/router/api/public/authentication/validateLoginStep': self.__validate_login_step,
      '/clientusersetting/api/private/market/getContactInformation': self.__get_contain_information,
      '/mktinvestor/api/private/transaction/search': self.__transaction_search,
      '/mktinvestor/api/private/transaction/list': self.__transaction_list,
    }.get(self.path, None)

    if handler:
      handler()
      return

    self.__respond(404)
    return

  def __respond(self, status, body=None):
    self.send_response(status)
    self.send_header('Content-type','application/json')
    self.end_headers()
    if body:
      self.wfile.write(json.dumps(body).encode('utf-8'))
