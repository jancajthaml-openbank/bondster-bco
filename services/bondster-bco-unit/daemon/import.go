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

package daemon

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/jancajthaml-openbank/bondster-bco-unit/config"
	"github.com/jancajthaml-openbank/bondster-bco-unit/http"
	"github.com/jancajthaml-openbank/bondster-bco-unit/model"
	"github.com/jancajthaml-openbank/bondster-bco-unit/persistence"
	"github.com/jancajthaml-openbank/bondster-bco-unit/utils"

	localfs "github.com/jancajthaml-openbank/local-fs"
	log "github.com/sirupsen/logrus"
)

// BondsterImport represents bondster gateway to wall import subroutine
type BondsterImport struct {
	Support
	tenant          string
	bondsterGateway string
	wallGateway     string
	vaultGateway    string
	refreshRate     time.Duration
	storage         *localfs.Storage
	metrics         *Metrics
	system          *ActorSystem
	httpClient      http.Client
}

// NewBondsterImport returns bondster import fascade
func NewBondsterImport(ctx context.Context, cfg config.Configuration, metrics *Metrics, system *ActorSystem, storage *localfs.Storage) BondsterImport {
	return BondsterImport{
		Support:         NewDaemonSupport(ctx),
		tenant:          cfg.Tenant,
		bondsterGateway: cfg.BondsterGateway,
		wallGateway:     cfg.WallGateway,
		vaultGateway:    cfg.VaultGateway,
		refreshRate:     cfg.SyncRate,
		storage:         storage,
		metrics:         metrics,
		system:          system,
		httpClient:      http.NewClient(),
	}
}

func (bondster BondsterImport) getLoginScenario(device string, channel string) error {
	var (
		err      error
		response []byte
		code     int
		uri      string
	)

	uri = bondster.bondsterGateway + "/router/api/public/authentication/getLoginScenario"

	headers := map[string]string{
		"device":            device,
		"channeluuid":       channel,
		"x-active-language": "cs",
		"host":              "bondster.com",
		"origin":            "https://bondster.com",
		"referer":           "https://bondster.com/ib/cs",
		"accept":            "application/json",
	}

	response, code, err = bondster.httpClient.Post(uri, nil, headers)
	if err != nil {
		return fmt.Errorf("bondster get login scenario Error %+v", err)
		return err
	} else if code != 200 {
		return fmt.Errorf("bondster get login scenario error %d %+v", code, string(response))
	}

	var scenario = new(model.LoginScenario)
	err = utils.JSON.Unmarshal(response, scenario)
	if err != nil {
		return err
	}

	if scenario.Value != "USR_PWD" {
		return fmt.Errorf("bondster unsupported login scenario %s", string(response))
	}

	return nil
}

func (bondster BondsterImport) validateLoginStep(device string, channel string, token model.Token) (*model.JWT, error) {
	var (
		err      error
		response []byte
		request  []byte
		code     int
		uri      string
	)

	step := model.LoginStep{
		Code: "USR_PWD",
		Values: []model.LoginStepValue{
			{
				Type:  "USERNAME",
				Value: token.Username,
			},
			{
				Type:  "PWD",
				Value: token.Password,
			},
		},
	}

	request, err = utils.JSON.Marshal(step)
	if err != nil {
		return nil, err
	}

	uri = bondster.bondsterGateway + "/router/api/public/authentication/validateLoginStep"

	headers := map[string]string{
		"device":            device,
		"channeluuid":       channel,
		"x-active-language": "cs",
		"host":              "bondster.com",
		"origin":            "https://bondster.com",
		"referer":           "https://bondster.com/ib/cs",
		"accept":            "application/json",
	}

	response, code, err = bondster.httpClient.Post(uri, request, headers)
	if err != nil {
		return nil, err
	} else if code != 200 {
		return nil, fmt.Errorf("bondster validate login step error %d %+v", code, string(response))
	}

	var session = new(model.JWT)
	err = utils.JSON.Unmarshal(response, session)
	if err != nil {
		return nil, err
	}
	return session, nil
}

func (bondster BondsterImport) getActiveTokens() ([]model.Token, error) {
	return persistence.LoadTokens(bondster.storage)
}

