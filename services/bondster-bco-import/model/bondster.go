// Copyright (c) 2016-2019, Jan Cajthaml <jan.cajthaml@gmail.com>
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
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/jancajthaml-openbank/bondster-bco-import/utils"
)

// BondsterImportEnvelope represents bondster marketplace import statement entity
type BondsterImportEnvelope struct {
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

// GetTransactions return list of bondster transactions
func (envelope *BondsterImportEnvelope) GetTransactions(tenant string) []Transaction {
	if envelope == nil {
		return nil
	}

	var set = make(map[string][]Transfer)

	sort.SliceStable(envelope.Transactions, func(i, j int) bool {
		return envelope.Transactions[i].ValueDate.Before(envelope.Transactions[j].ValueDate)
	})

	var nostro = envelope.Currency + "_NOSTRO"

	for _, transfer := range envelope.Transactions {
		valueDate := transfer.ValueDate.Format("2006-01-02T15:04:05Z0700")

		if transfer.Direction == "DEBIT" {
			if transfer.Originator != nil {
				set[transfer.IDTransaction] = append(set[transfer.IDTransaction], Transfer{
					IDTransfer: transfer.IDTransfer,
					Credit: AccountPair{
						Tenant: tenant,
						Name:   envelope.Currency + "_" + transfer.Originator.Name,
					},
					Debit: AccountPair{
						Tenant: tenant,
						Name:   nostro,
					},
					ValueDate:    valueDate,
					ValueDateRaw: transfer.ValueDate,
					Amount:       transfer.Amount.Value,
					Currency:     transfer.Amount.Currency,
				})
				set[transfer.IDTransaction] = append(set[transfer.IDTransaction], Transfer{
					IDTransfer: transfer.IDTransfer + "_FWD",
					Credit: AccountPair{
						Tenant: tenant,
						Name:   envelope.Currency + "_" + transfer.Type,
					},
					Debit: AccountPair{
						Tenant: tenant,
						Name:   envelope.Currency + "_" + transfer.Originator.Name,
					},
					ValueDate:    valueDate,
					ValueDateRaw: transfer.ValueDate,
					Amount:       transfer.Amount.Value,
					Currency:     transfer.Amount.Currency,
				})
			} else {
				set[transfer.IDTransaction] = append(set[transfer.IDTransaction], Transfer{
					IDTransfer: transfer.IDTransfer,
					Credit: AccountPair{
						Tenant: tenant,
						Name:   envelope.Currency + "_" + transfer.Type,
					},
					Debit: AccountPair{
						Tenant: tenant,
						Name:   nostro,
					},
					ValueDate:    valueDate,
					ValueDateRaw: transfer.ValueDate,
					Amount:       transfer.Amount.Value,
					Currency:     transfer.Amount.Currency,
				})
			}
		} else {
			if transfer.Originator != nil {
				set[transfer.IDTransaction] = append(set[transfer.IDTransaction], Transfer{
					IDTransfer: transfer.IDTransfer,
					Credit: AccountPair{
						Tenant: tenant,
						Name:   nostro,
					},
					Debit: AccountPair{
						Tenant: tenant,
						Name:   envelope.Currency + "_ORIGINATOR_" + transfer.Originator.Name,
					},
					ValueDate:    valueDate,
					ValueDateRaw: transfer.ValueDate,
					Amount:       transfer.Amount.Value,
					Currency:     transfer.Amount.Currency,
				})
				set[transfer.IDTransaction] = append(set[transfer.IDTransaction], Transfer{
					IDTransfer: transfer.IDTransfer + "_FWD",
					Credit: AccountPair{
						Tenant: tenant,
						Name:   envelope.Currency + "_ORIGINATOR_" + transfer.Originator.Name,
					},
					Debit: AccountPair{
						Tenant: tenant,
						Name:   envelope.Currency + "_TYPE_" + transfer.Type,
					},
					ValueDate:    valueDate,
					ValueDateRaw: transfer.ValueDate,
					Amount:       transfer.Amount.Value,
					Currency:     transfer.Amount.Currency,
				})
			} else {
				set[transfer.IDTransaction] = append(set[transfer.IDTransaction], Transfer{
					IDTransfer: transfer.IDTransfer,
					Credit: AccountPair{
						Tenant: tenant,
						Name:   nostro,
					},
					Debit: AccountPair{
						Tenant: tenant,
						Name:   envelope.Currency + "_TYPE_" + transfer.Type,
					},
					ValueDate:    valueDate,
					ValueDateRaw: transfer.ValueDate,
					Amount:       transfer.Amount.Value,
					Currency:     transfer.Amount.Currency,
				})
			}
		}
	}

	result := make([]Transaction, 0)
	for transaction, transfers := range set {
		result = append(result, Transaction{
			IDTransaction: transaction,
			Transfers:     transfers,
		})
	}

	return result

}

// GetAccounts return list of bondster accounts
func (envelope *BondsterImportEnvelope) GetAccounts() []Account {
	if envelope == nil {
		return nil
	}

	var deduplicated = make(map[string]interface{})

	deduplicated[envelope.Currency+"_TYPE_NOSTRO"] = nil

	for _, transfer := range envelope.Transactions {
		if transfer.Originator != nil {
			deduplicated[envelope.Currency+"_ORIGINATOR_"+transfer.Originator.Name] = nil
		}
		deduplicated[envelope.Currency+"_TYPE_"+transfer.Type] = nil
		if transfer.External != nil {
			deduplicated[NormalizeAccountNumber(transfer.External.Number, transfer.External.BankCode)] = nil
		}
	}

	result := make([]Account, 0)
	for account := range deduplicated {
		result = append(result, Account{
			Name:           account,
			Currency:       envelope.Currency,
			IsBalanceCheck: false,
		})
	}

	return result
}

// LoginStep satisfaction of login scenario
type LoginStep struct {
	Code   string           `json:"scenarioCode"`
	Values []LoginStepValue `json:"authProcessStepValues"`
}

// LoginStepValue value of login step
type LoginStepValue struct {
	Type  string `json:"authDetailType"`
	Value string `json:"value"`
}

// LoginScenario holds code representing how service should log in
type LoginScenario struct {
	Value string
}

// TransfersSearchRequest request for search transfed
type TransfersSearchRequest struct {
	From time.Time
	To   time.Time
}

// TransfersSearchResult result of search transfers request
type TransfersSearchResult struct {
	IDs []string `json:"transferIdList"`
}

// MarshalJSON is json TransfersSearchResult marhalling companion
func (entity TransfersSearchResult) MarshalJSON() ([]byte, error) {
	ids := make([]string, len(entity.IDs))
	for i, id := range entity.IDs {
		ids[i] = "\"" + id + "\""
	}
	return []byte("{\"transactionIds\":[" + strings.Join(ids, ",") + "]}"), nil
}

// MarshalJSON is json TransfersSearchRequest marhalling companion
func (entity TransfersSearchRequest) MarshalJSON() ([]byte, error) {
	return []byte("{\"valueDateFrom\":{\"month\":\"" + strconv.FormatInt(int64(entity.From.Month()), 10) + "\",\"year\":\"" + strconv.FormatInt(int64(entity.From.Year()), 10) + "\"},\"valueDateTo\":{\"month\":\"" + strconv.FormatInt(int64(entity.To.Month()), 10) + "\",\"year\":\"" + strconv.FormatInt(int64(entity.To.Year()), 10) + "\"}}"), nil
}

// UnmarshalJSON is json LoginScenario unmarhalling companion
func (entity *LoginScenario) UnmarshalJSON(data []byte) error {
	if entity == nil {
		return fmt.Errorf("cannot unmarshall to nil pointer")
	}
	all := struct {
		Scenarios []struct {
			Code string `json:"code"`
		} `json:"scenarios"`
	}{}
	err := utils.JSON.Unmarshal(data, &all)
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

// JWT encrypted json web token of bondster session
type JWT struct {
	Value string
}

// UnmarshalJSON is json JWT unmarhalling companion
func (entity *JWT) UnmarshalJSON(data []byte) error {
	if entity == nil {
		return fmt.Errorf("cannot unmarshall to nil pointer")
	}
	all := struct {
		Result string `json:"result"`
		JWT    struct {
			Value string `json:"value"`
		} `json:"jwt"`
	}{}
	err := utils.JSON.Unmarshal(data, &all)
	if err != nil {
		return err
	}
	if all.Result != "FINISH" {
		return fmt.Errorf("result %s has not finished, bailing out", all.Result)
	}
	if all.JWT.Value == "" {
		return fmt.Errorf("missing \"jwt\" value field")
	}
	entity.Value = all.JWT.Value
	return nil
}

// Session hold bondster session headers
type Session struct {
	JWT     string
	Device  string
	Channel string
}

// PotrfolioCurrencies hold currencies of account portfolio
type PotrfolioCurrencies struct {
	Value []string
}

// UnmarshalJSON is json PotrfolioCurrencies unmarhalling companion
func (entity *PotrfolioCurrencies) UnmarshalJSON(data []byte) error {
	if entity == nil {
		return fmt.Errorf("cannot unmarshall to nil pointer")
	}
	all := struct {
		MarketAccounts struct {
			AccountsMap map[string]interface{} `json:"currencyToAccountMap"`
		} `json:"marketVerifiedExternalAccount"`
	}{}
	err := utils.JSON.Unmarshal(data, &all)
	if err != nil {
		return err
	}
	entity.Value = make([]string, 0)
	for currency := range all.MarketAccounts.AccountsMap {
		entity.Value = append(entity.Value, currency)
	}
	return nil
}
