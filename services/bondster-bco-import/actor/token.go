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
	"encoding/json"
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

		tokenHydration := persistence.LoadToken(s.EncryptedStorage, state.ID)

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
			tokenResult := persistence.CreateToken(s.EncryptedStorage, state.ID, msg.Username, msg.Password)

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
			if !persistence.DeleteToken(s.EncryptedStorage, state.ID) {
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
			if !persistence.DeleteToken(s.EncryptedStorage, state.ID) {
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

func importAccountsFromStatemets(
	wg *sync.WaitGroup,
	currency string,
	plaintextStorage localfs.Storage,
	token *model.Token,
	vaultClient *http.VaultClient,
) {
	defer func() {
		recover()
		wg.Done()
	}()

	log.Info().Msgf("Token %s creating accounts from statements for currency %s", token.ID, currency)

	ids, err := plaintextStorage.ListDirectory("token/" + token.ID + "/statements/" + currency, true)
	if err != nil {
		log.Warn().Msgf("Unable to obtain transaction ids from storage for token %s currency %s", token.ID, currency)
		return
	}

	accounts := make(map[string]bool)
	accounts[currency + "_TYPE_NOSTRO"] = true

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

		accounts[currency + "_TYPE_" + statement.Type] = true
	}

	for account := range accounts {
		log.Debug().Msgf("Token %s creating account %s", token.ID, account)
	}

}

func importStatementsForCurrency(
	wg *sync.WaitGroup,
	mutex *sync.RWMutex,
	currency string,
	encryptedStorage localfs.Storage,
	plaintextStorage localfs.Storage,
	token *model.Token,
	bondsterClient *http.BondsterClient,
) {
	defer func() {
		recover()
		wg.Done()
	}()

	mutex.Lock()
	startTime, ok := token.LastSyncedFrom[currency]
	mutex.Unlock()

	if !ok {
		startTime = time.Date(2017, 1, 1, 0, 0, 0, 0, time.UTC)
		mutex.Lock()
		token.LastSyncedFrom[currency] = startTime
		mutex.Unlock()
		if !persistence.UpdateToken(encryptedStorage, token) {
			log.Warn().Msgf("unable to update token %s", token.ID)
			return
		}
	}

	// Stage of discovering new transfer ids within given time range

	endTime := time.Now()

	log.Info().Msgf("Token %s discovering new statements for currency %s between %d/%d and %d/%d", token.ID, currency, startTime.Month(), startTime.Year(), endTime.Month(), endTime.Year())

	for _, interval := range timeshift.PartitionInterval(startTime, endTime) {
		ids, err := bondsterClient.GetTransferIdsInInterval(currency, interval)
		if err != nil {
			log.Warn().Msgf("Unable to obtain transaction ids for token %s currency %s and interval %d/%d -> %d/%d", token.ID, currency, interval.StartTime.Month(), interval.StartTime.Year(), interval.EndTime.Month(), interval.EndTime.Year())
			return
		}

		for _, id := range ids {
			exists, err := plaintextStorage.Exists("token/" + token.ID + "/statements/" + currency + "/" + id)
			if err != nil {
				log.Warn().Msgf("Unable to check if transaction %s exists for token %s currency %s and interval %d/%d -> %d/%d", id, token.ID, currency, interval.StartTime.Month(), interval.StartTime.Year(), interval.EndTime.Month(), interval.EndTime.Year())
				return
			}
			if exists {
				continue
			}
			err = plaintextStorage.WriteFileExclusive("token/" + token.ID + "/statements/" + currency + "/" + id + "/mark", []byte(interval.EndTime.Format("2006-01-02T15:04:05Z0700")))
			if err != nil {
				log.Warn().Msgf("Unable to mark transaction %s as known for token %s currency %s and interval %d/%d -> %d/%d", id, token.ID, currency, interval.StartTime.Month(), interval.StartTime.Year(), interval.EndTime.Month(), interval.EndTime.Year())
				return
			}
		}
	}

	// FIXME to separate method

	// Stage when ensuring that all transfer ids in given time range have downloaded statements

	log.Info().Msgf("Token %s synchronizing statements from gateway for currency %s", token.ID, currency)

	ids, err := plaintextStorage.ListDirectory("token/" + token.ID + "/statements/" + currency, true)
	if err != nil {
		log.Warn().Msgf("Unable to obtain transaction ids from storage for token %s currency %s", token.ID, currency)
		return
	}

	unsynchronized := make([]string, 0)
	for _, id := range ids {
		exists, err := plaintextStorage.Exists("token/" + token.ID + "/statements/" + currency + "/" + id + "/data")
		if err != nil {
			log.Warn().Msgf("Unable to check if statement %s/%s/%s data exists", token.ID, currency, id)
			continue
		}
		if exists {
			continue
		}
		// FIXME in-place if reached 100 synchronize, don't load all in memory
		unsynchronized = append(unsynchronized, id)
	}

	batchSize := 100
	batches := make([][]string, 0, (len(unsynchronized) + batchSize - 1) / batchSize)

	for batchSize < len(unsynchronized) {
	  unsynchronized, batches = unsynchronized[batchSize:], append(batches, unsynchronized[0:batchSize:batchSize])
	}

	batches = append(batches, unsynchronized)

	for _, ids := range batches {
		//log.Debug().Msgf("Following stamenents needs to be downloaded from gateway %+v", ids)

		envelope, err := bondsterClient.GetTransactionDetails(currency, ids)
		if err != nil {
			log.Warn().Msgf("Unable to download statements details for currency %s", currency)
			return
		}

		for _, transaction := range envelope.Transactions {
			if transaction.ValueDate.After(startTime) {
				startTime = transaction.ValueDate
			}
		}

		mutex.Lock()
		token.LastSyncedFrom[currency] = startTime
		mutex.Unlock()

		log.Debug().Msgf("Updating last synchronized time for token %s and currency %s to %s", token.ID, currency, startTime.Format(time.RFC3339))

		if !persistence.UpdateToken(encryptedStorage, token) {
			log.Warn().Msgf("unable to update token %s", token.ID)
		}

		log.Debug().Msgf("Downloading new statements for token %s and currency %s", token.ID, currency)

		for _, transaction := range envelope.Transactions {
			data, err := json.Marshal(transaction)
			if err != nil {
				log.Warn().Msgf("Unable to marshal statement details of %s/%s/%s", token.ID, currency, transaction.IDTransfer)
				continue
			}
			err = plaintextStorage.WriteFileExclusive("token/" + token.ID + "/statements/" + currency + "/" + transaction.IDTransfer + "/data", data)
			if err != nil {
				log.Warn().Msgf("Unable to persist statement details of %s/%s/%s with %+v", token.ID, currency, transaction.IDTransfer, err)
				continue
			}

		}

	}

}

func importStatements(s *System, token model.Token, complete func()) {
	defer complete()

	log.Debug().Msgf("token %s Importing statements Start", token.ID)

	bondsterClient := http.NewBondsterClient(s.BondsterGateway, token)
	vaultClient := http.NewVaultClient(s.VaultGateway)
	//ledgerClient := http.NewLedgerClient(s.LedgerGateway)

	currencies, err := bondsterClient.GetCurrencies()
	if err != nil {
		log.Warn().Msgf("token %s Unable to get currencies because %+v", token.ID, err)
		return
	}

	// Bondster stage
	var wg sync.WaitGroup
	wg.Add(len(currencies))
	mutex := sync.RWMutex{}

	for _, currency := range currencies {
		go importStatementsForCurrency(
			&wg,
			&mutex,
			currency,
			s.EncryptedStorage,
			s.PlaintextStorage,
			&token,
			&bondsterClient,
		)
	}
	wg.Wait()

	// Vault stage
	wg.Add(len(currencies))

	for _, currency := range currencies {
		go importAccountsFromStatemets(
			&wg,
			currency,
			s.PlaintextStorage,
			&token,
			&vaultClient,
		)
	}
	wg.Wait()

	// Ledger stage
	// FIXME TBD

	log.Debug().Msgf("token %s Importing statements End", token.ID)
}
