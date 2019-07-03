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

package metrics

import (
	"bytes"
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/jancajthaml-openbank/bondster-bco-import/utils"
	metrics "github.com/rcrowley/go-metrics"
)

// Metrics represents metrics subroutine
type Metrics struct {
	utils.DaemonSupport
	output                   string
	refreshRate              time.Duration
	createdTokens            metrics.Counter
	deletedTokens            metrics.Counter
	transactionSearchLatency metrics.Timer
	transactionListLatency   metrics.Timer
	importedTransfers        metrics.Meter
	importedTransactions     metrics.Meter
}

// NewMetrics returns metrics fascade
func NewMetrics(ctx context.Context, output string, refreshRate time.Duration) Metrics {
	return Metrics{
		DaemonSupport:            utils.NewDaemonSupport(ctx),
		output:                   output,
		refreshRate:              refreshRate,
		createdTokens:            metrics.NewCounter(),
		deletedTokens:            metrics.NewCounter(),
		importedTransfers:        metrics.NewMeter(),
		importedTransactions:     metrics.NewMeter(),
		transactionSearchLatency: metrics.NewTimer(),
		transactionListLatency:   metrics.NewTimer(),
	}
}

// MarshalJSON serialises Metrics as json bytes
func (metrics *Metrics) MarshalJSON() ([]byte, error) {
	if metrics == nil {
		return nil, fmt.Errorf("cannot marshall nil")
	}

	if metrics.createdTokens == nil || metrics.deletedTokens == nil ||
		metrics.transactionSearchLatency == nil || metrics.transactionListLatency == nil ||
		metrics.importedTransfers == nil || metrics.importedTransactions == nil {
		return nil, fmt.Errorf("cannot marshall nil references")
	}

	var buffer bytes.Buffer

	buffer.WriteString("{\"createdTokens\":")
	buffer.WriteString(strconv.FormatInt(metrics.createdTokens.Count(), 10))
	buffer.WriteString(",\"deletedTokens\":")
	buffer.WriteString(strconv.FormatInt(metrics.deletedTokens.Count(), 10))
	buffer.WriteString(",\"transactionSearchLatency\":")
	buffer.WriteString(strconv.FormatFloat(metrics.transactionSearchLatency.Percentile(0.95), 'f', -1, 64))
	buffer.WriteString(",\"transactionListLatency\":")
	buffer.WriteString(strconv.FormatFloat(metrics.transactionListLatency.Percentile(0.95), 'f', -1, 64))
	buffer.WriteString(",\"importedTransfers\":")
	buffer.WriteString(strconv.FormatInt(metrics.importedTransfers.Count(), 10))
	buffer.WriteString(",\"importedTransactions\":")
	buffer.WriteString(strconv.FormatInt(metrics.importedTransactions.Count(), 10))
	buffer.WriteString("}")

	return buffer.Bytes(), nil
}

// UnmarshalJSON deserializes Metrics from json bytes
func (metrics *Metrics) UnmarshalJSON(data []byte) error {
	if metrics == nil {
		return fmt.Errorf("cannot unmarshall to nil")
	}

	if metrics.createdTokens == nil || metrics.deletedTokens == nil ||
		metrics.transactionSearchLatency == nil || metrics.transactionListLatency == nil ||
		metrics.importedTransfers == nil || metrics.importedTransactions == nil {
		return fmt.Errorf("cannot unmarshall to nil references")
	}

	aux := &struct {
		CreatedTokens            int64   `json:"createdTokens"`
		DeletedTokens            int64   `json:"deletedTokens"`
		ImportedTransfers        int64   `json:"importedTransfers"`
		ImportedTransactions     int64   `json:"importedTransactions"`
		TransactionSearchLatency float64 `json:"transactionSearchLatency"`
		TransactionListLatency   float64 `json:"transactionListLatency"`
	}{}

	if err := utils.JSON.Unmarshal(data, &aux); err != nil {
		return err
	}

	metrics.createdTokens.Clear()
	metrics.createdTokens.Inc(aux.CreatedTokens)
	metrics.deletedTokens.Clear()
	metrics.deletedTokens.Inc(aux.DeletedTokens)
	metrics.transactionSearchLatency.Update(time.Duration(aux.TransactionSearchLatency))
	metrics.transactionListLatency.Update(time.Duration(aux.TransactionListLatency))
	metrics.importedTransfers.Mark(aux.ImportedTransfers)
	metrics.importedTransactions.Mark(aux.ImportedTransactions)

	return nil
}