package ynab

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMilliunitsToFloat(t *testing.T) {
	tests := []struct {
		name       string
		milliunits int64
		expected   float64
	}{
		{name: "positive amount", milliunits: 123930, expected: 123.93},
		{name: "negative amount", milliunits: -294230, expected: -294.23},
		{name: "zero", milliunits: 0, expected: 0.0},
		{name: "small positive", milliunits: 220, expected: 0.22},
		{name: "small negative", milliunits: -220, expected: -0.22},
		{name: "large amount", milliunits: 4924340, expected: 4924.34},
		{name: "exact dollar", milliunits: 100000, expected: 100.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			result := MilliunitsToFloat(tt.milliunits)

			// Assert
			assert.InDelta(t, tt.expected, result, 0.001)
		})
	}
}

func TestFloatToMilliunits(t *testing.T) {
	tests := []struct {
		name     string
		amount   float64
		expected int64
	}{
		{name: "positive amount", amount: 123.93, expected: 123930},
		{name: "negative amount", amount: -294.23, expected: -294230},
		{name: "zero", amount: 0.0, expected: 0},
		{name: "small positive", amount: 0.22, expected: 220},
		{name: "small negative", amount: -0.22, expected: -220},
		{name: "large amount", amount: 4924.34, expected: 4924340},
		{name: "exact dollar", amount: 100.0, expected: 100000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			result := FloatToMilliunits(tt.amount)

			// Assert
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGenerateImportID(t *testing.T) {
	tests := []struct {
		name       string
		amount     int64
		date       time.Time
		occurrence int
		expected   string
	}{
		{
			name:       "typical negative amount",
			amount:     -294230,
			date:       time.Date(2015, 12, 30, 0, 0, 0, 0, time.UTC),
			occurrence: 1,
			expected:   "YNAB:-294230:2015-12-30:1",
		},
		{
			name:       "positive amount (income)",
			amount:     500000,
			date:       time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC),
			occurrence: 1,
			expected:   "YNAB:500000:2026-01-15:1",
		},
		{
			name:       "second occurrence same day",
			amount:     -294230,
			date:       time.Date(2015, 12, 30, 0, 0, 0, 0, time.UTC),
			occurrence: 2,
			expected:   "YNAB:-294230:2015-12-30:2",
		},
		{
			name:       "zero amount",
			amount:     0,
			date:       time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
			occurrence: 1,
			expected:   "YNAB:0:2026-06-01:1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			result := GenerateImportID(tt.amount, tt.date, tt.occurrence)

			// Assert
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestLimitTransactions(t *testing.T) {
	// Create sample transactions
	transactions := []Transaction{
		{ID: "tx-1", Date: "2026-01-01"},
		{ID: "tx-2", Date: "2026-01-02"},
		{ID: "tx-3", Date: "2026-01-03"},
		{ID: "tx-4", Date: "2026-01-04"},
		{ID: "tx-5", Date: "2026-01-05"},
	}

	tests := []struct {
		name        string
		limit       int
		expectedLen int
		expectedIDs []string
	}{
		{
			name:        "limit less than total",
			limit:       3,
			expectedLen: 3,
			expectedIDs: []string{"tx-3", "tx-4", "tx-5"},
		},
		{
			name:        "limit equals total",
			limit:       5,
			expectedLen: 5,
			expectedIDs: []string{"tx-1", "tx-2", "tx-3", "tx-4", "tx-5"},
		},
		{
			name:        "limit greater than total",
			limit:       10,
			expectedLen: 5,
			expectedIDs: []string{"tx-1", "tx-2", "tx-3", "tx-4", "tx-5"},
		},
		{
			name:        "limit of 1",
			limit:       1,
			expectedLen: 1,
			expectedIDs: []string{"tx-5"},
		},
		{
			name:        "limit of 0 returns all",
			limit:       0,
			expectedLen: 5,
			expectedIDs: []string{"tx-1", "tx-2", "tx-3", "tx-4", "tx-5"},
		},
		{
			name:        "negative limit returns all",
			limit:       -1,
			expectedLen: 5,
			expectedIDs: []string{"tx-1", "tx-2", "tx-3", "tx-4", "tx-5"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			result := LimitTransactions(transactions, tt.limit)

			// Assert
			assert.Len(t, result, tt.expectedLen)
			for i, expectedID := range tt.expectedIDs {
				assert.Equal(t, expectedID, result[i].ID)
			}
		})
	}
}

func TestClearedStatus_Constants(t *testing.T) {
	// Assert cleared status constants are correctly defined
	assert.Equal(t, ClearedStatus("cleared"), ClearedStatusCleared)
	assert.Equal(t, ClearedStatus("uncleared"), ClearedStatusUncleared)
	assert.Equal(t, ClearedStatus("reconciled"), ClearedStatusReconciled)
}
