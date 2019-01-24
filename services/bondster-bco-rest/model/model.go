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

package model

import (
	"strings"
	"time"
)

// ReplyTimeout message
type ReplyTimeout struct{}

// TokenCreated message
type TokenCreated struct{}

// TokenDeleted message
type TokenDeleted struct{}

// Token represents metadata of token entity
type Token struct {
	Value          string `json:"value"`
	Username       string
	Password       string
	LastSyncedFrom map[string]time.Time
}

func (entity *Token) MarshalJSON() ([]byte, error) {
	return []byte("{\"value\":\"" + entity.Value + "\"}"), nil
}

// Hydrate deserializes Token entity from persistent data
func (entity *Token) Hydrate(data []byte) {
	if entity == nil {
		return
	}
	lines := strings.Split(string(data), "\n")
	if len(lines) < 3 {
		return
	}
	entity.LastSyncedFrom = make(map[string]time.Time)
	for _, syncTime := range lines[2:] {
		if from, err := time.Parse("01/2006", syncTime[4:]); err == nil {
			entity.LastSyncedFrom[syncTime[:3]] = from
		} else {
			entity.LastSyncedFrom[syncTime[:3]] = time.Date(2000, 1, 0, 0, 0, 0, 0, time.UTC)
		}
	}
}
