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
	"strings"
	"time"
)

// Configuration of application
type Configuration struct {
	// Tenant represent tenant of given vault
	Tenant string
	// BondsterGateway represent bondster gateway uri
	BondsterGateway string
	// SyncRate represents interval in which new statements are synchronized
	SyncRate time.Duration
	// LedgerGateway represent ledger-rest gateway uri
	LedgerGateway string
	// VaultGateway represent vault-rest gateway uri
	VaultGateway string
	// RootStorage gives where to store journals
	RootStorage string
	// EncryptionKey represents current encryption key
	EncryptionKey []byte
	// LakeHostname represent hostname of openbank lake service
	LakeHostname string
	// LogLevel ignorecase log level
	LogLevel string
	// MetricsRefreshRate represents interval in which in memory metrics should be
	// persisted to disk
	MetricsRefreshRate time.Duration
	// MetricsOutput represents output file for metrics persistence
	MetricsOutput string
}

// LoadConfig loads application configuration
func LoadConfig() Configuration {
	return Configuration{
		Tenant:             envString("BONDSTER_BCO_TENANT", ""),
		RootStorage:        envString("BONDSTER_BCO_STORAGE", "/data") + "/t_" + envString("BONDSTER_BCO_TENANT", "") + "/import/bondster",
		EncryptionKey:      envSecret("BONDSTER_BCO_ENCRYPTION_KEY", nil),
		BondsterGateway:    envString("BONDSTER_BCO_BONDSTER_GATEWAY", "https://bondster.com/ib/proxy"),
		LedgerGateway:      envString("BONDSTER_BCO_LEDGER_GATEWAY", "https://127.0.0.1:4401"),
		VaultGateway:       envString("BONDSTER_BCO_VAULT_GATEWAY", "https://127.0.0.1:4400"),
		LakeHostname:       envString("BONDSTER_BCO_LAKE_HOSTNAME", "127.0.0.1"),
		SyncRate:           envDuration("BONDSTER_BCO_SYNC_RATE", 22*time.Second),
		LogLevel:           strings.ToUpper(envString("BONDSTER_BCO_LOG_LEVEL", "INFO")),
		MetricsRefreshRate: envDuration("BONDSTER_BCO_METRICS_REFRESHRATE", time.Second),
		MetricsOutput:      envFilename("BONDSTER_BCO_METRICS_OUTPUT", "/tmp/bondster-bco-import-metrics"),
	}
}
