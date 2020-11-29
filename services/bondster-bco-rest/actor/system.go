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
	"github.com/jancajthaml-openbank/bondster-bco-rest/metrics"

	system "github.com/jancajthaml-openbank/actor-system"
)

// System represents actor system subroutine
type System struct {
	system.System
	Metrics *metrics.Metrics
}

// NewActorSystem returns actor system fascade
func NewActorSystem(endpoint string, metrics *metrics.Metrics) *System {
	sys, err := system.New("BondsterRest", endpoint)
	if err != nil {
		log.Error().Msgf("Failed to register actor system %+v", err)
		return nil
	}
	result := new(System)
	result.System = sys
	result.Metrics = metrics
	result.System.RegisterOnMessage(ProcessMessage(result))
	return result
}

func (system *System) Setup() error {
	return nil
}

func (system *System) Work() {
	if system == nil {
		return
	}
	system.System.Start()
}

func (system *System) Cancel() {
	if system == nil {
		return
	}
	system.System.Stop()
}

func (system *System) Done() <- chan interface{} {
	done := make(chan interface{})
	close(done)
	return done
}
