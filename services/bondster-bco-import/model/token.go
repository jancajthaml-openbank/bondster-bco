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

package model

import (
	"bytes"
	"fmt"
	"strings"
	"sync"
	"time"
)

// Token represents metadata of token entity
type Token struct {
	ID         string
	Username   string
	Password   string
	CreatedAt  time.Time
	lastSynced map[string]time.Time
	mutex      sync.RWMutex
}

// NewToken returns new Token
func NewToken(id string) Token {
	return Token{
		ID:         id,
		lastSynced: make(map[string]time.Time),
		CreatedAt:  time.Now().UTC(),
		mutex:      sync.RWMutex{},
	}
}

// GetLastSyncedTime returns last synced time for given currency
func (entity *Token) GetLastSyncedTime(currency string) *time.Time {
	if entity == nil {
		return nil
	}
	entity.mutex.Lock()
	lastSyncedTime, ok := entity.lastSynced[currency]
	entity.mutex.Unlock()
	if !ok {
		return nil
	}
	return &lastSyncedTime
}

// SetLastSyncedTime sets last synced time for given currency
func (entity *Token) SetLastSyncedTime(currency string, lastSyncedTime time.Time) error {
	if entity == nil {
		return fmt.Errorf("called Token.SetLastSyncedTime over nil")
	}
	entity.mutex.Lock()
	entity.lastSynced[currency] = lastSyncedTime
	entity.mutex.Unlock()
	return nil
}

// GetCurrencies returns know currencies to token
func (entity *Token) GetCurrencies() []string {
	keys := make([]string, 0)
	if entity == nil {
		return keys
	}
	entity.mutex.Lock()
	for k := range entity.lastSynced {
		keys = append(keys, k)
	}
	entity.mutex.Unlock()
	return keys
}

// Serialize Token entity to persistable data
func (entity *Token) Serialize() ([]byte, error) {
	if entity == nil {
		return nil, fmt.Errorf("called Token.Serialize over nil")
	}
	var buffer bytes.Buffer
	buffer.WriteString(entity.CreatedAt.Format(time.RFC3339))
	buffer.WriteString("\n")
	buffer.WriteString(entity.Username)
	buffer.WriteString("\n")
	buffer.WriteString(entity.Password)
	for currency, syncTime := range entity.lastSynced {
		buffer.WriteString("\n")
		buffer.WriteString(currency)
		buffer.WriteString(" ")
		buffer.WriteString(syncTime.Format("02/01/2006"))
	}
	return buffer.Bytes(), nil
}

// Deserialize Token entity from persistent data
func (entity *Token) Deserialize(data []byte) error {
	if entity == nil {
		return fmt.Errorf("called Token.Deserialize over nil")
	}
	entity.lastSynced = make(map[string]time.Time)

	// FIXME more optimal split
	lines := strings.Split(string(data), "\n")
	if len(lines) < 3 {
		return fmt.Errorf("malformed data")
	}

	if cast, err := time.Parse(time.RFC3339, lines[0]); err == nil {
		entity.CreatedAt = cast
	}

	entity.Username = lines[1]
	entity.Password = lines[2]
	for _, syncTime := range lines[3:] {
		if len(syncTime) < 7 {
			continue
		}
		if from, err := time.Parse("02/01/2006", syncTime[4:]); err == nil {
			entity.lastSynced[syncTime[:3]] = from
		} else {
			entity.lastSynced[syncTime[:3]] = time.Date(2017, 1, 1, 0, 0, 0, 0, time.UTC)
		}
	}

	return nil
}
