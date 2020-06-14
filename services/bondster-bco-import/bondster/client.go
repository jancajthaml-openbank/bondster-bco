// Copyright (c) 2016-2020, Jan Cajthaml <jan.cajthaml@gmail.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package bondster

import (
  "fmt"
  "time"
  "regexp"

  "github.com/jancajthaml-openbank/bondster-bco-import/model"
  "github.com/jancajthaml-openbank/bondster-bco-import/utils"
  "github.com/jancajthaml-openbank/bondster-bco-import/http"
)

var whitespaceRegex = regexp.MustCompile(`\s`)

// BondsterClient represents fascade for http client
type BondsterClient struct {
  underlying http.HttpClient
  gateway string
  token model.Token
  session *Session
}

// NewBondsterClient returns new bondster http client
func NewBondsterClient(gateway string, token model.Token) BondsterClient {
  return BondsterClient{
    gateway: gateway,
    underlying: http.NewHttpClient(),
    token: token,
    session: nil,
  }
}

// Post performs http POST request for given url with given body
func (client BondsterClient) Post(url string, body []byte) (http.Response, error) {
  headers := map[string]string{
    "device":            client.session.Device,
    "channeluuid":       client.session.Channel,
    "x-active-language": "cs",
    "host":              "ib.bondster.com",
    "origin":            client.gateway,
    "referer":           client.gateway + "/cs",
  }

  if client.session.JWT != nil {
    headers["authorization"] = "Bearer " + client.session.JWT.Value
  }

  if client.session.SSID != nil {
    headers["ssid"] = client.session.SSID.Value
  }

  return client.underlying.Post(client.gateway+url, body, headers)
}

// Get performs http GET request for given url
func (client BondsterClient) Get(url string, session *Session) (http.Response, error) {
  headers := map[string]string{
    "device":            client.session.Device,
    "channeluuid":       client.session.Channel,
    "x-active-language": "cs",
    "host":              "ib.bondster.com",
    "origin":            client.gateway,
    "referer":           client.gateway + "/cs",
  }

  if client.session.JWT != nil {
    headers["authorization"] = "Bearer " + client.session.JWT.Value
  }

  if client.session.SSID != nil {
    headers["ssid"] = client.session.SSID.Value
  }

  return client.underlying.Get(client.gateway+url, headers)
}


// GetSession returns session for bondster client
func (client *BondsterClient) checkSession() error {
  if client == nil {
    return fmt.Errorf("nil defference")
  }
  if client.session != nil {
    return nil
  }
  if client.session == nil || client.session.IsSSIDExpired() {
    return client.login()
  }
  if client.session.IsJWTExpired() {
    return client.prolong()
  }
  return nil
}

// FIXME tied to session
func (client *BondsterClient) login() error {
  if client == nil {
    return fmt.Errorf("nil defference")
  }

  session := NewSession()
  client.session = &session

  var (
    err      error
    response http.Response
  )

  response, err = client.Post("/proxy/router/api/public/authentication/getLoginScenario", nil)
  if err != nil {
    return fmt.Errorf("bondster get login scenario Error %+v", err)
  }
  if response.Status != 200 {
    return fmt.Errorf("bondster get login scenario error %s", response.String())
  }

  var scenario = new(LoginScenario)
  err = utils.JSON.Unmarshal(response.Data, scenario)
  if err != nil {
    return fmt.Errorf("bondster unsupported login scenario invalid response %s", response.String())
  }

  if scenario.Value != "USR_PWD" {
    return fmt.Errorf("bondster unsupported login scenario %s", response.String())
  }

  request := whitespaceRegex.ReplaceAllString(fmt.Sprintf(`
    {
      "scenarioCode": "USR_PWD",
      "authProcessStepValues": [
        {
          "authDetailType": "USERNAME",
          "value": "%s"
        },
        {
          "authDetailType": "PWD",
          "value": "%s"
        }
      ]
    }
  `, client.token.Username, client.token.Password), "")

  response, err = client.Post("/proxy/router/api/public/authentication/validateLoginStep", []byte(request))
  if err != nil {
    return err
  }
  if response.Status != 200 {
    return fmt.Errorf("bondster validate login step error %s", response.String())
  }

  var webToken = new(WebToken)
  err = utils.JSON.Unmarshal(response.Data, webToken)
  if err != nil {
    return fmt.Errorf("bondster validate login step invalid response %s with error %+v", response.String(), err)
  }

  log.Debugf("Logged in with token %s", client.token.ID)

  client.session.JWT = &(webToken.JWT)
  client.session.SSID = &(webToken.SSID)

  return nil
}

