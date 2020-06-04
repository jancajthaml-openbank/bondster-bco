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

package config

import (
	"encoding/hex"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

func loadConfFromEnv() Configuration {
	logLevel := strings.ToUpper(getEnvString("BONDSTER_BCO_LOG_LEVEL", "DEBUG"))
	encryptionKey := getEnvString("BONDSTER_BCO_ENCRYPTION_KEY", "")
	rootStorage := getEnvString("BONDSTER_BCO_STORAGE", "/data")
	tenant := getEnvString("BONDSTER_BCO_TENANT", "")
	bondsterGateway := getEnvString("BONDSTER_BCO_BONDSTER_GATEWAY", "https://bondster.com/ib/proxy")
	syncRate := getEnvDuration("BONDSTER_BCO_SYNC_RATE", 22*time.Second)
	ledgerGateway := getEnvString("BONDSTER_BCO_LEDGER_GATEWAY", "https://127.0.0.1:4401")
	vaultGateway := getEnvString("BONDSTER_BCO_VAULT_GATEWAY", "https://127.0.0.1:4400")
	lakeHostname := getEnvString("BONDSTER_BCO_LAKE_HOSTNAME", "")
	metricsOutput := getEnvFilename("BONDSTER_BCO_METRICS_OUTPUT", "/tmp")
	metricsRefreshRate := getEnvDuration("BONDSTER_BCO_METRICS_REFRESHRATE", time.Second)

	if tenant == "" || lakeHostname == "" || rootStorage == "" || encryptionKey == "" {
		log.Fatal("missing required parameter to run")
	}

	keyData, err := ioutil.ReadFile(encryptionKey)
	if err != nil {
		log.Fatalf("unable to load encryption key from %s", encryptionKey)
	}

	key, err := hex.DecodeString(string(keyData))
	if err != nil {
		log.Fatalf("invalid encryption key %+v at %s", err, encryptionKey)
	}

	return Configuration{
		Tenant:             tenant,
		RootStorage:        rootStorage + "/" + tenant + "/import/bondster",
		EncryptionKey:      []byte(key),
		BondsterGateway:    bondsterGateway,
		SyncRate:           syncRate,
		LedgerGateway:      ledgerGateway,
		VaultGateway:       vaultGateway,
		LakeHostname:       lakeHostname,
		LogLevel:           logLevel,
		MetricsRefreshRate: metricsRefreshRate,
		MetricsOutput:      metricsOutput,
	}
}

func getEnvFilename(key string, fallback string) string {
	var value = strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	value = filepath.Clean(value)
	if os.MkdirAll(value, os.ModePerm) != nil {
		return fallback
	}
	return value
}
func getEnvString(key string, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func getEnvInteger(key string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	cast, err := strconv.Atoi(value)
	if err != nil {
		log.Errorf("invalid value of variable %s", key)
		return fallback
	}
	return cast
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	cast, err := time.ParseDuration(value)
	if err != nil {
		log.Errorf("invalid value of variable %s", key)
		return fallback
	}
	return cast
}
