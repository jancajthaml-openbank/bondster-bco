package utils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSliceByMonths(t *testing.T) {
	start := time.Date(2000, time.January, 1, 2, 3, 4, 5, time.UTC)
	end := time.Date(2000, time.December, 10, 2, 3, 4, 5, time.UTC)

	timeRanges := SliceByMonths(start, end)

	actual := make([]string, 0)
	for _, timeRange := range timeRanges {
		actual = append(actual, timeRange.String())
	}

	expected := []string{
		"2000-01-01T02:03:04Z - 2000-01-31T00:00:00Z",
		"2000-02-01T00:00:00Z - 2000-02-29T00:00:00Z",
		"2000-03-01T00:00:00Z - 2000-03-31T00:00:00Z",
		"2000-04-01T00:00:00Z - 2000-04-30T00:00:00Z",
		"2000-05-01T00:00:00Z - 2000-05-31T00:00:00Z",
		"2000-06-01T00:00:00Z - 2000-06-30T00:00:00Z",
		"2000-07-01T00:00:00Z - 2000-07-31T00:00:00Z",
		"2000-08-01T00:00:00Z - 2000-08-31T00:00:00Z",
		"2000-09-01T00:00:00Z - 2000-09-30T00:00:00Z",
		"2000-10-01T00:00:00Z - 2000-10-31T00:00:00Z",
		"2000-11-01T00:00:00Z - 2000-11-30T00:00:00Z",
		"2000-12-01T00:00:00Z - 2000-12-10T02:03:04Z",
	}

	assert.Equal(t, expected, actual)
}

func TestPartitionInterval(t *testing.T) {
	start := time.Date(2000, time.January, 1, 2, 3, 4, 5, time.UTC)
	end := time.Date(2000, time.March, 5, 2, 3, 4, 5, time.UTC)

	timeRanges := PartitionInterval(start, end)

	actual := make([]string, 0)
	for _, timeRange := range timeRanges {
		actual = append(actual, timeRange.String())
	}

	expected := []string{
		"2000-01-01T02:03:04Z - 2000-01-31T00:00:00Z",
		"2000-02-01T00:00:00Z - 2000-02-29T00:00:00Z",
		"2000-03-01T00:00:00Z - 2000-03-05T02:03:04Z",
	}

	assert.Equal(t, expected, actual)
}
