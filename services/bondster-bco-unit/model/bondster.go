// Copyright (c) 2016-2018, Jan Cajthaml <jan.cajthaml@gmail.com>
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

	"github.com/jancajthaml-openbank/bondster-bco-unit/utils"
)

type BondsterImportEnvelope struct {
	Transactions []BondsterTransaction
	Currency     string
}

type BondsterTransaction struct {
	IdTransaction string                   `json:"idTransaction"`
	IdTransfer    string                   `json:"idTransfer"`
	Type          string                   `json:"transactionType"`
	Direction     string                   `json:"direction"`
	ValueDate     time.Time                `json:"valueDate"`
	External      *BondsterExternalAccount `json:"externalAccount"`
	Amount        BondsterAmount           `json:"amount"`
}

type BondsterExternalAccount struct {
	Number   string `json:"accountNumber"`
	BankCode string `json:"bankCode"`
}

type BondsterAmount struct {
	Value    float64 `json:"amount"`
	Currency string  `json:"currencyCode"`
}

func (envelope *BondsterImportEnvelope) GetTransactions() []Transaction {
	if envelope == nil {
		return nil
	}

	var set = make(map[string][]Transfer)

	sort.SliceStable(envelope.Transactions, func(i, j int) bool {
		return envelope.Transactions[i].ValueDate.Before(envelope.Transactions[j].ValueDate)
	})

	var nostro = envelope.Currency + "_NOSTRO"
	var credit string
	var debit string

	for _, transfer := range envelope.Transactions {
		if transfer.Direction == "DEBIT" {
			credit = envelope.Currency + "_" + transfer.Type
			debit = nostro
		} else {
			credit = nostro
			debit = envelope.Currency + "_" + transfer.Type
		}

		set[transfer.IdTransaction] = append(set[transfer.IdTransaction], Transfer{
			IDTransfer:   transfer.IdTransfer,
			Credit:       credit,
			Debit:        debit,
			ValueDate:    transfer.ValueDate.Format("2006-01-02T15:04:05Z0700"),
			ValueDateRaw: transfer.ValueDate,
			Amount:       transfer.Amount.Value,
			Currency:     transfer.Amount.Currency,
		})
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

func (envelope *BondsterImportEnvelope) GetAccounts() []Account {
	if envelope == nil {
		return nil
	}

	var deduplicated = make(map[string]interface{})

	deduplicated[envelope.Currency+"_NOSTRO"] = nil
	deduplicated[envelope.Currency+"_INVESTOR_INVESTMENT_FEE"] = nil
	deduplicated[envelope.Currency+"_PRIMARY_MARKET_FINANCIAL"] = nil
	deduplicated[envelope.Currency+"_PRINCIPAL_PAYMENT_FINANCIAL"] = nil
	deduplicated[envelope.Currency+"_INVESTOR_DEPOSIT"] = nil
	deduplicated[envelope.Currency+"_SANCTION_PAYMENT"] = nil

	for _, transfer := range envelope.Transactions {
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

type LoginStep struct {
	Code   string           `json:"scenarioCode"`
	Values []LoginStepValue `json:"authProcessStepValues"`
}

type LoginStepValue struct {
	Type  string `json:"authDetailType"`
	Value string `json:"value"`
}

type LoginScenario struct {
	Value string
}

type TransfersSearchRequest struct {
	From time.Time
	To   time.Time
}

type TransfersSearchResult struct {
	IDs []string `json:"transferIdList"`
}

func (entity *TransfersSearchResult) MarshalJSON() ([]byte, error) {
	return []byte("{\"transactionIds\":[" + strings.Join(entity.IDs, ",") + "]}"), nil
}

func (entity *TransfersSearchRequest) MarshalJSON() ([]byte, error) {
	return []byte("{\"valueDateFrom\":\"{\"month\":\"" + strconv.FormatInt(int64(entity.From.Month()), 10) + "\",\"year\":\"" + strconv.FormatInt(int64(entity.From.Year()), 10) + "\"},\"valueDateTo\":\"{\"month\":\"" + strconv.FormatInt(int64(entity.To.Month()), 10) + "\",\"year\":\"" + strconv.FormatInt(int64(entity.To.Year()), 10) + "\"}}"), nil
}

func (entity *LoginScenario) UnmarshalJSON(data []byte) error {
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
		return fmt.Errorf("No login scenarios")
	}
	if all.Scenarios[0].Code == "" {
		return fmt.Errorf("Missing code value field")
	}
	entity.Value = all.Scenarios[0].Code
	return nil
}

type JWT struct {
	Value string
}

func (entity *JWT) UnmarshalJSON(data []byte) error {
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
		return fmt.Errorf("Result %+v does not represent ready session credentials", all.Result)
	}
	if all.JWT.Value == "" {
		return fmt.Errorf("Missing jwt value field")
	}
	entity.Value = all.JWT.Value
	return nil
}

type Session struct {
	JWT     string
	Device  string
	Channel string
}

type PotrfolioCurrencies struct {
	Value []string
}

func (entity *PotrfolioCurrencies) UnmarshalJSON(data []byte) error {
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
