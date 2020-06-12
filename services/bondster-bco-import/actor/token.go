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
	"fmt"
	"time"

	"github.com/jancajthaml-openbank/bondster-bco-import/integration"
	"github.com/jancajthaml-openbank/bondster-bco-import/model"
	"github.com/jancajthaml-openbank/bondster-bco-import/metrics"
	"github.com/jancajthaml-openbank/bondster-bco-import/persistence"
	"github.com/jancajthaml-openbank/bondster-bco-import/utils"

	localfs "github.com/jancajthaml-openbank/local-fs"
	system "github.com/jancajthaml-openbank/actor-system"
)

// NilToken represents token that is neither existing neither non existing
func NilToken(s *ActorSystem) func(interface{}, system.Context) {
	return func(t_state interface{}, context system.Context) {
		state := t_state.(model.Token)

		tokenHydration := persistence.LoadToken(s.Storage, state.ID)

		if tokenHydration == nil {
			context.Self.Become(state, NonExistToken(s))
			log.WithField("token", state.ID).Debugf("Nil -> NonExist")
		} else {
			context.Self.Become(*tokenHydration, ExistToken(s))
			log.WithField("token", state.ID).Debugf("Nil -> Exist")
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
				log.WithField("token", state.ID).Debugf("(NonExist CreateToken) Error")
				return
			}

			s.SendMessage(RespCreateToken, context.Sender, context.Receiver)
			log.WithField("token", state.ID).Infof("New Token Created")
			log.WithField("token", state.ID).Debugf("(NonExist CreateToken) OK")
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
			log.WithField("token", state.ID).Debugf("(NonExist Unknown Message) Error")
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

func importStatementsForInterval(bondsterGateway string, vaultGateway string, ledgerGateway string, tenant string, httpClient integration.Client, storage *localfs.EncryptedStorage, metrics *metrics.Metrics, token *model.Token, currency string, session *model.Session, interval utils.TimeRange) error {
	log.Debugf("Importing bondster statements for interval %s", interval.String())

	var (
		err      error
		transactionIds []string
		statements *model.BondsterImportEnvelope
		response integration.Response
		request  []byte
	)

	metrics.TimeTransactionSearchLatency(func() {
		transactionIds, err = integration.GetTransactionIdsInInterval(httpClient, bondsterGateway, session, currency, interval)
	})
	if err != nil {
		return err
	}

	log.WithField("token", token.ID).Debugf("found %d transactions", len(transactionIds))

	if len(transactionIds) == 0 {
		if interval.EndTime.After(token.LastSyncedFrom[currency]) {
			token.LastSyncedFrom[currency] = interval.EndTime
			if !persistence.UpdateToken(storage, token) {
				log.WithField("token", token.ID).Warn("Unable to update token last synced")
			}
		}
		return nil
	}

	metrics.TimeTransactionListLatency(func() {
		statements, err = integration.GetTransactionDetails(httpClient, bondsterGateway, session, currency, transactionIds)
	})
	if err != nil {
		return err
	}

	// FIXME getStatements end here

	accounts := statements.GetAccounts()
	log.WithField("token", token.ID).Debugf("importing %d accounts", len(accounts))

	for _, account := range accounts {
		request, err = utils.JSON.Marshal(account)
		if err != nil {
			return err
		}

		uri := vaultGateway + "/account/" + tenant
		response, err = httpClient.Post(uri, request, nil)
		if err != nil {
			return fmt.Errorf("vault-rest create account %s error %+v", uri, err)
		}
		if response.Status == 400 {
			return fmt.Errorf("vault-rest account malformed request %+v", string(request))
		}
		if response.Status == 504 {
			return fmt.Errorf("vault-rest create account timeout")
		}
		if response.Status != 200 && response.Status != 409 {
			return fmt.Errorf("vault-rest create account %s error %+v", uri, response)
		}
	}

	var lastSynced time.Time = token.LastSyncedFrom[currency]

	transactions := statements.GetTransactions(tenant)
	log.WithField("token", token.ID).Debugf("importing %d transactions", len(transactions))

	for _, transaction := range transactions {
		for {
			request, err = utils.JSON.Marshal(transaction)
			if err != nil {
				return err
			}
			uri := ledgerGateway + "/transaction/" + tenant
			response, err = httpClient.Post(uri, request, nil)
			if err != nil {
				return fmt.Errorf("ledger-rest create transaction %s error %+v", uri, err)
			}
			if response.Status == 409 {
				// FIXME in future, follback original transaction and create new based on
				// union of existing transaction and new (needs persistence)
				transaction.IDTransaction = transaction.IDTransaction + "_"
				continue
			}
			if response.Status == 400 {
				return fmt.Errorf("ledger-rest transaction malformed request %+v", string(request))
			}
			if response.Status == 504 {
				return fmt.Errorf("ledger-rest create transaction timeout")
			}
			if response.Status != 200 && response.Status != 201 && response.Status != 202 {
				return fmt.Errorf("ledger-rest create transaction %s error %+v", uri, response)
			}
			break
		}

		metrics.TransactionImported()
		metrics.TransfersImported(int64(len(transaction.Transfers)))

		for _, transfer := range transaction.Transfers {
			if transfer.ValueDateRaw.After(lastSynced) {
				lastSynced = transfer.ValueDateRaw
			}
		}

		if lastSynced.After(token.LastSyncedFrom[currency]) {
			token.LastSyncedFrom[currency] = lastSynced
			if !persistence.UpdateToken(storage, token) {
				log.WithField("token", token.ID).Warn("Unable to update token")
			}
		}
	}

	return nil
}

func importNewStatements(s *ActorSystem, token *model.Token, currency string, session *model.Session) {
	for _, interval := range utils.PartitionInterval(token.LastSyncedFrom[currency], time.Now()) {
		err := importStatementsForInterval(s.BondsterGateway, s.VaultGateway, s.LedgerGateway, s.Tenant, s.HttpClient, s.Storage, s.Metrics, token, currency, session, interval)
		if err != nil {
			log.WithField("token", token.ID).Errorf("Import statements failed with %+v", err)
			return
		}
	}
}

func importStatements(s *ActorSystem, token model.Token, callback func()) {
	log.WithField("token", token.ID).Debugf("Importing statements")

	session, err := integration.GetSession(s.HttpClient, s.BondsterGateway, token)
	if err != nil {
		log.WithField("token", token.ID).Warnf("Unable to get session because %+v", err)
		return
	}

	currencies, err := integration.GetCurrencies(s.HttpClient, s.BondsterGateway, session)
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
		importNewStatements(s, &token, currency, session)
		log.WithField("token", token.ID).Debugf("Import for currency %s End", currency)
		callback()
	}
}
