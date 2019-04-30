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

package actor

import (
	"fmt"
	"time"

	"github.com/jancajthaml-openbank/bondster-bco-import/daemon"
	"github.com/jancajthaml-openbank/bondster-bco-import/integration"
	"github.com/jancajthaml-openbank/bondster-bco-import/model"
	"github.com/jancajthaml-openbank/bondster-bco-import/persistence"
	"github.com/jancajthaml-openbank/bondster-bco-import/utils"

	system "github.com/jancajthaml-openbank/actor-system"
	log "github.com/sirupsen/logrus"
)

// NilToken represents token that is neither existing neither non existing
func NilToken(s *daemon.ActorSystem) func(interface{}, system.Context) {
	return func(t_state interface{}, context system.Context) {
		state := t_state.(model.Token)

		tokenHydration := persistence.LoadToken(s.Storage, state.ID)

		if tokenHydration == nil {
			context.Receiver.Become(state, NonExistToken(s))
			log.Debugf("%s ~ Nil -> NonExist", state.ID)
		} else {
			context.Receiver.Become(*tokenHydration, ExistToken(s))
			log.Debugf("%s ~ Nil -> Exist", state.ID)
		}

		context.Receiver.Receive(context)
	}
}

// NonExistToken represents token that does not exist
func NonExistToken(s *daemon.ActorSystem) func(interface{}, system.Context) {
	return func(t_state interface{}, context system.Context) {
		state := t_state.(model.Token)

		switch msg := context.Data.(type) {

		case model.CreateToken:
			tokenResult := persistence.CreateToken(s.Storage, state.ID, msg.Username, msg.Password)

			if tokenResult == nil {
				s.SendRemote(context.Sender.Region, FatalErrorMessage(context.Receiver.Name, context.Sender.Name))
				log.Debugf("%s ~ (NonExist CreateToken) Error", state.ID)
				return
			}

			s.SendRemote(context.Sender.Region, TokenCreatedMessage(context.Receiver.Name, context.Sender.Name))
			log.Infof("New Token %s Created", state.ID)
			log.Debugf("%s ~ (NonExist CreateToken) OK", state.ID)
			s.Metrics.TokenCreated()

			context.Receiver.Become(*tokenResult, ExistToken(s))
			context.Receiver.Tell(model.SynchronizeToken{}, context.Sender)

		case model.DeleteToken:
			s.SendRemote(context.Sender.Region, FatalErrorMessage(context.Receiver.Name, context.Sender.Name))
			log.Debugf("%s ~ (NonExist DeleteToken) Error", state.ID)

		case model.SynchronizeToken:
			break

		default:
			s.SendRemote(context.Sender.Region, FatalErrorMessage(context.Receiver.Name, context.Sender.Name))
			log.Debugf("%s ~ (NonExist Unknown Message) Error", state.ID)
		}

		return
	}
}

// ExistToken represents account that does exist
func ExistToken(s *daemon.ActorSystem) func(interface{}, system.Context) {
	return func(t_state interface{}, context system.Context) {
		state := t_state.(model.Token)

		switch context.Data.(type) {

		case model.CreateToken:
			s.SendRemote(context.Sender.Region, FatalErrorMessage(context.Receiver.Name, context.Sender.Name))
			log.Debugf("%s ~ (Exist CreateToken) Error", state.ID)

		case model.SynchronizeToken:
			log.Debugf("%s ~ (Exist SynchronizeToken) Begin", state.ID)
			importStatements(s, state)
			log.Debugf("%s ~ (Exist SynchronizeToken) End", state.ID)

		case model.DeleteToken:
			if !persistence.DeleteToken(s.Storage, state.ID) {
				s.SendRemote(context.Sender.Region, FatalErrorMessage(context.Receiver.Name, context.Sender.Name))
				log.Debugf("%s ~ (Exist DeleteToken) Error", state.ID)
				return
			}
			log.Infof("Token %s Deleted", state.ID)
			log.Debugf("%s ~ (Exist DeleteToken) OK", state.ID)
			s.Metrics.TokenDeleted()
			s.SendRemote(context.Sender.Region, TokenDeletedMessage(context.Receiver.Name, context.Sender.Name))
			context.Receiver.Become(state, NonExistToken(s))

		default:
			s.SendRemote(context.Sender.Region, FatalErrorMessage(context.Receiver.Name, context.Sender.Name))
			log.Warnf("%s ~ (Exist Unknown Message) Error", state.ID)

		}

		return
	}
}

