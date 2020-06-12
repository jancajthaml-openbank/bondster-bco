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

package ledger

import (
  "github.com/jancajthaml-openbank/bondster-bco-import/http"
)

// LedgerClient represents fascade for http client
type LedgerClient struct {
  underlying http.HttpClient
  gateway string
}

// NewLedgerClient returns new ledger http client
func NewLedgerClient(gateway string) LedgerClient {
  return LedgerClient{
    gateway: gateway,
    underlying: http.NewHttpClient(),
  }
}

// Post performs http POST request for given url with given body
func (client LedgerClient) Post(url string, body []byte, headers map[string]string) (http.Response, error) {
  return client.underlying.Post(client.gateway+url, body, headers)
}

// Get performs http GET request for given url
func (client LedgerClient) Get(url string, headers map[string]string) (http.Response, error) {
  return client.underlying.Get(client.gateway+url, headers)
}
