// Copyright (c) 2016-2019, Jan Cajthaml <jan.cajthaml@gmail.com>
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
		response []byte
		code     int
		uri      string
	)

	uri = gateway + "/clientusersetting/api/private/market/getContactInformation"

	headers := map[string]string{
		"device":        session.Device,
		"channeluuid":   session.Channel,
		"authorization": "Bearer " + session.JWT,
	}

	response, code, err = client.Post(uri, nil, headers)
	if err != nil {
		return nil, fmt.Errorf("bondster get contact information error %+v", err)
	}
	if code != 200 {
		return nil, fmt.Errorf("bondster get contact information error %d %+v", code, string(response))
	}

	var currencies = new(model.PotrfolioCurrencies)
	err = utils.JSON.Unmarshal(response, currencies)
	if err != nil {
		return nil, err
	}

	return currencies.Value, nil
}
