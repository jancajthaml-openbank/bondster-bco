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

package metrics

import (
	"context"
	"time"

	"github.com/jancajthaml-openbank/bondster-bco-import/utils"
	localfs "github.com/jancajthaml-openbank/local-fs"
	metrics "github.com/rcrowley/go-metrics"
)

// Metrics holds metrics counters
type Metrics struct {
	utils.DaemonSupport
	storage                  localfs.Storage
	tenant                   string
	refreshRate              time.Duration
	createdTokens            metrics.Counter
	deletedTokens            metrics.Counter
	transactionSearchLatency metrics.Timer
	transactionListLatency   metrics.Timer
	importedTransfers        metrics.Meter
	importedTransactions     metrics.Meter
}

// NewMetrics returns blank metrics holder
func NewMetrics(ctx context.Context, output string, tenant string, refreshRate time.Duration) *Metrics {
	storage, err := localfs.NewPlaintextStorage(output)
	if err != nil {
		log.Error().Msgf("Failed to ensure storage %+v", err)
		return nil
	}
	return &Metrics{
		DaemonSupport:            utils.NewDaemonSupport(ctx, "metrics"),
		storage:                  storage,
		tenant:                   tenant,
		refreshRate:              refreshRate,
		createdTokens:            metrics.NewCounter(),
		deletedTokens:            metrics.NewCounter(),
		importedTransfers:        metrics.NewMeter(),
		importedTransactions:     metrics.NewMeter(),
		transactionSearchLatency: metrics.NewTimer(),
		transactionListLatency:   metrics.NewTimer(),
	}
}

// TokenCreated increments token created by one
func (metrics *Metrics) TokenCreated() {
	if metrics == nil {
		return
	}
	metrics.createdTokens.Inc(1)
}

// TokenDeleted increments token deleted by one
func (metrics *Metrics) TokenDeleted() {
	if metrics == nil {
		return
	}
	metrics.deletedTokens.Inc(1)
}

// TransfersImported increments transfers created by count
func (metrics *Metrics) TransfersImported(count int64) {
	if metrics == nil {
		return
	}
	metrics.importedTransfers.Mark(count)
}

// TransactionImported increments transactions created by count
func (metrics *Metrics) TransactionImported() {
	if metrics == nil {
		return
	}
	metrics.importedTransactions.Mark(1)
}

// TimeTransactionSearchLatency measures time of transaction search duration
func (metrics *Metrics) TimeTransactionSearchLatency(f func()) {
	if metrics == nil {
		return
	}
	metrics.transactionSearchLatency.Time(f)
}

// TimeTransactionListLatency measures time of transaction list duration
func (metrics *Metrics) TimeTransactionListLatency(f func()) {
	if metrics == nil {
		return
	}
	metrics.transactionListLatency.Time(f)
}

// Start handles everything needed to start metrics daemon
func (metrics *Metrics) Start() {
	if metrics == nil {
		return
	}

	ticker := time.NewTicker(metrics.refreshRate)
	defer ticker.Stop()

	if err := metrics.Hydrate(); err != nil {
		log.Warn().Msg(err.Error())
	}

	metrics.Persist()
	metrics.MarkReady()

	select {
	case <-metrics.CanStart:
		break
	case <-metrics.Done():
		metrics.MarkDone()
		return
	}

	log.Info().Msgf("Start metrics daemon, update file each %v", metrics.refreshRate)

	go func() {
		for {
			select {
			case <-metrics.Done():
				metrics.Persist()
				metrics.MarkDone()
				return
			case <-ticker.C:
				metrics.Persist()
			}
		}
	}()

	metrics.WaitStop()
	log.Info().Msg("Stop metrics daemon")
}
