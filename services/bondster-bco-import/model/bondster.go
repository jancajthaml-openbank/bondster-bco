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

package model

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/jancajthaml-openbank/bondster-bco-import/utils"
)

// WebToken encrypted json web token and ssid of bondster session
type WebToken struct {
	JWT  valueWithExpiration
	SSID valueWithExpiration
}

type valueWithExpiration struct {
	Value     string
	ExpiresAt time.Time
}

// Session hold bondster session headers
type Session struct {
	JWT     *valueWithExpiration
	SSID    *valueWithExpiration
	Device  string
	Channel string
}

// NewSession returns new authenticated client session
func NewSession() Session {
	return Session{
		JWT:     nil,
		SSID:    nil,
		Device:  utils.RandDevice(),
		Channel: utils.UUID(),
	}
}

// IsSSIDExpired tells whenever session is expired
func (session Session) IsSSIDExpired() bool {
	if session.JWT == nil || session.SSID == nil {
		return true
	}
	if time.Now().After(session.SSID.ExpiresAt.Add(time.Second * time.Duration(-10))) {
		return true
	}
	return false
}

// IsJWTExpired tells whenever JsonWebToken is expired
func (session Session) IsJWTExpired() bool {
	if session.JWT == nil {
		return true
	}
	if time.Now().After(session.JWT.ExpiresAt.Add(time.Second * time.Duration(-10))) {
		return true
	}
	return false
}

// UnmarshalJSON is json JWT unmarhalling companion
func (entity *WebToken) UnmarshalJSON(data []byte) error {
	if entity == nil {
		return fmt.Errorf("cannot unmarshal to nil pointer")
	}
	all := struct {
		Result string `json:"result"`
		JWT    struct {
			Value     string `json:"value"`
			ExpiresAt string `json:"expirationDate"`
		} `json:"jwt"`
		SSID struct {
			Value     string `json:"value"`
			ExpiresAt string `json:"expirationDate"`
		} `json:"ssid"`
	}{}
	err := json.Unmarshal(data, &all)
	if err != nil {
		return err
	}
	if all.Result != "FINISH" {
		return fmt.Errorf("result %s has not finished, bailing out", all.Result)
	}
	if all.JWT.Value == "" || all.JWT.ExpiresAt == "" {
		return fmt.Errorf("missing \"jwt\" value field")
	}
	if all.SSID.Value == "" || all.SSID.ExpiresAt == "" {
		return fmt.Errorf("missing \"ssid\" value field")
	}

	jwtExpiration, err := time.Parse("2006-01-02T15:04:05.000Z", all.JWT.ExpiresAt)
	if err != nil {
		return err
	}

	ssidExpiration, err := time.Parse("2006-01-02T15:04:05.000Z", all.SSID.ExpiresAt)
	if err != nil {
		return err
	}

	entity.JWT = valueWithExpiration{
		Value:     all.JWT.Value,
		ExpiresAt: jwtExpiration,
	}
	entity.SSID = valueWithExpiration{
		Value:     all.SSID.Value,
		ExpiresAt: ssidExpiration,
	}

	return nil
}

// BondsterStatement repsenents result of /proxy/mktinvestor/api/private/transaction/list
type BondsterStatement struct {
	IDTransaction string                   `json:"idTransaction"`
	IDTransfer    string                   `json:"idTransfer"`
	Type          string                   `json:"transactionType"`
	Direction     string                   `json:"direction"`
	LoanID        *string                  `json:"loanNumber"`
	ValueDate     time.Time                `json:"valueDate"`
	Originator    *bondsterOriginator      `json:"originator"`
	External      *bondsterExternalAccount `json:"externalAccount"`
	Amount        bondsterAmount           `json:"amount"`
	IsStorno      bool                     `json:"storno"`
}

type bondsterExternalAccount struct {
	Number   string `json:"accountNumber"`
	BankCode string `json:"bankCode"`
}

type bondsterOriginator struct {
	ID   string `json:"idOriginator"`
	Name string `json:"originatorName"`
}

type bondsterAmount struct {
	Value    float64 `json:"amount"`
	Currency string  `json:"currencyCode"`
}

// LoginScenario holds code representing how service should log in
type LoginScenario struct {
	Value string
}

// UnmarshalJSON is json LoginScenario unmarhalling companion
func (entity *LoginScenario) UnmarshalJSON(data []byte) error {
	if entity == nil {
		return fmt.Errorf("cannot unmarshal to nil pointer")
	}
	all := struct {
		Scenarios []struct {
			Code string `json:"code"`
		} `json:"scenarios"`
	}{}
	err := json.Unmarshal(data, &all)
	if err != nil {
		return err
	}
	if len(all.Scenarios) == 0 {
		return fmt.Errorf("no login scenarios")
	}
	if all.Scenarios[0].Code == "" {
		return fmt.Errorf("missing \"code\" value field")
	}
	entity.Value = all.Scenarios[0].Code
	return nil
}
