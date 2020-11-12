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
	"encoding/json"
	"fmt"
	"github.com/jancajthaml-openbank/bondster-bco-import/model"
	"strconv"
	"time"
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

// ImportEnvelope represents bondster marketplace import statement entity
type ImportEnvelope struct {
	Transactions []bondsterTransaction
	Currency     string
}

type bondsterTransaction struct {
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

// GetTransactions return generator of bondster transactions over given envelope
func (envelope *ImportEnvelope) GetTransactions(tenant string) <-chan model.Transaction {
	chnl := make(chan model.Transaction)
	if envelope == nil {
		close(chnl)
		return chnl
	}

	var previousIDTransaction = ""
	var buffer = make([]model.Transfer, 0)

	go func() {
		defer close(chnl)

		for _, transfer := range envelope.Transactions {
			if transfer.IsStorno {
				log.Warn().Msgf("storno transfer %+v", transfer)
			}

			valueDate := transfer.ValueDate.Format("2006-01-02T15:04:05Z0700")

			credit := model.AccountPair{
				Tenant: tenant,
			}
			debit := model.AccountPair{
				Tenant: tenant,
			}

			if transfer.Direction == "CREDIT" {
				if transfer.Originator != nil {
					credit.Name = envelope.Currency + "_TYPE_" + transfer.Type
					debit.Name = envelope.Currency + "_ORIGINATOR_" + transfer.Originator.Name
				} else {
					credit.Name = envelope.Currency + "_TYPE_NOSTRO"
					debit.Name = envelope.Currency + "_TYPE_" + transfer.Type
				}
			} else {
				if transfer.Originator != nil {
					credit.Name = envelope.Currency + "_ORIGINATOR_" + transfer.Originator.Name
					debit.Name = envelope.Currency + "_TYPE_" + transfer.Type
				} else {
					credit.Name = envelope.Currency + "_TYPE_" + transfer.Type
					debit.Name = envelope.Currency + "_TYPE_NOSTRO"
				}
			}

			idTransaction := transfer.IDTransaction

			if previousIDTransaction == "" {
				previousIDTransaction = idTransaction
			} else if previousIDTransaction != idTransaction {
				transfers := make([]model.Transfer, len(buffer))
				copy(transfers, buffer)
				buffer = make([]model.Transfer, 0)
				chnl <- model.Transaction{
					Tenant:        tenant,
					IDTransaction: previousIDTransaction,
					Transfers:     transfers,
				}
				previousIDTransaction = idTransaction
			}

			buffer = append(buffer, model.Transfer{
				IDTransfer:   transfer.IDTransfer,
				Credit:       credit,
				Debit:        debit,
				ValueDate:    valueDate,
				ValueDateRaw: transfer.ValueDate,
				Amount:       strconv.FormatFloat(transfer.Amount.Value, 'f', -1, 64),
				Currency:     transfer.Amount.Currency,
			})
		}

		if len(buffer) == 0 {
			return
		}

		transfers := make([]model.Transfer, len(buffer))
		copy(transfers, buffer)
		buffer = make([]model.Transfer, 0)
		chnl <- model.Transaction{
			Tenant:        tenant,
			IDTransaction: previousIDTransaction,
			Transfers:     transfers,
		}

	}()

	return chnl
}

// GetAccounts return generator of bondster accounts over given envelope
func (envelope *ImportEnvelope) GetAccounts(tenant string) <-chan model.Account {
	chnl := make(chan model.Account)
	if envelope == nil {
		close(chnl)
		return chnl
	}

	var visited = make(map[string]interface{})

	go func() {
		defer close(chnl)

		chnl <- model.Account{
			Tenant:         tenant,
			Name:           envelope.Currency + "_TYPE_NOSTRO",
			Format:         "BONDSTER_TECHNICAL",
			Currency:       envelope.Currency,
			IsBalanceCheck: false,
		}

		for _, transfer := range envelope.Transactions {
			if transfer.Originator != nil {
				if _, ok := visited[envelope.Currency+"_ORIGINATOR_"+transfer.Originator.Name]; !ok {
					chnl <- model.Account{
						Tenant:         tenant,
						Name:           envelope.Currency + "_ORIGINATOR_" + transfer.Originator.Name,
						Format:         "BONDSTER_ORIGINATOR",
						Currency:       envelope.Currency,
						IsBalanceCheck: false,
					}
					visited[envelope.Currency+"_ORIGINATOR_"+transfer.Originator.Name] = nil
				}
			}

			if _, ok := visited[envelope.Currency+"_TYPE_"+transfer.Type]; !ok {
				chnl <- model.Account{
					Tenant:         tenant,
					Name:           envelope.Currency + "_TYPE_" + transfer.Type,
					Format:         "BONDSTER_TECHNICAL",
					Currency:       envelope.Currency,
					IsBalanceCheck: false,
				}
				visited[envelope.Currency+"_TYPE_"+transfer.Type] = nil
			}
		}
	}()

	return chnl
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