func (bondster BondsterImport) importNewTransactions(token *model.Token, currency string, session *model.Session) error {
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

	uri = bondster.bondsterGateway + "/mktinvestor/api/private/transaction/search"

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

	response, code, err = bondster.httpClient.Post(uri, request, headers)
	if err != nil {
		return fmt.Errorf("bondster transaction search error %+v, request: %+v", err, string(request))
	} else if code != 200 {
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

	uri = bondster.bondsterGateway + "/mktinvestor/api/private/transaction/list"

	response, code, err = bondster.httpClient.Post(uri, request, headers)
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
		uri := bondster.vaultGateway + "/account/" + bondster.tenant
		err = utils.Retry(10, time.Second, func() (err error) {
			response, code, err = bondster.httpClient.Post(uri, request, nil)
			if code == 200 || code == 409 || code == 400 {
				return
			} else if code >= 500 && err == nil {
				err = fmt.Errorf("vault POST %s error %d %+v", uri, code, string(response))
			}
			return
		})

		if err != nil {
			return err
		} else if code == 400 {
			return fmt.Errorf("vault account malformed request %+v", string(request))
		} else if code != 200 && code != 409 {
			return fmt.Errorf("vault POST %s error %d %+v", uri, code, string(response))
		}
	}

	var lastSynced time.Time = token.LastSyncedFrom[currency]

	for _, transaction := range envelope.GetTransactions() {

		for _, transfer := range transaction.Transfers {
			if transfer.ValueDateRaw.After(lastSynced) {
				lastSynced = transfer.ValueDateRaw
			}
		}

		request, err = utils.JSON.Marshal(transaction)
		if err != nil {
			return err
		}

		uri := bondster.wallGateway + "/transaction/" + bondster.tenant
		err = utils.Retry(10, time.Second, func() (err error) {
			response, code, err = bondster.httpClient.Post(uri, request, nil)
			if code == 200 || code == 201 || code == 400 {
				return
			} else if code >= 500 && err == nil {
				err = fmt.Errorf("wall POST %s error %d %+v", uri, code, string(response))
			}
			return
		})

		if err != nil {
			return err
		} else if code == 409 {
			return fmt.Errorf("wall transaction duplicate %+v", string(request))
		} else if code == 400 {
			return fmt.Errorf("wall transaction malformed request %+v", string(request))
		} else if code != 200 && code != 201 {
			return fmt.Errorf("wall POST %s error %d %+v", uri, code, string(response))
		}

		if lastSynced.After(token.LastSyncedFrom[currency]) {
			token.LastSyncedFrom[currency] = lastSynced
			if !persistence.UpdateToken(bondster.storage, token) {
				log.Warnf("Unable to update token %+v", token)
			}
		}

	}

	return nil
}

func (bondster BondsterImport) login(token model.Token) (session *model.Session, err error) {
	var jwt *model.JWT

	device := utils.RandDevice()
	channel := utils.UUID()

	if err = bondster.getLoginScenario(device, channel); err != nil {
		log.Warnf("Unable to get login scenario for token %+v", token.Value)
		return
	}

	if jwt, err = bondster.validateLoginStep(device, channel, token); err != nil {
		log.Warnf("Unable to validate login step for token %+v", token.Value)
		return
	}
	log.Debugf("Logged in with token %s", token.Value)

	session = &model.Session{
		JWT:     jwt.Value,
		Device:  device,
		Channel: channel,
	}

	return
}

func (bondster BondsterImport) getCurrencies(session *model.Session) ([]string, error) {
	var (
		err      error
		response []byte
		code     int
		uri      string
	)

	uri = bondster.bondsterGateway + "/clientusersetting/api/private/market/getContactInformation"

	headers := map[string]string{
		"device":        session.Device,
		"channeluuid":   session.Channel,
		"authorization": "Bearer " + session.JWT,
	}

	response, code, err = bondster.httpClient.Post(uri, nil, headers)
	if err != nil {
		return nil, fmt.Errorf("bondster get contact information error %+v", err)
	} else if code != 200 {
		return nil, fmt.Errorf("bondster get contact information error %d %+v", code, string(response))
	}

	var currencies = new(model.PotrfolioCurrencies)
	err = utils.JSON.Unmarshal(response, currencies)
	if err != nil {
		return nil, err
	}

	return currencies.Value, nil
}

func (bondster BondsterImport) importStatements(token model.Token) {
	session, err := bondster.login(token)
	if err != nil {
		log.Warnf("Unable to login because %+v", err)
		return
	}

	currencies, err := bondster.getCurrencies(session)
	if err != nil {
		log.Warnf("Unable to get contact information because %+v", err)
		return
	}

	if token.UpdateCurrencies(currencies) && !persistence.UpdateToken(bondster.storage, &token) {
		log.Errorf("update of token currencies has failed, currencies : %+v, token: %+v", currencies, token)
	}

	for currency := range token.LastSyncedFrom {
		if err := bondster.importNewTransactions(&token, currency, session); err != nil {
			log.Warnf("Import token %s statements failed with %+v", token.Value, err)
			continue
		}
	}
}

func (bondster BondsterImport) importRoundtrip() {
	var wg sync.WaitGroup

	tokens, err := bondster.getActiveTokens()
	if err != nil {
		log.Errorf("unable to get active tokens %+v", err)
		return
	}

	for _, item := range tokens {
		wg.Add(1)
		go func(token model.Token) {
			defer wg.Done()
			bondster.importStatements(token)
		}(item)
	}

	wg.Wait()
}

// Start handles everything needed to start bondster import daemon
func (bondster BondsterImport) Start() {
	defer bondster.MarkDone()

	log.Infof("Start bondster-import daemon, sync %v now and then each %v", bondster.bondsterGateway, bondster.refreshRate)
	bondster.MarkReady()

	bondster.importRoundtrip()

	for {
		select {
		case <-bondster.Done():
			log.Info("Stopping bondster-import daemon")
			log.Info("Stop bondster-import daemon")
			return
		case <-time.After(bondster.refreshRate):
			bondster.importRoundtrip()
		}
	}
}
