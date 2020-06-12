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

package http

import (
  "fmt"

  "github.com/jancajthaml-openbank/bondster-bco-import/utils"
)

type Response struct {
  Status int
  Data []byte
  Header map[string]string
}

func (value *Response) String() string {
  if value == nil {
    return "<nil>"
  }
  return fmt.Sprintf("Response{ Status: %d, Data: %s, Header: %+v }", value.Status, string(value.Data), value.Header)
}

// WebToken encrypted json web token and ssid of bondster session
type WebToken struct {
  JWT string
  SSID string
}

// Session hold bondster session headers
type Session struct {
  JWT     string
  Device  string
  Channel string
  SSID    string
}

// UnmarshalJSON is json JWT unmarhalling companion
func (entity *WebToken) UnmarshalJSON(data []byte) error {
  if entity == nil {
    return fmt.Errorf("cannot unmarshall to nil pointer")
  }
  all := struct {
    Result string `json:"result"`
    JWT    struct {
      Value string `json:"value"`
    } `json:"jwt"`
    SSID    struct {
      Value string `json:"value"`
    } `json:"ssid"`
  }{}
  err := utils.JSON.Unmarshal(data, &all)
  if err != nil {
    return err
  }
  if all.Result != "FINISH" {
    return fmt.Errorf("result %s has not finished, bailing out", all.Result)
  }
  if all.JWT.Value == "" {
    return fmt.Errorf("missing \"jwt\" value field")
  }
  if all.SSID.Value == "" {
    return fmt.Errorf("missing \"ssid\" value field")
  }
  entity.JWT = all.JWT.Value
  entity.SSID = all.SSID.Value
  return nil
}
