// Copyright (c) 2016-2021, Jan Cajthaml <jan.cajthaml@gmail.com>
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
	"encoding/json"
	"fmt"
	"time"
	_http "net/http"
	"sync"

	"github.com/jancajthaml-openbank/bondster-bco-import/model"
	"github.com/jancajthaml-openbank/bondster-bco-import/support/http"
)

type AuthorizedClient struct {
	gateway    string
	httpClient http.Client
	token      *model.Token
	session    Session
	mutex      sync.RWMutex
}

func NewAuthorizedClient(gateway string, token *model.Token) *AuthorizedClient {
	return &AuthorizedClient{
		gateway:    gateway,
		httpClient: http.NewClient(),
		token:      token,
		session:    NewSession(),
		mutex:      sync.RWMutex{},
	}
}

func (client *AuthorizedClient) login() error {
	if client == nil {
		return fmt.Errorf("nil deference")
	}

	client.session.Clear()

	req, err := http.NewRequest("GET", client.gateway+"/proxy/router/api/public/authentication/getLoginScenario", nil)
	if err != nil {
		return fmt.Errorf("get login scenario error %w", err)
	}

	req.SetHeader("device", client.session.Device)
	req.SetHeader("channeluuid", client.session.Channel)
	req.SetHeader("x-active-language", "cs")
	req.SetHeader("authority", client.gateway)

	resp, err := client.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("get login scenario error %w", err)
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("get login scenario status error %s", resp.Status)
	}

	scenario := new(loginScenario)
	err = json.NewDecoder(resp.Body).Decode(scenario)
	if err != nil {
		return fmt.Errorf("unsupported login scenario invalid response %w", err)
	}

	if scenario.value != "USR_PWD" {
		return fmt.Errorf("unsupported login scenario %s", scenario.value)
	}

	loginScenario := loginScenarioAnswer{
		ScenarioCode: "USR_PWD",
		AuthProcessStepValues: []loginScenarioStep{
			{
				AuthDetailType: "USERNAME",
				Value: client.token.Username,
			},
			{
				AuthDetailType: "PWD",
				Value: client.token.Password,
			},
		},
	}

	payload, err := json.Marshal(loginScenario)
	if err != nil {
		return fmt.Errorf("unsupported login scenario %w", err)
	}

  	req, err = http.NewRequest("POST", client.gateway+"/proxy/router/api/public/authentication/validateLoginStep", payload)
	if err != nil {
		return fmt.Errorf("validate login step error %w", err)
	}
	req.SetHeader("Content-Type", "application/json")
	req.SetHeader("device", client.session.Device)
	req.SetHeader("channeluuid", client.session.Channel)
	req.SetHeader("x-active-language", "cs")
	req.SetHeader("authority", client.gateway)

	resp, err = client.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("validate login step error %w", err)
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("validate login step status error %s", resp.Status)
	}

	webToken := new(WebToken)
	err = json.NewDecoder(resp.Body).Decode(webToken)
	if err != nil {
		return fmt.Errorf("validate login step invalid response with error %w", err)
	}

	client.session.JWT = &(webToken.JWT)
	client.session.SSID = &(webToken.SSID)

	log.Debug().Msgf("logged in with token %s, valid until %s", client.token.ID, webToken.JWT.ExpiresAt.Format(time.RFC3339))

	return nil
}

func (client *AuthorizedClient) prolong() error {
	if client == nil {
		return fmt.Errorf("nil defference")
	}

	req, err := http.NewRequest("POST", client.gateway+"/proxy/router/api/private/token/prolong", nil)
	if err != nil {
		return fmt.Errorf("get login scenario Error %w", err)
	}

	req.SetHeader("device", client.session.Device)
	req.SetHeader("channeluuid", client.session.Channel)
	req.SetHeader("x-active-language", "cs")
	req.SetHeader("authority", client.gateway)

	if client.session.JWT != nil {
		// FIXME possible nil
		req.SetHeader("authorization", "Bearer " + client.session.JWT.Value)
	}

	if client.session.SSID != nil {
		// FIXME possible nil
		req.SetHeader("ssid", client.session.SSID.Value)
	}

	resp, err := client.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("get prolong token error %w", err)
	}

	if resp.StatusCode == 401 {
		return fmt.Errorf("prolong failed")
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("get prolong token error %s", resp.Status)
	}

	// fixme to type struct
	all := struct {
		JWT struct {
			Value     string `json:"value"`
			ExpiresAt string `json:"expirationDate"`
		} `json:"jwtToken"`
	}{}

	err = json.NewDecoder(resp.Body).Decode(&all)
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

	log.Info().Msgf("session for %s prolonged, valid until %s", client.token.ID, jwtExpiration.Format(time.RFC3339))

	return nil
}

// Do perform authorized http.Request
func (client *AuthorizedClient) Do(req *http.Request) (*_http.Response, error) {
	if client == nil || req == nil {
		return nil, fmt.Errorf("nil deference")
	}

	var err error

	client.mutex.Lock()
	if client.session.IsJWTExpired() {
		err = client.prolong()
	}
	if err != nil || client.session.IsSSIDExpired() {
		err = client.login()
	}
	client.mutex.Unlock()

	if err != nil {
		return nil, err
	}

	req.SetHeader("device", client.session.Device)
	req.SetHeader("channeluuid", client.session.Channel)
	req.SetHeader("x-active-language", "cs")
	req.SetHeader("origin", client.gateway)
	req.SetHeader("referer", client.gateway + "/cs")
	
	client.mutex.Lock()
	if client.session.JWT != nil {
		req.SetHeader("authorization", "Bearer " + client.session.JWT.Value)
	}
	if client.session.SSID != nil {
		req.SetHeader("ssid", client.session.SSID.Value)
	}
	client.mutex.Unlock()

	resp, err := client.httpClient.Do(req)

	if resp.StatusCode == 401 {
		log.Warn().Msgf("authorization lost")
		client.mutex.Lock()
		client.session.Clear()
		client.mutex.Unlock()
		return client.Do(req)
	}

	return resp, err
}
