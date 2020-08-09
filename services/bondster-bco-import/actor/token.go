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
	"sort"
	"time"

	"github.com/jancajthaml-openbank/bondster-bco-import/bondster"
	"github.com/jancajthaml-openbank/bondster-bco-import/ledger"
	"github.com/jancajthaml-openbank/bondster-bco-import/metrics"
	"github.com/jancajthaml-openbank/bondster-bco-import/model"
	"github.com/jancajthaml-openbank/bondster-bco-import/persistence"
	"github.com/jancajthaml-openbank/bondster-bco-import/utils"
	"github.com/jancajthaml-openbank/bondster-bco-import/vault"

	system "github.com/jancajthaml-openbank/actor-system"
	localfs "github.com/jancajthaml-openbank/local-fs"
)

// NilToken represents token that is neither existing neither non existing
func NilToken(s *ActorSystem) func(interface{}, system.Context) {
	return func(t_state interface{}, context system.Context) {
		state := t_state.(model.Token)

		tokenHydration := persistence.LoadToken(s.Storage, state.ID)

		if tokenHydration == nil {
			context.Self.Become(state, NonExistToken(s))
			log.WithField("token", state.ID).Debug("Nil -> NonExist")
		} else {
			context.Self.Become(*tokenHydration, ExistToken(s))
			log.WithField("token", state.ID).Debug("Nil -> Exist")
		}

		context.Self.Receive(context)
	}
}

// NonExistToken represents token that does not exist
func NonExistToken(s *ActorSystem) func(interface{}, system.Context) {
	return func(t_state interface{}, context system.Context) {
		state := t_state.(model.Token)

		switch msg := context.Data.(type) {

		case model.ProbeMessage:
			break

		case model.CreateToken:
			tokenResult := persistence.CreateToken(s.Storage, state.ID, msg.Username, msg.Password)

			if tokenResult == nil {
				s.SendMessage(FatalError, context.Sender, context.Receiver)
				log.WithField("token", state.ID).Debug("(NonExist CreateToken) Error")
				return
			}

			s.SendMessage(RespCreateToken, context.Sender, context.Receiver)
			log.WithField("token", state.ID).Info("New Token Created")
			log.WithField("token", state.ID).Debug("(NonExist CreateToken) OK")
			s.Metrics.TokenCreated()

			context.Self.Become(*tokenResult, ExistToken(s))
			context.Self.Tell(model.SynchronizeToken{}, context.Receiver, context.Sender)

		case model.DeleteToken:
			s.SendMessage(FatalError, context.Sender, context.Receiver)
			log.WithField("token", state.ID).Debug("(NonExist DeleteToken) Error")

		case model.SynchronizeToken:
			break

		default:
			s.SendMessage(FatalError, context.Sender, context.Receiver)
			log.WithField("token", state.ID).Debug("(NonExist Unknown Message) Error")
		}

		return
	}
}

// ExistToken represents account that does exist
func ExistToken(s *ActorSystem) func(interface{}, system.Context) {
	return func(t_state interface{}, context system.Context) {
		state := t_state.(model.Token)

		switch context.Data.(type) {

		case model.ProbeMessage:
			break

		case model.CreateToken:
			s.SendMessage(FatalError, context.Sender, context.Receiver)
			log.WithField("token", state.ID).Debug("(Exist CreateToken) Error")

		case model.SynchronizeToken:
			log.WithField("token", state.ID).Debug("(Exist SynchronizeToken)")
			context.Self.Become(t_state, SynchronizingToken(s))
			go importStatements(s, state, func() {
				context.Self.Become(t_state, NilToken(s))
				context.Self.Tell(model.ProbeMessage{}, context.Receiver, context.Receiver)
			})

		case model.DeleteToken:
			if !persistence.DeleteToken(s.Storage, state.ID) {
				s.SendMessage(FatalError, context.Sender, context.Receiver)
				log.WithField("token", state.ID).Debug("(Exist DeleteToken) Error")
				return
			}
			log.WithField("token", state.ID).Info("Token Deleted")
			log.WithField("token", state.ID).Debug("(Exist DeleteToken) OK")
			s.Metrics.TokenDeleted()
			s.SendMessage(RespDeleteToken, context.Sender, context.Receiver)
			context.Self.Become(state, NonExistToken(s))

		default:
			s.SendMessage(FatalError, context.Sender, context.Receiver)
			log.WithField("token", state.ID).Warn("(Exist Unknown Message) Error")

		}

		return
	}
}

