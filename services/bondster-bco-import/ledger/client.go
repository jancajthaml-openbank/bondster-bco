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
	"fmt"

	"github.com/jancajthaml-openbank/bondster-bco-import/http"
	"github.com/jancajthaml-openbank/bondster-bco-import/model"
	"github.com/jancajthaml-openbank/bondster-bco-import/utils"
)

// LedgerClient represents fascade for http client
type LedgerClient struct {
	underlying http.HttpClient
	gateway    string
}

// NewLedgerClient returns new ledger http client
func NewLedgerClient(gateway string) LedgerClient {
	return LedgerClient{
		gateway:    gateway,
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

func (client LedgerClient) CreateTransaction(tenant string, transaction model.Transaction) error {
	bounce := 0
	for {
		request, err := utils.JSON.Marshal(transaction)
		if err != nil {
			return err
		}
		uri := "/transaction/" + tenant
		response, err := client.Post(uri, request, nil)
		if err != nil {
			return fmt.Errorf("ledger-rest create transaction %s error %+v", uri, err)
		}
		if response.Status == 409 {
			bounce += 1
			if bounce > 3 {
				return fmt.Errorf("ledger-rest create transaction %s duplicate", uri)
			}
			continue
		}
		if response.Status == 400 {
			return fmt.Errorf("ledger-rest transaction malformed request %+v", string(request))
		}
		if response.Status == 504 {
			return fmt.Errorf("ledger-rest create transaction timeout")
		}
		if response.Status != 200 && response.Status != 201 && response.Status != 202 {
			return fmt.Errorf("ledger-rest create transaction %s error %s", uri, response.String())
		}
		return nil
	}
}
