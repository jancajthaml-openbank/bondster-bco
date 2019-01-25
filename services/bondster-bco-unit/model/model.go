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
	"bytes"
	"strings"
	"time"

	"github.com/jancajthaml-openbank/bondster-bco-unit/utils"
)

// Token represents metadata of token entity
type Token struct {
	Value          string
	Username       string
	Password       string
	LastSyncedFrom map[string]time.Time
}

// ListTokens is inbound request for listing of existing tokens
type ListTokens struct {
}

// CreateToken is inbound request for creation of new token
type CreateToken struct {
	Value    string
	Username string
	Password string
}

// DeleteToken is inbound request for deletion of token
type DeleteToken struct {
	Value string
}

// GetToken is inbound request for token details
type GetToken struct {
}

// NewToken returns new Token
func NewToken(value string, username string, password string) Token {
	return Token{
		Value:          value,
		Username:       username,
		Password:       password,
		LastSyncedFrom: make(map[string]time.Time),
	}
}

func (entity *Token) UpdateCurrencies(currencies []string) bool {
	var updated = false
	for _, currency := range currencies {
		if _, ok := entity.LastSyncedFrom[currency]; !ok {
			entity.LastSyncedFrom[currency] = time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
			updated = true
		}
	}
	return updated
}

// Persist serializes Token entity to persistable data
func (entity *Token) Persist() []byte {
	if entity == nil {
		return nil
	}
	var buffer bytes.Buffer
	buffer.WriteString(utils.EncryptString(entity.Username))
	buffer.WriteString("\n")
	buffer.WriteString(utils.EncryptString(entity.Password))
	for currency, syncTime := range entity.LastSyncedFrom {
		buffer.WriteString("\n")
		buffer.WriteString(currency)
		buffer.WriteString(" ")
		buffer.WriteString(syncTime.Format("01/2006"))
	}
	return buffer.Bytes()
}

// Hydrate deserializes Token entity from persistent data
func (entity *Token) Hydrate(data []byte) {
	if entity == nil {
		return
	}
	lines := strings.Split(string(data), "\n")
	if len(lines) < 2 {
		return
	}
	entity.Username = utils.DecryptString(lines[0])
	entity.Password = utils.DecryptString(lines[1])
	entity.LastSyncedFrom = make(map[string]time.Time)
	for _, syncTime := range lines[2:] {
		if from, err := time.Parse("01/2006", syncTime[4:]); err == nil {
			entity.LastSyncedFrom[syncTime[:3]] = from
		} else {
			entity.LastSyncedFrom[syncTime[:3]] = time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
		}
	}
}
