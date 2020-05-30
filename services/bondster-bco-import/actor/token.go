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
	log "github.com/sirupsen/logrus"
)

// NilToken represents token that is neither existing neither non existing
func NilToken(s *ActorSystem) func(interface{}, system.Context) {
	return func(t_state interface{}, context system.Context) {
		state := t_state.(model.Token)

		tokenHydration := persistence.LoadToken(s.Storage, state.ID)

		if tokenHydration == nil {
			context.Self.Become(state, NonExistToken(s))
			log.Debugf("%s ~ Nil -> NonExist", state.ID)
		} else {
			context.Self.Become(*tokenHydration, ExistToken(s))
			log.Debugf("%s ~ Nil -> Exist", state.ID)
		}

		context.Self.Receive(context)
	}
}

// NonExistToken represents token that does not exist
func NonExistToken(s *ActorSystem) func(interface{}, system.Context) {
	return func(t_state interface{}, context system.Context) {
		state := t_state.(model.Token)

		switch msg := context.Data.(type) {

		case model.CreateToken:
			tokenResult := persistence.CreateToken(s.Storage, state.ID, msg.Username, msg.Password)

			if tokenResult == nil {
				s.SendMessage(FatalError, context.Sender, context.Receiver)
				log.Debugf("%s ~ (NonExist CreateToken) Error", state.ID)
				return
			}

			s.SendMessage(RespCreateToken, context.Sender, context.Receiver)
			log.Infof("New Token %s Created", state.ID)
			log.Debugf("%s ~ (NonExist CreateToken) OK", state.ID)
			s.Metrics.TokenCreated()

			context.Self.Become(*tokenResult, ExistToken(s))
			context.Self.Tell(model.SynchronizeToken{}, context.Receiver, context.Sender)

		case model.DeleteToken:
			s.SendMessage(FatalError, context.Sender, context.Receiver)
			log.Debugf("%s ~ (NonExist DeleteToken) Error", state.ID)

		case model.SynchronizeToken:
			break

		default:
			s.SendMessage(FatalError, context.Sender, context.Receiver)
			log.Debugf("%s ~ (NonExist Unknown Message) Error", state.ID)
		}

		return
	}
}

// ExistToken represents account that does exist
func ExistToken(s *ActorSystem) func(interface{}, system.Context) {
	return func(t_state interface{}, context system.Context) {
		state := t_state.(model.Token)

		switch context.Data.(type) {

		case model.CreateToken:
			s.SendMessage(FatalError, context.Sender, context.Receiver)
			log.Debugf("%s ~ (Exist CreateToken) Error", state.ID)

		case model.SynchronizeToken:
			log.Debugf("%s ~ (Exist SynchronizeToken) Begin", state.ID)
			importStatements(s, state)
			log.Debugf("%s ~ (Exist SynchronizeToken) End", state.ID)

		case model.DeleteToken:
			if !persistence.DeleteToken(s.Storage, state.ID) {
				s.SendMessage(FatalError, context.Sender, context.Receiver)
				log.Debugf("%s ~ (Exist DeleteToken) Error", state.ID)
				return
			}
			log.Infof("Token %s Deleted", state.ID)
			log.Debugf("%s ~ (Exist DeleteToken) OK", state.ID)
			s.Metrics.TokenDeleted()
			s.SendMessage(RespDeleteToken, context.Sender, context.Receiver)
			context.Self.Become(state, NonExistToken(s))

		default:
			s.SendMessage(FatalError, context.Sender, context.Receiver)
			log.Warnf("%s ~ (Exist Unknown Message) Error", state.ID)

		}

		return
	}
}

