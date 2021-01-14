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

package actor

import (
	//"fmt"
	"time"
	"sync"

	//"github.com/jancajthaml-openbank/bondster-bco-import/metrics"
	"github.com/jancajthaml-openbank/bondster-bco-import/model"
	"github.com/jancajthaml-openbank/bondster-bco-import/persistence"
	"github.com/jancajthaml-openbank/bondster-bco-import/support/http"
	"github.com/jancajthaml-openbank/bondster-bco-import/support/timeshift"

	system "github.com/jancajthaml-openbank/actor-system"
	localfs "github.com/jancajthaml-openbank/local-fs"
)

// NilToken represents token that is neither existing neither non existing
func NilToken(s *System) func(interface{}, system.Context) {
	return func(t_state interface{}, context system.Context) {
		state := t_state.(model.Token)

		tokenHydration := persistence.LoadToken(s.Storage, state.ID)

		if tokenHydration == nil {
			context.Self.Become(state, NonExistToken(s))
			log.Debug().Msgf("token %s Nil -> NonExist", state.ID)
		} else {
			context.Self.Become(*tokenHydration, ExistToken(s))
			log.Debug().Msgf("token %s Nil -> Exist", state.ID)
		}

		context.Self.Receive(context)
	}
}

// NonExistToken represents token that does not exist
func NonExistToken(s *System) func(interface{}, system.Context) {
	return func(t_state interface{}, context system.Context) {
		state := t_state.(model.Token)

		switch msg := context.Data.(type) {

		case ProbeMessage:
			break

		case CreateToken:
			tokenResult := persistence.CreateToken(s.Storage, state.ID, msg.Username, msg.Password)

			if tokenResult == nil {
				s.SendMessage(FatalError, context.Sender, context.Receiver)
				log.Debug().Msgf("token %s (NonExist CreateToken) Error", state.ID)
				return
			}

			s.SendMessage(RespCreateToken, context.Sender, context.Receiver)
			log.Info().Msgf("New Token %s Created", state.ID)
			log.Debug().Msgf("token %s (NonExist CreateToken) OK", state.ID)
			s.Metrics.TokenCreated()

			context.Self.Become(*tokenResult, ExistToken(s))
			context.Self.Tell(SynchronizeToken{}, context.Receiver, context.Sender)

		case DeleteToken:
			s.SendMessage(FatalError, context.Sender, context.Receiver)
			log.Debug().Msgf("token %s (NonExist DeleteToken) Error", state.ID)

		case SynchronizeToken:
			break

		default:
			s.SendMessage(FatalError, context.Sender, context.Receiver)
			log.Debug().Msgf("token %s (NonExist Unknown Message) Error", state.ID)
		}

		return
	}
}

// ExistToken represents account that does exist
func ExistToken(s *System) func(interface{}, system.Context) {
	return func(t_state interface{}, context system.Context) {
		state := t_state.(model.Token)

		switch context.Data.(type) {

		case ProbeMessage:
			break

		case CreateToken:
			s.SendMessage(FatalError, context.Sender, context.Receiver)
			log.Debug().Msgf("token %s (Exist CreateToken) Error", state.ID)

		case SynchronizeToken:
			log.Debug().Msgf("token %s (Exist SynchronizeToken)", state.ID)
			context.Self.Become(t_state, SynchronizingToken(s))
			go importStatements(s, state, func() {
				context.Self.Become(t_state, NilToken(s))
				context.Self.Tell(ProbeMessage{}, context.Receiver, context.Receiver)
			})

		case DeleteToken:
			if !persistence.DeleteToken(s.Storage, state.ID) {
				s.SendMessage(FatalError, context.Sender, context.Receiver)
				log.Debug().Msgf("token %s (Exist DeleteToken) Error", state.ID)
				return
			}
			log.Info().Msgf("Token %s Deleted", state.ID)
			log.Debug().Msgf("token %s (Exist DeleteToken) OK", state.ID)
			s.Metrics.TokenDeleted()
			s.SendMessage(RespDeleteToken, context.Sender, context.Receiver)
			context.Self.Become(state, NonExistToken(s))

		default:
			s.SendMessage(FatalError, context.Sender, context.Receiver)
			log.Debug().Msgf("token %s (Exist Unknown Message) Error", state.ID)

		}

		return
	}
}

