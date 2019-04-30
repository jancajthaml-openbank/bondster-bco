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

	"github.com/jancajthaml-openbank/bondster-bco-import/http"
	"github.com/jancajthaml-openbank/bondster-bco-import/model"
	"github.com/jancajthaml-openbank/bondster-bco-import/utils"

	log "github.com/sirupsen/logrus"
)

func GetSession(client http.Client, gateway string, token model.Token) (*model.Session, error) {
	device := utils.RandDevice()
	channel := utils.UUID()

	var (
		err      error
		response []byte
		request  []byte
		code     int
		uri      string
	)

	headers := map[string]string{
		"device":            device,
		"channeluuid":       channel,
		"x-active-language": "cs",
		"host":              "bondster.com",
		"origin":            "https://bondster.com",
		"referer":           "https://bondster.com/ib/cs",
		"accept":            "application/json",
	}

	uri = gateway + "/router/api/public/authentication/getLoginScenario"
	response, code, err = client.Post(uri, nil, headers)
	if err != nil {
		return nil, fmt.Errorf("bondster get login scenario Error %+v", err)
	}
	if code != 200 {
		return nil, fmt.Errorf("bondster get login scenario error %d %+v", code, string(response))
	}

	var scenario = new(model.LoginScenario)
	err = utils.JSON.Unmarshal(response, scenario)
	if err != nil {
		return nil, err
	}

	if scenario.Value != "USR_PWD" {
		return nil, fmt.Errorf("bondster unsupported login scenario %s", string(response))
	}

	step := model.LoginStep{
		Code: "USR_PWD",
		Values: []model.LoginStepValue{
			{
				Type:  "USERNAME",
				Value: token.Username,
			},
			{
				Type:  "PWD",
				Value: token.Password,
			},
		},
	}

	request, err = utils.JSON.Marshal(step)
	if err != nil {
		return nil, err
	}

	uri = gateway + "/router/api/public/authentication/validateLoginStep"
	response, code, err = client.Post(uri, request, headers)
	if err != nil {
		return nil, err
	}
	if code != 200 {
		return nil, fmt.Errorf("bondster validate login step error %d %+v", code, string(response))
	}

	var jwt = new(model.JWT)
	err = utils.JSON.Unmarshal(response, jwt)
	if err != nil {
		return nil, err
	}

	log.Debugf("Logged in with token %s", token.ID)

	session := &model.Session{
		JWT:     jwt.Value,
		Device:  device,
		Channel: channel,
	}

	return session, nil
}
