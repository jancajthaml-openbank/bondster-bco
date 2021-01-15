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

package integration

import (
	"encoding/json"
	"strconv"
	"sync"
	"time"

	"github.com/jancajthaml-openbank/bondster-bco-import/model"
	"github.com/jancajthaml-openbank/bondster-bco-import/persistence"
	"github.com/jancajthaml-openbank/bondster-bco-import/support/http"
	"github.com/jancajthaml-openbank/bondster-bco-import/support/timeshift"

	localfs "github.com/jancajthaml-openbank/local-fs"
)

// Workflow represents import integration workflow
type Workflow struct {
	Token            *model.Token
	Tenant           string
	BondsterClient   *http.BondsterClient
	VaultClient      *http.VaultClient
	LedgerClient     *http.LedgerClient
	EncryptedStorage localfs.Storage
	PlaintextStorage localfs.Storage
}

// NewWorkflow returns fascade for integration workflow
func NewWorkflow(
	token *model.Token,
	tenant string,
	bondsterGateway string,
	vaultGateway string,
	ledgerGateway string,
	encryptedStorage localfs.Storage,
	plaintextStorage localfs.Storage,
) Workflow {
	return Workflow{
		Token:            token,
		Tenant:           tenant,
		BondsterClient:   http.NewBondsterClient(bondsterGateway, *token),
		VaultClient:      http.NewVaultClient(vaultGateway),
		LedgerClient:     http.NewLedgerClient(ledgerGateway),
		EncryptedStorage: encryptedStorage,
		PlaintextStorage: plaintextStorage,
	}
}

func importAccountsFromStatemets(
	wg *sync.WaitGroup,
	currency string,
	plaintextStorage localfs.Storage,
	token *model.Token,
	tenant string,
	vaultClient *http.VaultClient,
) {
	defer wg.Done()

	log.Info().Msgf("token %s creating accounts from statements for currency %s", token.ID, currency)

	ids, err := plaintextStorage.ListDirectory("token/"+token.ID+"/statements/"+currency, true)
	if err != nil {
		log.Warn().Msgf("Unable to obtain transaction ids from storage for token %s currency %s", token.ID, currency)
		return
	}

	accounts := make(map[string]bool)
	idsNeedingConfirmation := make([]string, 0)

	for _, id := range ids {
		exists, err := plaintextStorage.Exists("token/" + token.ID + "/statements/" + currency + "/" + id + "/accounts")
		if err != nil {
			log.Warn().Msgf("Unable to check if statement %s/%s/%s accounts exists", token.ID, currency, id)
			continue
		}
		if exists {
			continue
		}

		data, err := plaintextStorage.ReadFileFully("token/" + token.ID + "/statements/" + currency + "/" + id + "/data")
		if err != nil {
			log.Warn().Msgf("Unable to load statement %s/%s/%s", token.ID, currency, id)
			continue
		}

		statement := new(model.BondsterStatement)
		if json.Unmarshal(data, statement) != nil {
			log.Warn().Msgf("Unable to unmarshal statement %s/%s/%s", token.ID, currency, id)
			continue
		}

		accounts[currency+"_TYPE_"+statement.Type] = true
		idsNeedingConfirmation = append(idsNeedingConfirmation, id)
	}

	if len(idsNeedingConfirmation) == 0 {
		return
	}

	accounts[currency+"_TYPE_NOSTRO"] = true

	for account := range accounts {
		log.Debug().Msgf("Creating account %s", account)

		request := model.Account{
			Tenant:         tenant,
			Name:           account,
			Currency:       currency,
			Format:         "BONDSTER_TECHNICAL",
			IsBalanceCheck: false,
		}
		err = vaultClient.CreateAccount(request)
		if err != nil {
			log.Warn().Msgf("Unable to create account %s/%s with %+v", tenant, account, err)
			return
		}
	}

	for _, id := range idsNeedingConfirmation {
		err = plaintextStorage.TouchFile("token/" + token.ID + "/statements/" + currency + "/" + id + "/accounts")
		if err != nil {
			log.Warn().Msgf("Unable to mark account discovery for %s/%s/%s", token.ID, currency, id)
		}
	}

}