// SynchronizingToken represents account that is currently synchronizing
func SynchronizingToken(s *System) func(interface{}, system.Context) {
	return func(t_state interface{}, context system.Context) {
		state := t_state.(model.Token)

		switch context.Data.(type) {

		case ProbeMessage:
			break

		case CreateToken:
			s.SendMessage(FatalError, context.Sender, context.Receiver)
			log.Debug().Msgf("token %s (Synchronizing CreateToken) Error", state.ID)

		case SynchronizeToken:
			log.Debug().Msgf("token %s (Synchronizing SynchronizeToken)", state.ID)

		case DeleteToken:
			if !persistence.DeleteToken(s.Storage, state.ID) {
				s.SendMessage(FatalError, context.Sender, context.Receiver)
				log.Debug().Msgf("token %s (Synchronizing DeleteToken) Error", state.ID)
				return
			}
			log.Info().Msgf("Token %s Deleted", state.ID)
			log.Debug().Msgf("token %s (Synchronizing DeleteToken) OK", state.ID)
			s.Metrics.TokenDeleted()
			s.SendMessage(RespDeleteToken, context.Sender, context.Receiver)
			context.Self.Become(state, NonExistToken(s))

		default:
			s.SendMessage(FatalError, context.Sender, context.Receiver)
			log.Debug().Msgf("token %s (Synchronizing Unknown Message) Error", state.ID)
		}

		return
	}
}



/*
func importStatements(tenant string, bondsterClient *http.BondsterClient, vaultClient *http.VaultClient, ledgerClient *http.LedgerClient, storage localfs.Storage, metrics metrics.Metrics, token *model.Token, currency string, interval timeshift.TimeRange) (time.Time, error) {
	log.Debug().Msgf("Importing bondster statements for currency %s and interval %d/%d - %d/%d", currency, interval.StartTime.Month(), interval.StartTime.Year(), interval.EndTime.Month(), interval.EndTime.Year())

	var err error
	var transactionIds []string
	var statements *model.ImportEnvelope

	lastSynced := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)

	transactionIds, err = bondsterClient.GetTransactionIdsInInterval(currency, interval)
	if err != nil {
		log.Warn().Msgf("token %s failed to obtain statements for this period", token.ID)
		return lastSynced, err
	}

	if len(transactionIds) == 0 {
		log.Info().Msgf("token %s no statements in this period", token.ID)
		return interval.EndTime, nil
	}

	statements, err = bondsterClient.GetTransactionDetails(currency, transactionIds)
	if err != nil {
		return lastSynced, err
	}

	log.Debug().Msgf("token %s importing accounts", token.ID)

	var accountsStageError error

	for account := range statements.GetAccounts(tenant) {
		if accountsStageError != nil {
			continue
		}
		log.Debug().Msgf("token %s importing account %+v", token.ID, account)
		accountsStageError = vaultClient.CreateAccount(account)
	}

	if accountsStageError != nil {
		log.Debug().Msgf("token %s importing accounts failed with %+v", token.ID, accountsStageError)
		return lastSynced, accountsStageError
	}

	log.Debug().Msgf("token %s importing transactions", token.ID)

	var transactionStageError error

	for transaction := range statements.GetTransactions(tenant) {
		if transactionStageError != nil {
			continue
		}
		log.Debug().Msgf("token %s importing transaction %+v", token.ID, transaction)
		transactionStageError = ledgerClient.CreateTransaction(transaction)
		if transactionStageError == nil {
			metrics.TransactionImported(len(transaction.Transfers))
			for _, transfer := range transaction.Transfers {
				if transfer.ValueDateRaw.After(lastSynced) {
					lastSynced = transfer.ValueDateRaw
				}
			}
		}
	}

	if transactionStageError != nil {
		log.Debug().Msgf("token %s importing transfers failed with %+v", token.ID, transactionStageError)
		return lastSynced, transactionStageError
	}

	return lastSynced, nil
}*/

