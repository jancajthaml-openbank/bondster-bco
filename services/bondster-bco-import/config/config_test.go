package config

import (
	"os"
	"strings"
	"testing"
	"time"
)

func TestGetConfig(t *testing.T) {
	for _, v := range os.Environ() {
		k := strings.Split(v, "=")[0]
		if strings.HasPrefix(k, "BONDSTER_BCO") {
			os.Unsetenv(k)
		}
	}

	t.Log("has defaults for all values")
	{
		config := LoadConfig()

		if config.Tenant != "" {
			t.Errorf("Tenant default value is not empty")
		}
		if config.SyncRate != 22*time.Second {
			t.Errorf("SyncRate default value is not 22s")
		}
		if config.BondsterGateway != "https://bondster.com/ib/proxy" {
			t.Errorf("BondsterGateway default value is not https://bondster.com/ib/proxy")
		}
		if config.LedgerGateway != "https://127.0.0.1:4401" {
			t.Errorf("LedgerGateway default value is not https://127.0.0.1:4401")
		}
		if config.VaultGateway != "https://127.0.0.1:4400" {
			t.Errorf("VaultGateway default value is not https://127.0.0.1:4400")
		}
		if config.RootStorage != "/data/t_/import/bondster" {
			t.Errorf("RootStorage default value is not /data/t_/import/bondster")
		}
		if config.EncryptionKey != nil {
			t.Errorf("EncryptionKey default value is not empty")
		}
		if config.LakeHostname != "127.0.0.1" {
			t.Errorf("LakeHostname default value is not 127.0.0.1")
		}
		if config.LogLevel != "INFO" {
			t.Errorf("LogLevel default value is not INFO")
		}
		if config.MetricsStastdEndpoint != "127.0.0.1:8125" {
			t.Errorf("MetricsStastdEndpoint default value is not 127.0.0.1:8125")
		}
	}
}
