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
	"github.com/jancajthaml-openbank/bondster-bco-rest/model"
)

const (
	// ReqTokens bondster message request code for "Get Tokens"
	ReqTokens = "GT"
	// RespTokens bondster message response code for "Get Tokens"
	RespTokens = "TG"
	// ReqCreateToken bondster message request code for "New Token"
	ReqCreateToken = "NT"
	// RespCreateToken bondster message response code for "New Token"
	RespCreateToken = "TN"
	// ReqDeleteToken bondster message request code for "Delete Token"
	ReqDeleteToken = "DT"
	// RespDeleteToken bondster message response code for "Delete Token"
	RespDeleteToken = "TD"
	// ReqSynchronizeToken bondster message request code for "Synchronize Token"
	ReqSynchronizeToken = "ST"
	// RespSynchronizeToken bondster message response code for "Synchronize Token"
	RespSynchronizeToken = "TS"
	// FatalError bondster message response code for "Error"
	FatalError = "EE"
)

// CreateTokenMessage is message for creation of new token
func CreateTokenMessage(token model.Token) string {
	return ReqCreateToken + " " + token.Username + " " + token.Password
}

// SynchronizeTokenMessage is message for immediate synchronization of existing token
func SynchronizeTokenMessage() string {
	return ReqSynchronizeToken
}

// DeleteTokenMessage is message for deletion of existing token
func DeleteTokenMessage() string {
	return ReqDeleteToken
}