// FIXME tied to session
func (client *BondsterClient) prolong() error {
  // Request URL: https://ib.bondster.com/proxy/router/api/private/token/prolong
  // {"jwtToken":{"value":"xxx","expirationDate":"2020-06-14T08:57:15.911Z"}}

  if client == nil {
    return fmt.Errorf("nil defference")
  }

  session := NewSession()
  client.session = &session

  var (
    err      error
    response http.Response
  )

  response, err = client.Post("/proxy/router/api/private/token/prolong", nil)
  if err != nil {
    return fmt.Errorf("bondster get prolong token Error %+v", err)
  }
  if response.Status != 200 {
    return fmt.Errorf("bondster get prolong token error %s", response.String())
  }

  all := struct {
    JWT struct {
      Value string `json:"value"`
      ExpiresAt string `json:"expirationDate"`
    } `json:"jwtToken"`
  }{}

  err = utils.JSON.Unmarshal(response.Data, &all)
  if err != nil {
    return err
  }

  if all.JWT.Value == "" {
    return fmt.Errorf("missing \"jwtToken\" value field")
  }

  jwtExpiration, err := time.Parse("2006-01-02T15:04:05.000Z", all.JWT.ExpiresAt)
  if err != nil {
    return err
  }

  client.session.JWT.Value = all.JWT.Value
  client.session.JWT.ExpiresAt = jwtExpiration

  return nil
}

func (client BondsterClient) GetCurrencies() ([]string, error) {
  err := client.checkSession()
  if err != nil {
    return nil, err
  }

  response, err := client.Post("/proxy/clientusersetting/api/private/market/getContactInformation", nil)
  if err != nil {
    return nil, fmt.Errorf("bondster get contact information error %+v", err)
  }
  if response.Status != 200 {
    return nil, fmt.Errorf("bondster get contact information error %s", response.String())
  }

  all := struct {
    MarketAccounts struct {
      AccountsMap map[string]interface{} `json:"currencyToAccountMap"`
    } `json:"marketVerifiedExternalAccount"`
  }{}

  err = utils.JSON.Unmarshal(response.Data, &all)
  if err != nil {
    return nil, err
  }

  currencies := make([]string, 0)
  for currency := range all.MarketAccounts.AccountsMap {
    currencies = append(currencies, currency)
  }

  return currencies, nil
}

func (client BondsterClient) GetTransactionIdsInInterval(currency string, interval utils.TimeRange) ([]string, error) {
  err := client.checkSession()
  if err != nil {
    return nil, err
  }
  request := whitespaceRegex.ReplaceAllString(fmt.Sprintf(`
    {
      "valueDateFrom": {
        "month": %d,
        "year": %d
      },
      "valueDateTo": {
        "month": %d,
        "year": %d
      }
    }
  `, interval.StartTime.Month(), interval.StartTime.Year(), interval.EndTime.Month(), interval.EndTime.Year()), "")

  response, err := client.Post("/proxy/mktinvestor/api/private/transaction/search", []byte(request))
  if err != nil {
    return nil, fmt.Errorf("bondster get contact information error %+v", err)
  }
  if response.Status != 200 {
    return nil, fmt.Errorf("bondster get contact information error %s", response.String())
  }

  all := struct {
    IDs []string `json:"transferIdList"`
  }{}

  err = utils.JSON.Unmarshal(response.Data, &all)
  if err != nil {
    return nil, err
  }

  return all.IDs, nil
}

func (client BondsterClient) GetTransactionDetails(currency string, transactionIds []string) (*BondsterImportEnvelope, error) {
  err := client.checkSession()
  if err != nil {
    return nil, err
  }
  ids := ""
  for _, id := range transactionIds {
    ids += "\"" + id + "\","
  }

  request := whitespaceRegex.ReplaceAllString(fmt.Sprintf(`
    {
      "transactionIds": [
        %s
      ]
    }
  `, ids[0:len(ids)-1]), "")

  response, err := client.Post("/proxy/mktinvestor/api/private/transaction/list", []byte(request))
  if err != nil {
    return nil, fmt.Errorf("bondster get contact information error %+v", err)
  }
  if response.Status != 200 {
    return nil, fmt.Errorf("bondster get contact information error %s", response.String())
  }

  var envelope = new(BondsterImportEnvelope)
  err = utils.JSON.Unmarshal(response.Data, &(envelope.Transactions))
  if err != nil {
    return nil, err
  }
  envelope.Currency = currency

  return envelope, nil
}
