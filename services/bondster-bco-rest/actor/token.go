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

package actor

import (
	"time"

	"github.com/rs/xid"

	"github.com/jancajthaml-openbank/bondster-bco-rest/model"

	system "github.com/jancajthaml-openbank/actor-system"
)

// CreateToken creates new token for target tenant
func CreateToken(sys *System, tenant string, token model.Token) interface{} {
	ch := make(chan interface{})

	envelope := system.NewActor("relay/"+xid.New().String(), nil)
	defer sys.UnregisterActor(envelope.Name)

	sys.RegisterActor(envelope, func(state interface{}, context system.Context) {
		ch <- context.Data
	})

	sys.SendMessage(
		CreateTokenMessage(token),
		system.Coordinates{
			Region: "BondsterImport/" + tenant,
			Name:   token.ID,
		},
		system.Coordinates{
			Region: "BondsterRest",
			Name:   envelope.Name,
		},
	)

	select {
	case result := <-ch:
		return result
	case <-time.After(5 * time.Second):
		return new(ReplyTimeout)
	}
}

// SynchronizeToken request immediate synchronization oftoken for target tenant
func SynchronizeToken(sys *System, tenant string, token string) interface{} {
	ch := make(chan interface{})

	envelope := system.NewActor("relay/"+xid.New().String(), nil)
	defer sys.UnregisterActor(envelope.Name)

	sys.RegisterActor(envelope, func(state interface{}, context system.Context) {
		ch <- context.Data
	})

	sys.SendMessage(
		SynchronizeTokenMessage(),
		system.Coordinates{
			Region: "BondsterImport/" + tenant,
			Name:   token,
		},
		system.Coordinates{
			Region: "BondsterRest",
			Name:   envelope.Name,
		},
	)

	select {
	case result := <-ch:
		return result
	case <-time.After(5 * time.Second):
		return new(ReplyTimeout)
	}
}

// DeleteToken deletes existing token for target tenant
func DeleteToken(sys *System, tenant string, tokenID string) interface{} {
	ch := make(chan interface{})

	envelope := system.NewActor("relay/"+xid.New().String(), nil)
	defer sys.UnregisterActor(envelope.Name)

	sys.RegisterActor(envelope, func(state interface{}, context system.Context) {
		ch <- context.Data
	})

	sys.SendMessage(
		DeleteTokenMessage(),
		system.Coordinates{
			Region: "BondsterImport/" + tenant,
			Name:   tokenID,
		},
		system.Coordinates{
			Region: "BondsterRest",
			Name:   envelope.Name,
		},
	)

	select {
	case result := <-ch:
		return result
	case <-time.After(5 * time.Second):
		return new(ReplyTimeout)
	}
}