// SynchronizingToken represents account that is currently synchronizing
func SynchronizingToken(s *ActorSystem) func(interface{}, system.Context) {
	return func(t_state interface{}, context system.Context) {
		state := t_state.(model.Token)

		switch context.Data.(type) {

		case model.ProbeMessage:
			break

		case model.CreateToken:
			s.SendMessage(FatalError, context.Sender, context.Receiver)
			log.WithField("token", state.ID).Debug("(Synchronizing CreateToken) Error")

		case model.SynchronizeToken:
			log.WithField("token", state.ID).Debug("(Synchronizing SynchronizeToken)")

		case model.DeleteToken:
			s.SendMessage(FatalError, context.Sender, context.Receiver)
			log.WithField("token", state.ID).Debug("(Synchronizing DeleteToken) Error")

		default:
			s.SendMessage(FatalError, context.Sender, context.Receiver)
			log.WithField("token", state.ID).Warn("(Synchronizing Unknown Message) Error")

		}

		return
	}
}

func importStatementsForInterval(tenant string, bondsterClient *bondster.BondsterClient, vaultClient *vault.VaultClient, ledgerClient *ledger.LedgerClient, storage *localfs.EncryptedStorage, metrics *metrics.Metrics, token *model.Token, currency string, interval utils.TimeRange) (time.Time, error) {
	log.Debugf("Importing bondster statements for interval %s", interval.String())

	var (
		err            error
		transactionIds []string
		statements     *bondster.BondsterImportEnvelope
		lastSynced     time.Time = token.LastSyncedFrom[currency]
	)

	metrics.TimeTransactionSearchLatency(func() {
		transactionIds, err = bondsterClient.GetTransactionIdsInInterval(currency, interval)
	})
	if err != nil {
		return lastSynced, err
	}

	if len(transactionIds) == 0 {
		return interval.EndTime, nil
	}

	metrics.TimeTransactionListLatency(func() {
		statements, err = bondsterClient.GetTransactionDetails(currency, transactionIds)
	})
	if err != nil {
		return lastSynced, err
	}
	if len(statements.Transactions) == 0 {
		return lastSynced, nil
	}

	// FIXME getStatements end here

	log.WithField("token", token.ID).Debugf("sorting statements")

	sort.SliceStable(statements.Transactions, func(i, j int) bool {
		return statements.Transactions[i].IDTransaction == statements.Transactions[j].IDTransaction
	})

	log.WithField("token", token.ID).Debugf("importing accounts")

	for account := range statements.GetAccounts() {
		err = vaultClient.CreateAccount(tenant, account)
		if err != nil {
			return lastSynced, err
		}
	}

	log.WithField("token", token.ID).Debugf("importing transactions")

	for transaction := range statements.GetTransactions(tenant) {
		err = ledgerClient.CreateTransaction(tenant, transaction)
		if err != nil {
			return lastSynced, err
		}

		metrics.TransactionImported()
		metrics.TransfersImported(int64(len(transaction.Transfers)))

		for _, transfer := range transaction.Transfers {
			if transfer.ValueDateRaw.After(lastSynced) {
				lastSynced = transfer.ValueDateRaw
			}
		}
	}

	return lastSynced, nil
}

func importNewStatements(tenant string, bondsterClient *bondster.BondsterClient, vaultClient *vault.VaultClient, ledgerClient *ledger.LedgerClient, storage *localfs.EncryptedStorage, metrics *metrics.Metrics, token *model.Token, currency string) error {
	for _, interval := range utils.PartitionInterval(token.LastSyncedFrom[currency], time.Now()) {
		lastSynced, err := importStatementsForInterval(tenant, bondsterClient, vaultClient, ledgerClient, storage, metrics, token, currency, interval)
		if lastSynced.After(token.LastSyncedFrom[currency]) {
			token.LastSyncedFrom[currency] = lastSynced
			if !persistence.UpdateToken(storage, token) {
				log.WithField("token", token.ID).Warn("Unable to update token")
			}
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func importStatements(s *ActorSystem, token model.Token, callback func()) {
	defer callback()

	log.WithField("token", token.ID).Debugf("Importing statements")

	bondsterClient := bondster.NewBondsterClient(s.BondsterGateway, token)
	vaultClient := vault.NewVaultClient(s.VaultGateway)
	ledgerClient := ledger.NewLedgerClient(s.LedgerGateway)

	// FIXME lines get + update correncies should be a bondster method
	currencies, err := bondsterClient.GetCurrencies()
	if err != nil {
		log.WithField("token", token.ID).Warnf("Unable to get currencies because %+v", err)
		return
	}

	if token.UpdateCurrencies(currencies) && !persistence.UpdateToken(s.Storage, &token) {
		log.WithField("token", token.ID).Warnf("Update of token currencies has failed, currencies: %s", currencies)
		return
	}

	for currency := range token.LastSyncedFrom {
		log.WithField("token", token.ID).Debugf("Import for currency %s Begin", currency)
		err := importNewStatements(s.Tenant, &bondsterClient, &vaultClient, &ledgerClient, s.Storage, s.Metrics, &token, currency)
		if err != nil {
			log.WithField("token", token.ID).Errorf("Import statements failed with %+v", err)
		}
		log.WithField("token", token.ID).Debugf("Import for currency %s End", currency)
	}
}
