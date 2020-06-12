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

package http

import (
  "fmt"

  "github.com/jancajthaml-openbank/bondster-bco-import/model"
  "github.com/jancajthaml-openbank/bondster-bco-import/utils"
)

// BondsterClient represents fascade for http client
type BondsterClient struct {
  underlying HttpClient
  gateway string
}

// NewBondsterClient returns new bondster http client
func NewBondsterClient(gateway string) BondsterClient {
  return BondsterClient{
    gateway: gateway,
    underlying: NewHttpClient(),
  }
}

// Post performs http POST request for given url with given body
func (client BondsterClient) Post(url string, body []byte, headers map[string]string) (Response, error) {
  return client.Post(client.gateway+url, body, headers)
}

// Get performs http GET request for given url
func (client BondsterClient) Get(url string, headers map[string]string) (Response, error) {
  return client.Get(client.gateway+url, headers)
}


// GetSession returns session for bondster client
func (client BondsterClient) GetSession(token model.Token) (*Session, error) {
  device := utils.RandDevice()
  channel := utils.UUID()

  var (
    err      error
    response Response
    request  []byte
    uri      string
  )

  headers := map[string]string{
    "device":            device,
    "channeluuid":       channel,
    "x-active-language": "cs",
    "host":              "ib.bondster.com",
    "origin":            "https://ib.bondster.com",
    "referer":           "https://ib.bondster.com/cs",
  }

  uri = "/proxy/router/api/public/authentication/getLoginScenario"
  response, err = client.Post(uri, nil, headers)
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

  step := model.LoginStep{
    Code: "USR_PWD",
    Values: []model.LoginStepValue{
      {
        Type:  "USERNAME",
        Value: token.Username,
      },
      {
        Type:  "PWD",
        Value: token.Password,
      },
    },
  }

  // FIXME if re-captcha then handle

  request, err = utils.JSON.Marshal(step)
  if err != nil {
    return nil, err
  }

  uri = "/proxy/router/api/public/authentication/validateLoginStep"
  response, err = client.Post(uri, request, headers)
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

  session := &Session{
    JWT:     webToken.JWT,
    Device:  device,
    Channel: channel,
    SSID:    webToken.SSID,
  }

  return session, nil
}


func (client BondsterClient)  GetCurrencies(session *Session) ([]string, error) {
  var (
    err      error
    uri      string
  )

  uri = "/proxy/clientusersetting/api/private/market/getContactInformation"

  headers := map[string]string{
    "device":            session.Device,
    "channeluuid":       session.Channel,
    "ssid":              session.SSID,
    "x-active-language": "cs",
    "authorization":     "Bearer " + session.JWT,
    "host":              "ib.bondster.com",
    "origin":            "https://ib.bondster.com",
    "referer":           "https://ib.bondster.com/cs",
  }

  response, err := client.Post(uri, nil, headers)
  if err != nil {
    return nil, fmt.Errorf("bondster get contact information error %+v", err)
  }
  if response.Status != 200 {
    return nil, fmt.Errorf("bondster get contact information error %s", response.String())
  }

  var currencies = new(model.PotrfolioCurrencies)
  err = utils.JSON.Unmarshal(response.Data, currencies)
  if err != nil {
    return nil, err
  }

  return currencies.Value, nil
}
