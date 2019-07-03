package metrics

import (
	"testing"
	"time"

	metrics "github.com/rcrowley/go-metrics"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMarshalJSON(t *testing.T) {

	t.Log("error when caller is nil")
	{
		var entity *Metrics
		_, err := entity.MarshalJSON()
		assert.EqualError(t, err, "cannot marshall nil")
	}

	t.Log("error when values are nil")
	{
		entity := Metrics{}
		_, err := entity.MarshalJSON()
		assert.EqualError(t, err, "cannot marshall nil references")
	}

	t.Log("happy path")
	{
		entity := Metrics{
			createdTokens:            metrics.NewCounter(),
			deletedTokens:            metrics.NewCounter(),
			importedTransfers:        metrics.NewMeter(),
			importedTransactions:     metrics.NewMeter(),
			transactionSearchLatency: metrics.NewTimer(),
			transactionListLatency:   metrics.NewTimer(),
		}

		entity.createdTokens.Inc(1)
		entity.deletedTokens.Inc(2)
		entity.transactionSearchLatency.Update(time.Duration(3))
		entity.transactionListLatency.Update(time.Duration(4))
		entity.importedTransfers.Mark(5)
		entity.importedTransactions.Mark(6)

		actual, err := entity.MarshalJSON()

		require.Nil(t, err)

		data := []byte("{\"createdTokens\":1,\"deletedTokens\":2,\"transactionSearchLatency\":3,\"transactionListLatency\":4,\"importedTransfers\":5,\"importedTransactions\":6}")

		assert.Equal(t, data, actual)
	}
}

func TestUnmarshalJSON(t *testing.T) {

	t.Log("error when caller is nil")
	{
		var entity *Metrics
		err := entity.UnmarshalJSON([]byte(""))
		assert.EqualError(t, err, "cannot unmarshall to nil")
	}

	t.Log("error when values are nil")
	{
		entity := Metrics{}
		err := entity.UnmarshalJSON([]byte(""))
		assert.EqualError(t, err, "cannot unmarshall to nil references")
	}

	t.Log("error on malformed data")
	{
		entity := Metrics{
			createdTokens:            metrics.NewCounter(),
			deletedTokens:            metrics.NewCounter(),
			importedTransfers:        metrics.NewMeter(),
			importedTransactions:     metrics.NewMeter(),
			transactionSearchLatency: metrics.NewTimer(),
			transactionListLatency:   metrics.NewTimer(),
		}

		data := []byte("{")
		assert.NotNil(t, entity.UnmarshalJSON(data))
	}

	t.Log("happy path")
	{
		entity := Metrics{
			createdTokens:            metrics.NewCounter(),
			deletedTokens:            metrics.NewCounter(),
			importedTransfers:        metrics.NewMeter(),
			importedTransactions:     metrics.NewMeter(),
			transactionSearchLatency: metrics.NewTimer(),
			transactionListLatency:   metrics.NewTimer(),
		}

		data := []byte("{\"createdTokens\":1,\"deletedTokens\":2,\"transactionSearchLatency\":3,\"transactionListLatency\":4,\"importedTransfers\":5,\"importedTransactions\":6}")
		require.Nil(t, entity.UnmarshalJSON(data))

		assert.Equal(t, int64(1), entity.createdTokens.Count())
		assert.Equal(t, int64(2), entity.deletedTokens.Count())
		assert.Equal(t, float64(3), entity.transactionSearchLatency.Percentile(0.95))
		assert.Equal(t, float64(4), entity.transactionListLatency.Percentile(0.95))
		assert.Equal(t, int64(5), entity.importedTransfers.Count())
		assert.Equal(t, int64(6), entity.importedTransactions.Count())
	}
}
