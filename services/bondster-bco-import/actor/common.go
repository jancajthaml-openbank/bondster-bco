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

	"github.com/jancajthaml-openbank/bondster-bco-import/model"

	system "github.com/jancajthaml-openbank/actor-system"
)

func parseMessage(msg string, to system.Coordinates) (interface{}, error) {
	start := 0
	end := len(msg)
	parts := make([]string, 3)
	idx := 0
	i := 0
	for i < end && idx < 3 {
		if msg[i] == 32 {
			if !(start == i && msg[start] == 32) {
				parts[idx] = msg[start:i]
				idx++
			}
			start = i + 1
		}
		i++
	}
	if idx < 3 && msg[start] != 32 && len(msg[start:]) > 0 {
		parts[idx] = msg[start:]
		idx++
	}

	if i != end {
		return nil, fmt.Errorf("message too large")
	}

	switch parts[0] {

	case SynchronizeTokens:
		return SynchronizeToken{}, nil

	case ReqCreateToken:
		if idx == 3 {
			return CreateToken{
				ID:       to.Name,
				Username: parts[1],
				Password: parts[2],
			}, nil
		}
		return nil, fmt.Errorf("invalid message %s", msg)

	case ReqDeleteToken:
		return DeleteToken{
			ID: to.Name,
		}, nil

	default:
		return nil, fmt.Errorf("unknown message %s", msg)
	}
}

// ProcessMessage processing of remote message to this bondster-bco
func ProcessMessage(s *ActorSystem) system.ProcessMessage {
	return func(msg string, to system.Coordinates, from system.Coordinates) {
		message, err := parseMessage(msg, to)
		if err != nil {
			log.Warnf("%s [remote %v -> local %v]", err, from, to)
			s.SendMessage(FatalError, from, to)
			return
		}
		ref, err := s.ActorOf(to.Name)
		if err != nil {
			ref, err = spawnTokenActor(s, to.Name)
		}
		if err != nil {
			log.Warnf("Actor not found [remote %v -> local %v]", from, to)
			s.SendMessage(FatalError, to, from)
			return
		}
		ref.Tell(message, to, from)
	}
}

func spawnTokenActor(s *ActorSystem, id string) (*system.Envelope, error) {
	envelope := system.NewEnvelope(id, model.NewToken(id))

	err := s.RegisterActor(envelope, NilToken(s))
	if err != nil {
		log.Warnf("%s ~ Spawning Actor Error unable to register", id)
		return nil, err
	}

	log.Debugf("%s ~ Actor Spawned", id)
	return envelope, nil
}
