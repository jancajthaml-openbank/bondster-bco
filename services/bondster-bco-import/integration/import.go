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

package integration

import (
	"context"
	"fmt"
	"time"

	"github.com/jancajthaml-openbank/bondster-bco-import/model"
	"github.com/jancajthaml-openbank/bondster-bco-import/persistence"
	"github.com/jancajthaml-openbank/bondster-bco-import/utils"

	system "github.com/jancajthaml-openbank/actor-system"
	localfs "github.com/jancajthaml-openbank/local-fs"
	log "github.com/sirupsen/logrus"
)

// BondsterImport represents bondster gateway to ledger import subroutine
type BondsterImport struct {
	utils.DaemonSupport
	callback        func(msg interface{}, to system.Coordinates, from system.Coordinates)
	bondsterGateway string
	refreshRate     time.Duration
	storage         *localfs.Storage
}

// NewBondsterImport returns bondster import fascade
func NewBondsterImport(ctx context.Context, bondsterEndpoint string, syncRate time.Duration, storage *localfs.Storage, callback func(msg interface{}, to system.Coordinates, from system.Coordinates)) BondsterImport {
	return BondsterImport{
		DaemonSupport:   utils.NewDaemonSupport(ctx),
		callback:        callback,
		bondsterGateway: bondsterEndpoint,
		refreshRate:     syncRate,
		storage:         storage,
	}
}

func (bondster BondsterImport) getActiveTokens() ([]string, error) {
	tokens, err := persistence.LoadTokens(bondster.storage)
	if err != nil {
		return nil, err
	}
	uniq := make([]string, 0)
	visited := make(map[string]bool)
	for _, token := range tokens {
		if _, ok := visited[token.Username]; !ok {
			visited[token.Username] = true
			uniq = append(uniq, token.ID)
		}
	}
	return uniq, nil
}

func (bondster BondsterImport) importRoundtrip() {
	tokens, err := bondster.getActiveTokens()
	if err != nil {
		log.Errorf("unable to get active tokens %+v", err)
		return
	}

	if bondster.IsDone() {
		return
	}

	for _, item := range tokens {
		log.Debugf("Request to import token %s", item)
		msg := model.SynchronizeToken{}
		to := system.Coordinates{Name: item}
		from := system.Coordinates{Name: "token_import_cron"}
		bondster.callback(msg, to, from)
	}
}

// WaitReady wait for bondster import to be ready
func (bondster BondsterImport) WaitReady(deadline time.Duration) (err error) {
	defer func() {
		if e := recover(); e != nil {
			switch x := e.(type) {
			case string:
				err = fmt.Errorf(x)
			case error:
				err = x
			default:
				err = fmt.Errorf("unknown panic")
			}
		}
	}()

	ticker := time.NewTicker(deadline)
	select {
	case <-bondster.IsReady:
		ticker.Stop()
		err = nil
		return
	case <-ticker.C:
		err = fmt.Errorf("daemon was not ready within %v seconds", deadline)
		return
	}
}

// Start handles everything needed to start bondster import daemon
func (bondster BondsterImport) Start() {
	defer bondster.MarkDone()

	bondster.MarkReady()

	select {
	case <-bondster.CanStart:
		break
	case <-bondster.Done():
		return
	}

	log.Infof("Start bondster-import daemon, sync %v now and then each %v", bondster.bondsterGateway, bondster.refreshRate)

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