func importTransactionsFromStatemets(
	wg *sync.WaitGroup,
	currency string,
	plaintextStorage localfs.Storage,
	token *model.Token,
	tenant string,
	ledgerClient *http.LedgerClient,
) {
	defer wg.Done()

	log.Info().Msgf("token %s creating transactions from statements for currency %s", token.ID, currency)

	ids, err := plaintextStorage.ListDirectory("token/"+token.ID+"/statements/"+currency, true)
	if err != nil {
		log.Warn().Msgf("Unable to obtain transaction ids from storage for token %s currency %s", token.ID, currency)
		return
	}

	for _, id := range ids {
		exists, err := plaintextStorage.Exists("token/" + token.ID + "/statements/" + currency + "/" + id + "/transactions")
		if err != nil {
			log.Warn().Msgf("Unable to check if statement %s/%s/%s transactions exists", token.ID, currency, id)
			continue
		}
		if exists {
			continue
		}

		data, err := plaintextStorage.ReadFileFully("token/" + token.ID + "/statements/" + currency + "/" + id + "/data")
		if err != nil {
			log.Warn().Msgf("Unable to load statement %s/%s/%s", token.ID, currency, id)
			continue
		}

		statement := new(model.BondsterStatement)
		if json.Unmarshal(data, statement) != nil {
			log.Warn().Msgf("Unable to unmarshal statement %s/%s/%s", token.ID, currency, id)
			continue
		}

		credit := model.AccountPair{
			Tenant: tenant,
		}
		debit := model.AccountPair{
			Tenant: tenant,
		}

		if statement.Direction == "CREDIT" {
			credit.Name = statement.Amount.Currency + "_TYPE_NOSTRO"
			debit.Name = statement.Amount.Currency + "_TYPE_" + statement.Type
		} else {
			credit.Name = statement.Amount.Currency + "_TYPE_" + statement.Type
			debit.Name = statement.Amount.Currency + "_TYPE_NOSTRO"
		}

		request := model.Transaction{
			Tenant:        tenant,
			IDTransaction: statement.IDTransfer,
			Transfers: []model.Transfer{
				{
					IDTransfer: statement.IDTransfer,
					Credit:     credit,
					Debit:      debit,
					ValueDate:  statement.ValueDate.Format("2006-01-02T15:04:05Z0700"),
					Amount:     strconv.FormatFloat(statement.Amount.Value, 'f', -1, 64),
					Currency:   statement.Amount.Currency,
				},
			},
		}

		log.Debug().Msgf("Creating transaction %s", statement.IDTransfer)

		err = ledgerClient.CreateTransaction(request)
		if err != nil {
			log.Warn().Msgf("Unable to create transaction %s/%s with %+v", tenant, statement.IDTransfer, err)
			continue
		}

		err = plaintextStorage.TouchFile("token/" + token.ID + "/statements/" + currency + "/" + id + "/transactions")
		if err != nil {
			log.Warn().Msgf("Unable to mark transactions discovery for %s/%s/%s", token.ID, currency, id)
			continue
		}

	}

}

func downloadStatements(
	ids []string,
	currency string,
	plaintextStorage localfs.Storage,
	tokenID string,
	bondsterClient *http.BondsterClient,
) time.Time {
	startTime := time.Date(2017, 1, 1, 0, 0, 0, 0, time.UTC)
	if len(ids) == 0 {
		return startTime
	}
	log.Debug().Msgf("Will synchronize %d statements in %s currency", len(ids), currency)
	statements, err := bondsterClient.GetStatements(currency, ids)
	if err != nil {
		log.Warn().Msgf("Unable to download statements details for currency %s", currency)
		return startTime
	}
	for _, transaction := range statements {
		if transaction.ValueDate.After(startTime) {
			startTime = transaction.ValueDate
		}
		data, err := json.Marshal(transaction)
		if err != nil {
			log.Warn().Msgf("Unable to marshal statement details of %s/%s/%s", tokenID, currency, transaction.IDTransfer)
			continue
		}
		err = plaintextStorage.WriteFileExclusive("token/"+tokenID+"/statements/"+currency+"/"+transaction.IDTransfer+"/data", data)
		if err != nil {
			log.Warn().Msgf("Unable to persist statement details of %s/%s/%s with %+v", tokenID, currency, transaction.IDTransfer, err)
			continue
		}
	}
	return startTime
}

func yieldUnsynchronizedStatementIds(
	batchSize int,
	tokenID string,
	currency string,
	plaintextStorage localfs.Storage,
) <-chan []string {
	chnl := make(chan []string)

	go func() {
		defer close(chnl)
		buffer := make([]string, 0)
		ids, err := plaintextStorage.ListDirectory("token/"+tokenID+"/statements/"+currency, true)
		if err != nil {
			log.Warn().Msgf("Unable to obtain transaction ids from storage for token %s currency %s", tokenID, currency)
			return
		}
		for _, id := range ids {
			exists, err := plaintextStorage.Exists("token/" + tokenID + "/statements/" + currency + "/" + id + "/data")
			if err != nil {
				log.Warn().Msgf("Unable to check if statement %s/%s/%s data exists", tokenID, currency, id)
				continue
			}
			if exists {
				continue
			}
			if len(buffer) == batchSize {
				chunk := make([]string, len(buffer))
				copy(chunk, buffer)
				buffer = make([]string, 0)
				chnl <- chunk
			}
			buffer = append(buffer, id)
		}
		chnl <- buffer
	}()

	return chnl
}

