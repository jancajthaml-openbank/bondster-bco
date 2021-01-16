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
	"time"
)

// Token represents metadata of token entity
type Token struct {
	ID             string
	Username       string
	Password       string
	CreatedAt      time.Time
	LastSyncedFrom map[string]time.Time
}

// NewToken returns new Token
func NewToken(id string) Token {
	return Token{
		ID:             id,
		LastSyncedFrom: make(map[string]time.Time),
	}
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
	for currency, syncTime := range entity.LastSyncedFrom {
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
	entity.LastSyncedFrom = make(map[string]time.Time)

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
			entity.LastSyncedFrom[syncTime[:3]] = from
		} else {
			entity.LastSyncedFrom[syncTime[:3]] = time.Date(2017, 1, 1, 0, 0, 0, 0, time.UTC)
		}
	}

	return nil
}