func importStatementsForTimeRange(bondsterGateway string, vaultGateway string, ledgerGateway string, tenant string, httpClient integration.Client, storage *localfs.EncryptedStorage, metrics *metrics.Metrics, token *model.Token, currency string, session *model.Session, fromDate time.Time, toDate time.Time) error {
	log.Debugf("Importing bondster statements from %+v to %+v", fromDate, toDate)

	var (
		err      error
		response []byte
		request  []byte
		code     int
		uri      string
	)

	request, err = utils.JSON.Marshal(model.TransfersSearchRequest{
		From: fromDate,
		To:   toDate,
	})
	if err != nil {
		return err
	}

	uri = bondsterGateway + "/mktinvestor/api/private/transaction/search"

	headers := map[string]string{
		"device":            session.Device,
		"channeluuid":       session.Channel,
		"authorization":     "Bearer " + session.JWT,
		"x-account-context": currency,
		"x-active-language": "cs",
		"host":              "ib.bondster.com",
		"origin":            "https://ib.bondster.com",
		"referer":           "https://ib.bondster.com/cs/statement",
	}

	metrics.TimeTransactionSearchLatency(func() {
		response, code, err = httpClient.Post(uri, request, headers)
	})

	if err != nil {
		return fmt.Errorf("bondster transaction search error %+v request: %+v", err, string(request))
	}
	if code != 200 {
		return fmt.Errorf("bondster transaction search error %d %+v request: %+v", code, string(response), string(request))
	}

	var search = new(model.TransfersSearchResult)
	err = utils.JSON.Unmarshal(response, search)
	if err != nil {
		return err
	}

	if len(search.IDs) == 0 {
		log.Debugf("No transaction occured between %+v and %+v", fromDate, toDate)

		if toDate.After(token.LastSyncedFrom[currency]) {
			token.LastSyncedFrom[currency] = toDate
			if !persistence.UpdateToken(storage, token) {
				log.Warnf("Unable to update token %+v", token)
			}
		}

		return nil
	}

	request, err = utils.JSON.Marshal(search)
	if err != nil {
		return err
	}

	uri = bondsterGateway + "/mktinvestor/api/private/transaction/list"

	metrics.TimeTransactionListLatency(func() {
		response, code, err = httpClient.Post(uri, request, headers)
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

		uri := vaultGateway + "/account/" + tenant
		response, code, err = httpClient.Post(uri, request, nil)
		if err != nil {
			return fmt.Errorf("vault-rest create account %s error %+v", uri, err)
		}
		if code == 400 {
			return fmt.Errorf("vault-rest account malformed request %+v", string(request))
		}
		if code == 504 {
			return fmt.Errorf("vault-rest create account timeout")
		}
		if code != 200 && code != 409 {
			return fmt.Errorf("vault-rest create account %s error %d %+v", uri, code, string(response))
		}
	}

	var lastSynced time.Time = token.LastSyncedFrom[currency]

	for _, transaction := range envelope.GetTransactions(tenant) {
		for _, transfer := range transaction.Transfers {
			if transfer.ValueDateRaw.After(lastSynced) {
				lastSynced = transfer.ValueDateRaw
			}
		}

		request, err = utils.JSON.Marshal(transaction)
		if err != nil {
			return err
		}

		uri := ledgerGateway + "/transaction/" + tenant
		response, code, err = httpClient.Post(uri, request, nil)
		if err != nil {
			return fmt.Errorf("ledger-rest create transaction %s error %+v", uri, err)
		}
		if code == 409 {
			return fmt.Errorf("ledger-rest transaction duplicate %+v", string(request))
		}
		if code == 400 {
			return fmt.Errorf("ledger-rest transaction malformed request %+v", string(request))
		}
		if code == 504 {
			return fmt.Errorf("ledger-rest create transaction timeout")
		}
		if code != 200 && code != 201 {
			return fmt.Errorf("ledger-rest create transaction %s error %d %+v", uri, code, string(response))
		}

		metrics.TransactionImported()
		metrics.TransfersImported(int64(len(transaction.Transfers)))

		if lastSynced.After(token.LastSyncedFrom[currency]) {
			token.LastSyncedFrom[currency] = lastSynced
			if !persistence.UpdateToken(storage, token) {
				log.Warnf("Unable to update token %+v", token)
			}
		}
	}

	return nil
}

func importNewStatements(s *ActorSystem, token *model.Token, currency string, session *model.Session) {
	startTime := token.LastSyncedFrom[currency]
	endTime := time.Now()

	months := utils.GetMonthsWithin(startTime, endTime)

	// FIXME import by weeks

	for _, firstDate := range months {
		lastDate := firstDate.AddDate(0, 1, 0).Add(time.Nanosecond*-1)
		if lastDate.After(endTime) {
			lastDate = endTime
		}
		if firstDate.Before(startTime) {
			firstDate = startTime
		}

		log.Debugf("Importing bondster statements from %+v to %+v", firstDate, lastDate)
		err := importStatementsForTimeRange(s.BondsterGateway, s.VaultGateway, s.LedgerGateway, s.Tenant, s.HttpClient, s.Storage, s.Metrics, token, currency, session, firstDate, lastDate)
		if err != nil {
			log.Errorf("Import token %s statements failed with %+v", token.ID, err)
			return
		}
	}

	return
}

func importStatements(s *ActorSystem, token model.Token) {
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
		importNewStatements(s, &token, currency, session)
		log.Debugf("Import %+v %s End", token.ID, currency)
	}
}
