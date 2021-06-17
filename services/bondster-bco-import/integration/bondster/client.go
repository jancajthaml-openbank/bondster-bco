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

package bondster

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/jancajthaml-openbank/bondster-bco-import/model"
	"github.com/jancajthaml-openbank/bondster-bco-import/support/http"
	"github.com/jancajthaml-openbank/bondster-bco-import/support/timeshift"
)

// Client represents fascade for bondster http interactions
type Client struct {
	gateway    string
	httpClient *AuthorizedClient
}

// NewClient returns new bondster http client
func NewClient(gateway string, token *model.Token) *Client {
	return &Client{
		gateway:    gateway,
		httpClient: NewAuthorizedClient(gateway, token),
	}
}

// GetCurrencies returns currencies tied to given token
func (client *Client) GetCurrencies() ([]string, error) {
	if client == nil {
		return nil, fmt.Errorf("nil deference")
	}
	req, err := http.NewRequest("POST", client.gateway+"/proxy/clientusersetting/api/private/market/getContactInformation", nil)
	if err != nil {
		return nil, fmt.Errorf("get contact information error %w", err)
	}
	resp, err := client.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("get contact information error %w", err)
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("get contact information error %s", resp.Status)
	}
	all := new(profileDetail)
	err = json.NewDecoder(resp.Body).Decode(all)
	if err != nil {
		return nil, err
	}
	currencies := make([]string, 0)
	for currency := range all.MarketAccounts.AccountsMap {
		currencies = append(currencies, currency)
	}
	sort.Strings(currencies)
	return currencies, nil
}

// GetStatementIdsInInterval returns transfer ids that happened during given interval
func (client *Client) GetStatementIdsInInterval(currency string, interval timeshift.TimeRange) ([]string, error) {
	if client == nil {
		return nil, fmt.Errorf("nil deference")
	}
	filter := searchFilter{
		ValueDateFrom: filterDate{
			Month: interval.StartTime.Month(),
			Year: interval.StartTime.Year(),
		},
		ValueDateTo: filterDate{
			Month: interval.EndTime.Month(),
			Year: interval.EndTime.Year(),
		},
	}
	payload, err := json.Marshal(filter)
	if err != nil {
		return nil, fmt.Errorf("get statements ids error %w", err)
	}
	req, err := http.NewRequest("POST", client.gateway+"/proxy/mktinvestor/api/private/transaction/search", payload)
	if err != nil {
		return nil, fmt.Errorf("get statements ids error %w", err)
	}
	req.SetHeader("Content-Type", "application/json")
	resp, err := client.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("get statements ids error %w", err)
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("get statements ids error %s", resp.Status)
	}
	all := new(transactionsIds)
	err = json.NewDecoder(resp.Body).Decode(all)
	if err != nil {
		return nil, err
	}
	return all.IDs, nil
}

// GetStatements returns statements for given currency and transaction ids
func (client *Client) GetStatements(currency string, transferIds []string) ([]BondsterStatement, error) {
	if client == nil {
		return nil, fmt.Errorf("nil deference")
	}
	filter := transactionFilter{
		TransactionIds: transferIds,
	}
	payload, err := json.Marshal(filter)
	if err != nil {
		return nil, fmt.Errorf("get statements error %w", err)
	}
	req, err := http.NewRequest("POST", client.gateway+"/proxy/mktinvestor/api/private/transaction/list", payload)
	if err != nil {
		return nil, fmt.Errorf("get statements error %w", err)
	}
	req.SetHeader("Content-Type", "application/json")
	resp, err := client.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("download statements error %w", err)
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("download statements error %s", resp.Status)
	}
	all := make([]BondsterStatement, 0)
	err = json.NewDecoder(resp.Body).Decode(&all)
	if err != nil {
		return nil, err
	}
	return all, nil
}
