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

package env

import (
	"io/ioutil"
	"encoding/hex"
	"path/filepath"
	"time"
	"os"
)

// Get retrieves the string value of the environment variable named by the key
func Get(key string) (string, bool) {
	if v := os.Getenv(key); v != "" {
		return v, true
	}
	return "", false
}

// String retrieves the string value from environment named by the key.
func String(key string, fallback string) string {
	if str, exists := Get(key); exists {
		return str
	}
	return fallback
}

// Duration retrieves duration value from the environment.
func Duration(key string, fallback time.Duration) time.Duration {
	if str, exists := Get(key); exists {
		v, err := time.ParseDuration(str)
		if err != nil {
			log.Warn().Msgf("invalid value in %s, using fallback", key)
			return fallback
		}
		return v
	}
	return fallback
}

// HexFile retrieves bytes value from hexfile path defined from the environment.
func HexFile(key string, fallback []byte) []byte {
	if str, exists := Get(key); exists {
		data, err := ioutil.ReadFile(filepath.Clean(str))
		if err != nil {
			log.Warn().Msgf("invalid value in %s, using fallback", key)
			return fallback
		}
		if len(data) == 0 {
			log.Warn().Msgf("invalid value in %s, using fallback", key)
			return fallback
		}
		v, err := hex.DecodeString(string(data))
		if err != nil {
			log.Warn().Msgf("invalid value in %s, using fallback", key)
			return fallback
		}
		return []byte(v)
	}
	return fallback
}
