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

		if config.RootStorage != "/data" {
			t.Errorf("RootStorage default value is not /data")
		}
		if config.EncryptionKey != nil {
			t.Errorf("EncryptionKey default value is not empty")
		}
		if config.ServerPort != 4001 {
			t.Errorf("ServerPort default value is not 4001")
		}
		if config.ServerKey != "" {
			t.Errorf("ServerKey default value is not empty")
		}
		if config.ServerCert != "" {
			t.Errorf("ServerCert default value is not empty")
		}
		if config.LakeHostname != "127.0.0.1" {
			t.Errorf("LakeHostname default value is not 127.0.0.1")
		}
		if config.LogLevel != "INFO" {
			t.Errorf("LogLevel default value is not INFO")
		}
		if config.MetricsRefreshRate != time.Second {
			t.Errorf("MetricsRefreshRate default value is not 1s")
		}
		if config.MetricsOutput != "/tmp/bondster-bco-rest-metrics" {
			t.Errorf("MetricsOutput default value is not /tmp/bondster-bco-rest-metrics")
		}
		if config.MinFreeDiskSpace != uint64(0) {
			t.Errorf("MinFreeDiskSpace default value is not 0")
		}
		if config.MinFreeMemory != uint64(0) {
			t.Errorf("MinFreeMemory default value is not 0")
		}
	}
}
