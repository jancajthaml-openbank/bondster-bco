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
}

// NewBondsterClient returns new bondster http client
func NewBondsterClient(gateway string) BondsterClient {
  return BondsterClient{
    gateway: gateway,
    underlying: http.NewHttpClient(),
  }
}

// Post performs http POST request for given url with given body
func (client BondsterClient) Post(url string, body []byte, session *Session) (http.Response, error) {

  headers := map[string]string{
    "device":            session.Device,
    "channeluuid":       session.Channel,
    "x-active-language": "cs",
    "host":              "ib.bondster.com",
    "origin":            client.gateway,
    "referer":           client.gateway + "/cs",
  }

  if session.JWT != nil {
    headers["authorization"] = "Bearer " + session.JWT.Value
  }

  if session.SSID != nil {
    headers["ssid"] = session.SSID.Value
  }

  return client.underlying.Post(client.gateway+url, body, headers)
}

// Get performs http GET request for given url
func (client BondsterClient) Get(url string, session *Session) (http.Response, error) {
  headers := map[string]string{
    "device":            session.Device,
    "channeluuid":       session.Channel,
    "x-active-language": "cs",
    "host":              "ib.bondster.com",
    "origin":            client.gateway,
    "referer":           client.gateway + "/cs",
  }

  if session.JWT != nil {
    headers["authorization"] = "Bearer " + session.JWT.Value
  }

  if session.SSID != nil {
    headers["ssid"] = session.SSID.Value
  }

  return client.underlying.Get(client.gateway+url, headers)
}

// GetSession returns session for bondster client
func (client BondsterClient) GetSession(token model.Token) (*Session, error) {
  session := new(Session)
  session.Device = utils.RandDevice()
  session.Channel = utils.UUID()

  var (
    err      error
    response http.Response
  )

  response, err = client.Post("/proxy/router/api/public/authentication/getLoginScenario", nil, session)
  if err != nil {
    return nil, fmt.Errorf("bondster get login scenario Error %+v", err)
  }
  if response.Status != 200 {
    return nil, fmt.Errorf("bondster get login scenario error %s", response.String())
  }

  var scenario = new(model.LoginScenario)
  err = utils.JSON.Unmarshal(response.Data, scenario)
  if err != nil {
    return nil, fmt.Errorf("bondster unsupported login scenario invalid response %s", response.String())
  }

  if scenario.Value != "USR_PWD" {
    return nil, fmt.Errorf("bondster unsupported login scenario %s", response.String())
  }

  request := whitespaceRegex.ReplaceAllString(fmt.Sprintf(`
    {
      "scenarioCode": "USR_PWD",
      "values": [
        {
          "authDetailType": "USERNAME",
          "value": %s
        },
        {
          "authDetailType": "PWD",
          "value": %s
        }
      ]
    }
  `, token.Username, token.Password), "")

  response, err = client.Post("/proxy/router/api/public/authentication/validateLoginStep", []byte(request), session)
  if err != nil {
    return nil, err
  }
  if response.Status != 200 {
    return nil, fmt.Errorf("bondster validate login step error %s", response.String())
  }

  var webToken = new(WebToken)
  err = utils.JSON.Unmarshal(response.Data, webToken)
  if err != nil {
    return nil, err
    return nil, fmt.Errorf("bondster validate login step invalid response %s", response.String())
  }

  log.Debugf("Logged in with token %s", token.ID)

  session.JWT = &(webToken.JWT)
  session.SSID = &(webToken.SSID)

  return session, nil
}

func (client BondsterClient) GetCurrencies(session *Session) ([]string, error) {
  var (
    err      error
  )

  response, err := client.Post("/proxy/clientusersetting/api/private/market/getContactInformation", nil, session)
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

func (client BondsterClient) GetTransactionIdsInInterval(session *Session, currency string, interval utils.TimeRange) ([]string, error) {
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

  response, err := client.Post("/proxy/mktinvestor/api/private/transaction/search", []byte(request), session)
  if err != nil {
    return nil, fmt.Errorf("bondster get contact information error %+v", err)
  }
  if response.Status != 200 {
    return nil, fmt.Errorf("bondster get contact information error %s", response.String())
  }

  var search = new(model.TransfersSearchResult)
  err = utils.JSON.Unmarshal(response.Data, search)
  if err != nil {
    return nil, err
  }

  return search.IDs, nil
}

func (client BondsterClient) GetTransactionDetails(session *Session, currency string, transactionIds []string) (*model.BondsterImportEnvelope, error) {
  ids := make([]string, len(transactionIds))
  for i, id := range transactionIds {
    ids[i] = "\"" + id + "\","
  }

  request := whitespaceRegex.ReplaceAllString(fmt.Sprintf(`
    {
      "transactionIds": [
        %s
      ]
    }
  `, ids[0:len(ids)-1]), "")

  response, err := client.Post("/proxy/mktinvestor/api/private/transaction/list", []byte(request), session)
  if err != nil {
    return nil, fmt.Errorf("bondster get contact information error %+v", err)
  }
  if response.Status != 200 {
    return nil, fmt.Errorf("bondster get contact information error %s", response.String())
  }

  var envelope = new(model.BondsterImportEnvelope)
  err = utils.JSON.Unmarshal(response.Data, &(envelope.Transactions))
  if err != nil {
    return nil, err
  }
  envelope.Currency = currency

  return envelope, nil
}
