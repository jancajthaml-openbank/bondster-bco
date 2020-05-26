package utils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetMonthsWithin(t *testing.T) {
	start := time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2001, time.January, 1, 0, 0, 0, 0, time.UTC).Add(time.Nanosecond*-1)

	months := GetMonthsWithin(start, end)

	actual := make([]string, 0)
	for _, month := range months {
		actual = append(actual, month.Format(time.RFC3339))
	}

	expected := []string{
		"2000-01-01T00:00:00Z",
		"2000-02-01T00:00:00Z",
		"2000-03-01T00:00:00Z",
		"2000-04-01T00:00:00Z",
		"2000-05-01T00:00:00Z",
		"2000-06-01T00:00:00Z",
		"2000-07-01T00:00:00Z",
		"2000-08-01T00:00:00Z",
		"2000-09-01T00:00:00Z",
		"2000-10-01T00:00:00Z",
		"2000-11-01T00:00:00Z",
		"2000-12-01T00:00:00Z",
	}

	assert.Equal(t, expected, actual)
}
