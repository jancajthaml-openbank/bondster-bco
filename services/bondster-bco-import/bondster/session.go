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

package bondster

import (
	"time"

	"github.com/jancajthaml-openbank/bondster-bco-import/utils"
)

// Session hold bondster session headers
type Session struct {
	JWT     *JWT
	SSID    *SSID
	Device  string
	Channel string
}

func NewSession() Session {
	return Session{
		JWT:     nil,
		SSID:    nil,
		Device:  utils.RandDevice(),
		Channel: utils.UUID(),
	}
}

func (session Session) IsSSIDExpired() bool {
	if session.JWT == nil || session.SSID == nil {
		return true
	}
	if time.Now().After(session.SSID.ExpiresAt.Add(time.Second * time.Duration(-10))) {
		return true
	}
	return false
}

func (session Session) IsJWTExpired() bool {
	if session.JWT == nil {
		return true
	}
	if time.Now().After(session.JWT.ExpiresAt.Add(time.Second * time.Duration(-10))) {
		return true
	}
	return false
}
