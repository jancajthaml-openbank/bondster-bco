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

package integration

import (
	"fmt"

	"github.com/jancajthaml-openbank/bondster-bco-import/model"
	"github.com/jancajthaml-openbank/bondster-bco-import/utils"
)

func GetCurrencies(client Client, gateway string, session *model.Session) ([]string, error) {
	var (
		err      error
		uri      string
	)

	uri = gateway + "/clientusersetting/api/private/market/getContactInformation"

	headers := map[string]string{
		"device":            session.Device,
		"channeluuid":       session.Channel,
		"ssid":              session.SSID,
		"x-active-language": "cs",
		"authorization":     "Bearer " + session.JWT,
		"host":              "ib.bondster.com",
		"origin":            "https://ib.bondster.com",
		"referer":           "https://ib.bondster.com/cs",
	}

	response, err := client.Post(uri, nil, headers)
	if err != nil {
		return nil, fmt.Errorf("bondster get contact information error %+v", err)
	}
	if response.Status != 200 {
		return nil, fmt.Errorf("bondster get contact information error %s", response.String())
	}

	var currencies = new(model.PotrfolioCurrencies)
	err = utils.JSON.Unmarshal(response.Data, currencies)
	if err != nil {
		return nil, err
	}

	return currencies.Value, nil
}

func GetTransactionIdsInInterval(client Client, gateway string, session *model.Session, currency string, interval utils.TimeRange) ([]string, error) {
	var (
		err      error
		response Response
		request  []byte
		uri      string
	)

	request, err = utils.JSON.Marshal(model.TransfersSearchRequest{
		From: interval.StartTime,
		To:   interval.EndTime,
	})
	if err != nil {
		return nil, err
	}

	uri = gateway + "/mktinvestor/api/private/transaction/search"

	headers := map[string]string{
		"device":            session.Device,
		"channeluuid":       session.Channel,
		"ssid":              session.SSID,
		"authorization":     "Bearer " + session.JWT,
		"x-account-context": currency,
		"x-active-language": "cs",
		"host":              "ib.bondster.com",
		"origin":            "https://ib.bondster.com",
		"referer":           "https://ib.bondster.com/cs/statement",
	}

	response, err = client.Post(uri, request, headers)
	if err != nil {
		return nil, fmt.Errorf("bondster get contact information error %+v", err)
	}
	if response.Status != 200 {
		return nil, fmt.Errorf("bondster get contact information error %s", response.String())
	}

	var search = new(model.TransfersSearchResult)
	err = utils.JSON.Unmarshal(response.Data, search)
	if err != nil {
		return nil, err
	}

	return search.IDs, nil
}



func GetTransactionDetails(client Client, gateway string, session *model.Session, currency string, transactionIds []string) (*model.BondsterImportEnvelope, error) {
	var (
		err      error
		response Response
		request  []byte
		uri      string
	)

	request, err = utils.JSON.Marshal(model.TransfersSearchResult{
		IDs: transactionIds,
	})
	if err != nil {
		return nil, err
	}

	uri = gateway + "/mktinvestor/api/private/transaction/list"

	headers := map[string]string{
		"device":            session.Device,
		"channeluuid":       session.Channel,
		"ssid":              session.SSID,
		"authorization":     "Bearer " + session.JWT,
		"x-account-context": currency,
		"x-active-language": "cs",
		"host":              "ib.bondster.com",
		"origin":            "https://ib.bondster.com",
		"referer":           "https://ib.bondster.com/cs/statement",
	}

	response, err = client.Post(uri, request, headers)
	if err != nil {
		return nil, fmt.Errorf("bondster get contact information error %+v", err)
	}
	if response.Status != 200 {
		return nil, fmt.Errorf("bondster get contact information error %s", response.String())
	}

	var envelope = new(model.BondsterImportEnvelope)
	err = utils.JSON.Unmarshal(response.Data, &(envelope.Transactions))
	if err != nil {
		return nil, err
	}
	envelope.Currency = currency

	return envelope, nil
}