func downloadStatementsForCurrency(
	wg *sync.WaitGroup,
	mutex *sync.RWMutex,
	currency string,
	encryptedStorage localfs.Storage,
	plaintextStorage localfs.Storage,
	token *model.Token,
	bondsterClient *http.BondsterClient,
) {
	defer wg.Done()

	mutex.Lock()
	lastTime, ok := token.LastSyncedFrom[currency]
	mutex.Unlock()
	if !ok {
		log.Warn().Msgf("token %s currency %s unable to obtain last synced time", token.ID, currency)
		return
	}

	// Stage of discovering new transfer ids within given time range
	endTime := time.Now()

	log.Info().Msgf("Token %s discovering new statements for currency %s between %s and %s", token.ID, currency, lastTime.Format("2006-01-02T15:04:05Z0700"), endTime.Format("2006-01-02T15:04:05Z0700"))

	for _, interval := range timeshift.PartitionInterval(lastTime, endTime) {
		ids, err := bondsterClient.GetStatementIdsInInterval(currency, interval)
		if err != nil {
			log.Warn().Msgf("Unable to obtain transaction ids for token %s currency %s", token.ID, currency)
			return
		}

		for _, id := range ids {
			exists, err := plaintextStorage.Exists("token/" + token.ID + "/statements/" + currency + "/" + id)
			if err != nil {
				log.Warn().Msgf("Unable to check if transaction %s exists for token %s currency %s", id, token.ID, currency)
				return
			}
			if exists {
				continue
			}
			err = plaintextStorage.TouchFile("token/" + token.ID + "/statements/" + currency + "/" + id + "/mark")
			if err != nil {
				log.Warn().Msgf("Unable to mark transaction %s as known for token %s currency %s", id, token.ID, currency)
				return
			}
		}
	}

	// Stage when ensuring that all transfer ids in given time range have downloaded statements
	for ids := range yieldUnsynchronizedStatementIds(100, token.ID, currency, plaintextStorage) {
		newTime := downloadStatements(
			ids,
			currency,
			plaintextStorage,
			token.ID,
			bondsterClient,
		)
		if newTime.Before(lastTime) {
			continue
		}
		lastTime = newTime
		mutex.Lock()
		token.LastSyncedFrom[currency] = lastTime
		mutex.Unlock()
		if !persistence.UpdateToken(encryptedStorage, token) {
			log.Warn().Msgf("unable to update token %s", token.ID)
		}
	}

}

func (workflow Workflow) SynchronizeCurrencies() {
	log.Debug().Msgf("token %s synchronizing currencies from Bondster gateway", workflow.Token.ID)

	currencies, err := workflow.BondsterClient.GetCurrencies()
	if err != nil {
		log.Warn().Msgf("token %s Unable to get currencies because %+v", workflow.Token.ID, err)
		return
	}

	for _, currency := range currencies {
		if _, ok := workflow.Token.LastSyncedFrom[currency]; !ok {
			workflow.Token.LastSyncedFrom[currency] = time.Date(2017, 1, 1, 0, 0, 0, 0, time.UTC)
			if !persistence.UpdateToken(workflow.EncryptedStorage, workflow.Token) {
				log.Warn().Msgf("unable to update token %s", workflow.Token.ID)
				continue
			}
		}
	}

	return
}

func (workflow Workflow) SynchronizeStatements() {
	var wg sync.WaitGroup
	wg.Add(len(workflow.Token.LastSyncedFrom))
	mutex := sync.RWMutex{}
	for currency := range workflow.Token.LastSyncedFrom {
		go downloadStatementsForCurrency(
			&wg,
			&mutex,
			currency,
			workflow.EncryptedStorage,
			workflow.PlaintextStorage,
			workflow.Token,
			workflow.BondsterClient,
		)
	}
	wg.Wait()
}

func (workflow Workflow) CreateAccounts() {
	log.Debug().Msgf("token %s ensuring accounts based on statements", workflow.Token.ID)

	var wg sync.WaitGroup
	wg.Add(len(workflow.Token.LastSyncedFrom))
	for currency := range workflow.Token.LastSyncedFrom {
		go importAccountsFromStatemets(
			&wg,
			currency,
			workflow.PlaintextStorage,
			workflow.Token,
			workflow.Tenant,
			workflow.VaultClient,
		)
	}
	wg.Wait()
}

func (workflow Workflow) CreateTransactions() {
	log.Debug().Msgf("token %s creating transactions based on statements", workflow.Token.ID)

	var wg sync.WaitGroup
	wg.Add(len(workflow.Token.LastSyncedFrom))
	for currency := range workflow.Token.LastSyncedFrom {
		go importTransactionsFromStatemets(
			&wg,
			currency,
			workflow.PlaintextStorage,
			workflow.Token,
			workflow.Tenant,
			workflow.LedgerClient,
		)
	}
	wg.Wait()
}
