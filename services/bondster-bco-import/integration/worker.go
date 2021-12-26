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
	"github.com/jancajthaml-openbank/bondster-bco-import/persistence"

	localfs "github.com/jancajthaml-openbank/local-fs"
)

// BondsterImport represents bondster gateway to ledger import subroutine
type BondsterImport struct {
	callback func(token string)
	storage  localfs.Storage
}

// NewBondsterImport returns bondster import fascade
func NewBondsterImport(rootStorage string, storageKey []byte, callback func(token string)) *BondsterImport {
	storage, err := localfs.NewEncryptedStorage(rootStorage, storageKey)
	if err != nil {
		log.Error().Err(err).Msg("Failed to ensure storage")
		return nil
	}
	return &BondsterImport{
		callback: callback,
		storage:  storage,
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

// Setup does nothing
func (bondster BondsterImport) Setup() error {
	return nil
}

// Work performs import roundtrip
func (bondster BondsterImport) Work() {
	tokens, err := bondster.getActiveTokens()
	if err != nil {
		log.Error().Err(err).Msg("unable to get active tokens")
		return
	}

	for _, token := range tokens {
		log.Debug().Msgf("Request to import token %s", token)
		bondster.callback(token)
	}
}

// Cancel does nothing
func (bondster BondsterImport) Cancel() {

}

// Done always returns done
func (bondster BondsterImport) Done() <-chan interface{} {
	done := make(chan interface{})
	close(done)
	return done
}