func importNewTransactions(s *daemon.ActorSystem, token *model.Token, currency string, session *model.Session) error {
	var (
		err      error
		response []byte
		request  []byte
		code     int
		uri      string
	)

	criteria := model.TransfersSearchRequest{
		From: token.LastSyncedFrom[currency],
		To:   time.Now(),
	}

	request, err = utils.JSON.Marshal(criteria)
	if err != nil {
		return err
	}

	uri = s.BondsterGateway + "/mktinvestor/api/private/transaction/search"

	headers := map[string]string{
		"device":            session.Device,
		"channeluuid":       session.Channel,
		"authorization":     "Bearer " + session.JWT,
		"x-account-context": currency,
		"x-active-language": "cs",
		"host":              "bondster.com",
		"origin":            "https://bondster.com",
		"referer":           "https://bondster.com/ib/cs",
		"accept":            "application/json",
	}

	s.Metrics.TimeTransactionSearchLatency(func() {
		response, code, err = s.HttpClient.Post(uri, request, headers)
	})

	if err != nil {
		return fmt.Errorf("bondster transaction search error %+v, request: %+v", err, string(request))
	}
	if code != 200 {
		return fmt.Errorf("bondster transaction search error %d %+v, request: %+v", code, string(response), string(request))
	}

	var search = new(model.TransfersSearchResult)
	err = utils.JSON.Unmarshal(response, search)

	if err != nil {
		return err
	}

	request, err = utils.JSON.Marshal(search)
	if err != nil {
		return err
	}

	uri = s.BondsterGateway + "/mktinvestor/api/private/transaction/list"

	s.Metrics.TimeTransactionListLatency(func() {
		response, code, err = s.HttpClient.Post(uri, request, headers)
	})

	if err != nil {
		return fmt.Errorf("bondster transaction list error %+v, request: %+v", err, string(request))
	} else if code != 200 {
		return fmt.Errorf("bondster transaction list error %d %+v, request: %+v", code, string(response), string(request))
	}

	var envelope = new(model.BondsterImportEnvelope)
	err = utils.JSON.Unmarshal(response, &(envelope.Transactions))
	if err != nil {
		return err
	}
	envelope.Currency = currency

	for _, account := range envelope.GetAccounts() {
		request, err = utils.JSON.Marshal(account)
		if err != nil {
			return err
		}

		uri := s.VaultGateway + "/account/" + s.Tenant
		response, code, err = s.HttpClient.Post(uri, request, nil)
		if err != nil {
			return fmt.Errorf("vault-rest create account %s error %+v", uri, err)
		}
		if code == 400 {
			return fmt.Errorf("vault-rest account malformed request %+v", string(request))
		}
		if code != 200 && code != 409 {
			return fmt.Errorf("vault-rest create account %s error %d %+v", uri, code, string(response))
		}
	}

	var lastSynced time.Time = token.LastSyncedFrom[currency]

	for _, transaction := range envelope.GetTransactions(s.Tenant) {

		for _, transfer := range transaction.Transfers {
			if transfer.ValueDateRaw.After(lastSynced) {
				lastSynced = transfer.ValueDateRaw
			}
		}

		request, err = utils.JSON.Marshal(transaction)
		if err != nil {
			return err
		}

		uri := s.LedgerGateway + "/transaction/" + s.Tenant
		response, code, err = s.HttpClient.Post(uri, request, nil)
		if err != nil {
			return fmt.Errorf("ledger-rest create transaction %s error %+v", uri, err)
		}
		if code == 409 {
			return fmt.Errorf("ledger-rest transaction duplicate %+v", string(request))
		}
		if code == 400 {
			return fmt.Errorf("ledger-rest transaction malformed request %+v", string(request))
		}
		if code != 200 && code != 201 {
			return fmt.Errorf("ledger-rest create transaction %s error %d %+v", uri, code, string(response))
		}

		s.Metrics.TransactionImported()
		s.Metrics.TransfersImported(int64(len(transaction.Transfers)))

		if lastSynced.After(token.LastSyncedFrom[currency]) {
			token.LastSyncedFrom[currency] = lastSynced
			if !persistence.UpdateToken(s.Storage, token) {
				log.Warnf("Unable to update token %+v", token)
			}
		}

	}

	return nil
}

func importStatements(s *daemon.ActorSystem, token model.Token) {
	log.Debugf("Importing statements for %s", token.ID)

	session, err := integration.GetSession(s.HttpClient, s.BondsterGateway, token)
	if err != nil {
		log.Warnf("Unable to get session for %s because %+v", token.ID, err)
		return
	}

	currencies, err := integration.GetCurrencies(s.HttpClient, s.BondsterGateway, session)
	if err != nil {
		log.Warnf("Unable to get currencies for %s because %+v", token.ID, err)
		return
	}

	if token.UpdateCurrencies(currencies) && !persistence.UpdateToken(s.Storage, &token) {
		log.Warnf("Update of token currencies has failed, currencies: %+v, token: %s", currencies, token.ID)
	}

	for currency := range token.LastSyncedFrom {
		log.Debugf("Import %+v %s Begin", token.ID, currency)
		if err := importNewTransactions(s, &token, currency, session); err != nil {
			log.Warnf("Import token %s statements failed with %+v", token.ID, err)
		}
		log.Debugf("Import %+v %s End", token.ID, currency)
	}
}