/*
func importNewStatements(tenant string, bondsterClient *http.BondsterClient, vaultClient *http.VaultClient, ledgerClient *http.LedgerClient, storage localfs.Storage, metrics metrics.Metrics, token *model.Token, currency string) (bool, error) {
	startTime, ok := token.LastSyncedFrom[currency]
	if !ok {
		startTime = time.Date(2017, 1, 1, 0, 0, 0, 0, time.UTC)
		token.LastSyncedFrom[currency] = startTime
	}

	for _, interval := range timeshift.PartitionInterval(startTime, time.Now()) {
		transactions, err := getTransactionIdsInInterval(
			tenant,
			bondsterClient,
			storage,
			token,
			currency,
			interval,
		)

		if err != nil {
			log.Warn().Msgf("Unable to obtain transactions with error %+v", err)
			continue
		}

		lastSynced := interval.EndTime

		log.Debug().Msgf("Transactions %+v", transactions)
		log.Debug().Msgf("Setting currency %s lastSynced to %+v", transactions, lastSynced)

		if lastSynced.After(token.LastSyncedFrom[currency]) {
			log.Debug().Msgf("token %s setting last synced for currency %s to %s", token.ID, currency, lastSynced.Format(time.RFC3339))
			token.LastSyncedFrom[currency] = lastSynced
			if !persistence.UpdateToken(storage, token) {
				err = fmt.Errorf("unable to update token")
			}
			return false, err
		} else if err != nil {
			return false, err
		}
	}
	return true, nil
}
*/

func importStatementsForCurrency(
	wg *sync.WaitGroup,
	currency string,
	storage localfs.Storage,
	token *model.Token,
	bondsterClient *http.BondsterClient,
) {
	defer func() {
		recover()
		log.Debug().Msgf("Import for token %s and currency %s finished", token.ID, currency)
		wg.Done()
	}()

	startTime, ok := token.LastSyncedFrom[currency]
	if !ok {
		startTime = time.Date(2017, 1, 1, 0, 0, 0, 0, time.UTC)
		token.LastSyncedFrom[currency] = startTime
		if !persistence.UpdateToken(storage, token) {
			log.Warn().Msgf("unable to update token %s", token.ID)
			return
		}
	}

	for _, interval := range timeshift.PartitionInterval(startTime, time.Now()) {
		log.Debug().Msgf("Start Importing statements for token %s currency %s and interval %d/%d -> %d/%d", token.ID, currency, interval.StartTime.Month(), interval.StartTime.Year(), interval.EndTime.Month(), interval.EndTime.Year())

		ids, err := bondsterClient.GetTransactionIdsInInterval(currency, interval)
		if err != nil {
			log.Warn().Msgf("Unable to obtain transaction ids for token %s currency %s and interval %d/%d -> %d/%d", token.ID, currency, interval.StartTime.Month(), interval.StartTime.Year(), interval.EndTime.Month(), interval.EndTime.Year())
			return
		}

		for _, id := range ids {
			exists, err := storage.Exists("token/" + token.ID + "/statements/" + id)

			if err != nil {
				log.Warn().Msgf("Unable to check if transaction %s exists for token %s currency %s and interval %d/%d -> %d/%d", id, token.ID, currency, interval.StartTime.Month(), interval.StartTime.Year(), interval.EndTime.Month(), interval.EndTime.Year())
				return
			}

			if exists {
				continue
			}

			err = storage.TouchFile("token/" + token.ID + "/statements/" + id)
			if err != nil {
				log.Warn().Msgf("Unable to mark transaction %s as known for token %s currency %s and interval %d/%d -> %d/%d", id, token.ID, currency, interval.StartTime.Month(), interval.StartTime.Year(), interval.EndTime.Month(), interval.EndTime.Year())
				return
			}

			//log.Info().Msgf("Token New %s transaction %s (import)", token.ID, id)
		}

		log.Debug().Msgf("End Importing statements for token %s currency %s and interval %d/%d -> %d/%d", token.ID, currency, interval.StartTime.Month(), interval.StartTime.Year(), interval.EndTime.Month(), interval.EndTime.Year())
	}
}

func importStatements(s *System, token model.Token, complete func()) {
	defer complete()

	log.Debug().Msgf("token %s Importing statements Start", token.ID)

	bondsterClient := http.NewBondsterClient(s.BondsterGateway, token)
	//vaultClient := http.NewVaultClient(s.VaultGateway)
	//ledgerClient := http.NewLedgerClient(s.LedgerGateway)

	currencies, err := bondsterClient.GetCurrencies()
	if err != nil {
		log.Warn().Msgf("token %s Unable to get currencies because %+v", token.ID, err)
		return
	}

	var wg sync.WaitGroup
	wg.Add(len(currencies))
	for _, currency := range currencies {
		go importStatementsForCurrency(
			&wg,
			currency,
			s.Storage,
			&token,
			&bondsterClient,
		)
	}
	wg.Wait()

	log.Debug().Msgf("token %s Importing statements End", token.ID)
}
