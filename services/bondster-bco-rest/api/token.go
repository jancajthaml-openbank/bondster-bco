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

package api

import (
	"encoding/json"
	"fmt"
	"github.com/jancajthaml-openbank/bondster-bco-rest/actor"
	"github.com/jancajthaml-openbank/bondster-bco-rest/model"
	"github.com/jancajthaml-openbank/bondster-bco-rest/persistence"
	localfs "github.com/jancajthaml-openbank/local-fs"
	"github.com/labstack/echo/v4"
	"io/ioutil"
	"net/http"
)

// DeleteToken removes existing token
func DeleteToken(system *actor.System) func(c echo.Context) error {
	return func(c echo.Context) error {
		c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSONCharsetUTF8)

		tenant := c.Param("tenant")
		if tenant == "" {
			return fmt.Errorf("missing tenant")
		}
		id := c.Param("id")
		if id == "" {
			return fmt.Errorf("missing id")
		}

		switch actor.DeleteToken(system, tenant, id).(type) {

		case *actor.TokenDeleted:
			log.Debug().Msgf("Token %s Deleted", id)
			c.Response().WriteHeader(http.StatusOK)
			return nil

		case *actor.TokenMissing:
			log.Debug().Msgf("Token %s Deletion does not exist", id)
			c.Response().WriteHeader(http.StatusNotFound)
			return nil

		case *actor.ReplyTimeout:
			log.Debug().Msgf("Token %s Deletion Timeout", id)
			c.Response().WriteHeader(http.StatusGatewayTimeout)
			return nil

		default:
			return fmt.Errorf("interval server error")

		}
	}
}

// CreateToken creates new token for given tenant
func CreateToken(system *actor.System) func(c echo.Context) error {
	return func(c echo.Context) error {
		c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSONCharsetUTF8)

		tenant := c.Param("tenant")
		if tenant == "" {
			return fmt.Errorf("missing tenant")
		}

		b, err := ioutil.ReadAll(c.Request().Body)
		defer c.Request().Body.Close()
		if err != nil {
			c.Response().WriteHeader(http.StatusBadRequest)
			return err
		}

		var req = new(model.Token)
		if json.Unmarshal(b, req) != nil {
			c.Response().WriteHeader(http.StatusBadRequest)
			return nil
		}

		switch actor.CreateToken(system, tenant, *req).(type) {

		case *actor.TokenCreated:
			log.Debug().Msgf("Token %s Created", req.ID)
			c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextPlainCharsetUTF8)
			c.Response().WriteHeader(http.StatusOK)
			c.Response().Write([]byte(req.ID))
			c.Response().Flush()
			return nil

		case *actor.ReplyTimeout:
			log.Debug().Msgf("Token %s Creation Timeout", req.ID)
			c.Response().WriteHeader(http.StatusGatewayTimeout)
			return nil

		default:
			return fmt.Errorf("interval server error")

		}
	}
}

// SynchronizeToken orders token to synchronize now
func SynchronizeToken(system *actor.System) func(c echo.Context) error {
	return func(c echo.Context) error {
		c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSONCharsetUTF8)

		tenant := c.Param("tenant")
		if tenant == "" {
			return fmt.Errorf("missing tenant")
		}

		id := c.Param("id")
		if id == "" {
			return fmt.Errorf("missing id")
		}

		switch actor.SynchronizeToken(system, tenant, id).(type) {

		case *actor.TokenSynchonizeAccepted:
			log.Debug().Msgf("Token %s Synchonizing", id)
			c.Response().WriteHeader(http.StatusOK)
			return nil
	
		case *actor.TokenMissing:
			log.Debug().Msgf("Token %s Synchonizing does not exist", id)
			c.Response().WriteHeader(http.StatusNotFound)
			return nil

		case *actor.ReplyTimeout:
			log.Debug().Msgf("Token %s Synchonizing Timeout", id)
			c.Response().WriteHeader(http.StatusGatewayTimeout)
			return nil

		
		default:
			return fmt.Errorf("interval server error")

		}
	}
}

// GetTokens return existing tokens of given tenant
func GetTokens(storage localfs.Storage) func(c echo.Context) error {
	return func(c echo.Context) error {
		c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSONCharsetUTF8)

		tenant := c.Param("tenant")
		if tenant == "" {
			return fmt.Errorf("missing tenant")
		}

		tokens, err := persistence.LoadTokens(storage, tenant)
		if err != nil {
			return err
		}

		c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextPlainCharsetUTF8)
		c.Response().WriteHeader(http.StatusOK)

		for idx, token := range tokens {
			chunk, err := json.Marshal(token)
			if err != nil {
				return err
			}
			if idx == len(tokens)-1 {
				c.Response().Write(chunk)
			} else {
				c.Response().Write(chunk)
				c.Response().Write([]byte("\n"))
			}
			c.Response().Flush()
		}

		return nil
	}
}
