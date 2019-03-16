package daemon

import (
	"context"
	"testing"
	"time"

	"github.com/jancajthaml-openbank/bondster-bco-import/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetFilename(t *testing.T) {
	assert.Equal(t, "/a/b/c.import.d.e", getFilename("/a/b/c.e", "d"))
	assert.Equal(t, "/a/b/c.d", getFilename("/a/b/c.d", ""))
}

func TestMetricsPersist(t *testing.T) {
	cfg := config.Configuration{}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	entity := NewMetrics(ctx, cfg)
	delay := 1e8
	delta := 1e8

	t.Log("TimeTransactionSearchLatency properly times synchronization latency")
	{
		require.Equal(t, int64(0), entity.transactionSearchLatency.Count())
		entity.TimeTransactionSearchLatency(func() {
			time.Sleep(time.Duration(delay))
		})
		assert.Equal(t, int64(1), entity.transactionSearchLatency.Count())
		assert.InDelta(t, entity.transactionSearchLatency.Percentile(0.95), delay, delta)
	}

	t.Log("TimeTransactionListLatency properly times run of ImportAccount function")
	{
		require.Equal(t, int64(0), entity.transactionListLatency.Count())
		entity.TimeTransactionListLatency(func() {
			time.Sleep(time.Duration(delay))
		})
		assert.Equal(t, int64(1), entity.transactionListLatency.Count())
		assert.InDelta(t, entity.transactionListLatency.Percentile(0.95), delay, delta)
	}

	t.Log("TransactionImported properly marks number of accounts imported")
	{
		require.Equal(t, int64(0), entity.importedTransactions.Count())
		entity.TransactionImported()
		assert.Equal(t, int64(1), entity.importedTransactions.Count())
	}

	t.Log("TransfersImported properly marks number of accounts exported")
	{
		require.Equal(t, int64(0), entity.importedTransfers.Count())
		entity.TransfersImported(4)
		assert.Equal(t, int64(4), entity.importedTransfers.Count())
		entity.TransfersImported(6)
		assert.Equal(t, int64(10), entity.importedTransfers.Count())
	}

	t.Log("TokenCreated properly marks number of accounts imported")
	{
		require.Equal(t, int64(0), entity.createdTokens.Count())
		entity.TokenCreated()
		assert.Equal(t, int64(1), entity.createdTokens.Count())
	}

	t.Log("TokenDeleted properly marks number of accounts exported")
	{
		require.Equal(t, int64(0), entity.deletedTokens.Count())
		entity.TokenDeleted()
		assert.Equal(t, int64(1), entity.deletedTokens.Count())
	}
}
