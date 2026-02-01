package milesmore

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/pgbytes/moneypenny/internal/domain"
	"github.com/stretchr/testify/suite"
)

// ParserTestSuite groups all parser tests.
type ParserTestSuite struct {
	suite.Suite
}

func TestParserTestSuite(t *testing.T) {
	suite.Run(t, new(ParserTestSuite))
}

// TestParse_WithValidCSV_ParsesAllTransactions tests successful parsing of valid CSV.
func (s *ParserTestSuite) TestParse_WithValidCSV_ParsesAllTransactions() {
	// Arrange
	csvPath := filepath.Join("testdata", "valid.csv")
	file, err := os.Open(csvPath)
	s.Require().NoError(err)
	defer file.Close()

	ctx := context.Background()

	// Act
	result, err := Parse(ctx, file, "valid.csv")

	// Assert
	s.NoError(err)
	s.NotNil(result)
	s.Equal(6, result.SuccessfulRows, "should parse all 6 transactions")
	s.Equal(6, len(result.Transactions))
	s.Empty(result.Errors, "should have no parsing errors")

	// Verify first transaction (foreign transaction fee)
	firstTx := result.Transactions[0]
	s.Equal("AUSLANDSEINSATZENTGELT", firstTx.Payee)
	s.Equal(-0.16, firstTx.Amount)
	s.Equal("EUR", firstTx.Currency)
	s.Equal("", firstTx.ForeignCurrency)
	s.Equal(0.0, firstTx.ForeignAmount)
	// Note: Fee comes before foreign transaction in CSV, so no association
	s.Equal("", firstTx.Memo)

	// Verify second transaction (foreign currency)
	secondTx := result.Transactions[1]
	s.Equal("RECALL, 19709 MIDDLETOWN, DE, USA", secondTx.Payee)
	s.Equal(-8.44, secondTx.Amount)
	s.Equal("EUR", secondTx.Currency)
	s.Equal("USD", secondTx.ForeignCurrency)
	s.Equal(-10.0, secondTx.ForeignAmount)
	s.Equal(1.18483, secondTx.ExchangeRate)
	s.Equal("valid.csv", secondTx.SourceFile)
	s.Greater(secondTx.SourceLine, 0)

	// Verify dates (second transaction dates)
	expectedVoucherDate := time.Date(2026, 1, 28, 0, 0, 0, 0, time.UTC)
	expectedPostingDate := time.Date(2026, 1, 29, 0, 0, 0, 0, time.UTC)
	s.Equal(expectedVoucherDate, secondTx.Date)
	s.Equal(expectedPostingDate, secondTx.PostingDate)

	// Verify third transaction (domestic EUR)
	thirdTx := result.Transactions[2]
	s.Equal("PAYPAL *rafaublacha, 10715 35314369001, DEU, DEU", thirdTx.Payee)
	s.Equal(-330.0, thirdTx.Amount)
	s.Equal("EUR", thirdTx.Currency)
	s.Equal("", thirdTx.ForeignCurrency)
}

// TestParse_WithInvalidRows_CollectsErrors tests lenient parsing with errors.
func (s *ParserTestSuite) TestParse_WithInvalidRows_CollectsErrors() {
	// Arrange
	csvPath := filepath.Join("testdata", "invalid_rows.csv")
	file, err := os.Open(csvPath)
	s.Require().NoError(err)
	defer file.Close()

	ctx := context.Background()

	// Act
	result, err := Parse(ctx, file, "invalid_rows.csv")

	// Assert
	s.NoError(err, "parser should not fail on invalid rows")
	s.NotNil(result)
	s.Equal(1, result.SuccessfulRows, "only 1 valid transaction")
	s.Equal(1, len(result.Transactions))
	s.NotEmpty(result.Errors, "should collect errors")
	s.Greater(len(result.Errors), 0)

	// Verify valid transaction was parsed
	validTx := result.Transactions[0]
	s.Equal("Valid Transaction", validTx.Payee)
	s.Equal(-10.50, validTx.Amount)

	// Verify errors were collected
	hasDateError := false
	hasMissingDataError := false
	hasEmptyPayeeError := false

	for _, parseErr := range result.Errors {
		if strings.Contains(parseErr.Error.Error(), "invalid voucher date") {
			hasDateError = true
		}
		if strings.Contains(parseErr.Error.Error(), "expected") {
			hasMissingDataError = true
		}
		if strings.Contains(parseErr.Error.Error(), "payee is required") {
			hasEmptyPayeeError = true
		}
	}

	s.True(hasDateError, "should detect invalid date format")
	s.True(hasMissingDataError || hasEmptyPayeeError, "should detect missing required fields")
}

// TestParse_WithEmptyCSV_ReturnsEmptyResult tests parsing of empty CSV.
func (s *ParserTestSuite) TestParse_WithEmptyCSV_ReturnsEmptyResult() {
	// Arrange
	csvPath := filepath.Join("testdata", "empty.csv")
	file, err := os.Open(csvPath)
	s.Require().NoError(err)
	defer file.Close()

	ctx := context.Background()

	// Act
	result, err := Parse(ctx, file, "empty.csv")

	// Assert
	s.NoError(err)
	s.NotNil(result)
	s.Equal(0, result.SuccessfulRows)
	s.Empty(result.Transactions)
}

// TestParse_WithBalanceLine_SkipsBalanceLine tests that balance lines are skipped.
func (s *ParserTestSuite) TestParse_WithBalanceLine_SkipsBalanceLine() {
	// Arrange
	csvPath := filepath.Join("testdata", "with_balance.csv")
	file, err := os.Open(csvPath)
	s.Require().NoError(err)
	defer file.Close()

	ctx := context.Background()

	// Act
	result, err := Parse(ctx, file, "with_balance.csv")

	// Assert
	s.NoError(err)
	s.NotNil(result)
	s.Equal(2, result.SuccessfulRows, "should parse 2 transactions")
	s.Equal(2, len(result.Transactions))
	s.Empty(result.Errors, "balance line should be skipped, not treated as error")

	// Verify transactions
	firstTx := result.Transactions[0]
	s.Equal("Test Transaction 1", firstTx.Payee)
	s.Equal(-10.50, firstTx.Amount)

	secondTx := result.Transactions[1]
	s.Equal("Test Transaction 2", secondTx.Payee)
	s.Equal(-20.00, secondTx.Amount)
}

// TestParse_WithCancelledContext_ReturnsError tests context cancellation.
func (s *ParserTestSuite) TestParse_WithCancelledContext_ReturnsError() {
	// Arrange
	csvContent := strings.NewReader(`Credit card transactions
Credit card;Customer number;Card number;Card holder
Miles & More Gold Credit Card;123;5426****1495;TEST
Billing date: 2/3/2026
Voucher date;Date of receipt;Reason for payment;Foreign currency;Amount;Exchange rate;Amount;Currency
1/29/2026;1/29/2026;Test Transaction;EUR;-10;1.00000;-10;EUR`)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// Act
	result, err := Parse(ctx, csvContent, "test.csv")

	// Assert
	s.Error(err)
	s.Nil(result)
	s.Contains(err.Error(), "cancelled")
}

// TestGenerateImportID_WithSameAmountAndDate_IncrementsOccurrence tests import ID generation.
func (s *ParserTestSuite) TestGenerateImportID_WithSameAmountAndDate_IncrementsOccurrence() {
	// Arrange
	occurrenceMap := make(map[string]int)
	date := time.Date(2026, 1, 29, 0, 0, 0, 0, time.UTC)

	tx1 := &domain.Transaction{
		Date:   date,
		Amount: -10.50,
	}

	tx2 := &domain.Transaction{
		Date:   date,
		Amount: -10.50,
	}

	// Act
	id1 := generateImportID(tx1, occurrenceMap)
	id2 := generateImportID(tx2, occurrenceMap)

	// Assert
	s.Equal("YNAB:-10500:2026-01-29:1", id1)
	s.Equal("YNAB:-10500:2026-01-29:2", id2, "second transaction should increment occurrence")
}

// TestGenerateImportID_WithDifferentAmounts_UsesDifferentKeys tests import ID uniqueness.
func (s *ParserTestSuite) TestGenerateImportID_WithDifferentAmounts_UsesDifferentKeys() {
	// Arrange
	occurrenceMap := make(map[string]int)
	date := time.Date(2026, 1, 29, 0, 0, 0, 0, time.UTC)

	tx1 := &domain.Transaction{
		Date:   date,
		Amount: -10.50,
	}

	tx2 := &domain.Transaction{
		Date:   date,
		Amount: -20.75,
	}

	// Act
	id1 := generateImportID(tx1, occurrenceMap)
	id2 := generateImportID(tx2, occurrenceMap)

	// Assert
	s.Equal("YNAB:-10500:2026-01-29:1", id1)
	s.Equal("YNAB:-20750:2026-01-29:1", id2, "different amounts should have different import IDs")
}

// TestParseDate_WithValidFormats_ParsesCorrectly tests date parsing.
func (s *ParserTestSuite) TestParseDate_WithValidFormats_ParsesCorrectly() {
	tests := []struct {
		name     string
		input    string
		expected time.Time
		wantErr  bool
	}{
		{
			name:     "single digit month and day",
			input:    "1/5/2026",
			expected: time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC),
			wantErr:  false,
		},
		{
			name:     "double digit month and day",
			input:    "12/25/2026",
			expected: time.Date(2026, 12, 25, 0, 0, 0, 0, time.UTC),
			wantErr:  false,
		},
		{
			name:     "empty string",
			input:    "",
			expected: time.Time{},
			wantErr:  true,
		},
		{
			name:     "invalid format",
			input:    "2026-01-29",
			expected: time.Time{},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			// Act
			result, err := parseDate(tt.input)

			// Assert
			if tt.wantErr {
				s.Error(err)
			} else {
				s.NoError(err)
				s.Equal(tt.expected, result)
			}
		})
	}
}

// TestParseAmount_WithVariousFormats_ParsesCorrectly tests amount parsing.
func (s *ParserTestSuite) TestParseAmount_WithVariousFormats_ParsesCorrectly() {
	tests := []struct {
		name     string
		input    string
		expected float64
		wantErr  bool
	}{
		{
			name:     "negative integer",
			input:    "-330",
			expected: -330.0,
			wantErr:  false,
		},
		{
			name:     "negative decimal",
			input:    "-8.44",
			expected: -8.44,
			wantErr:  false,
		},
		{
			name:     "small decimal",
			input:    "-0.16",
			expected: -0.16,
			wantErr:  false,
		},
		{
			name:     "positive amount",
			input:    "100.50",
			expected: 100.50,
			wantErr:  false,
		},
		{
			name:     "with whitespace",
			input:    "  -25.75  ",
			expected: -25.75,
			wantErr:  false,
		},
		{
			name:     "empty string",
			input:    "",
			expected: 0,
			wantErr:  true,
		},
		{
			name:     "non-numeric",
			input:    "abc",
			expected: 0,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			// Act
			result, err := parseAmount(tt.input)

			// Assert
			if tt.wantErr {
				s.Error(err)
			} else {
				s.NoError(err)
				s.Equal(tt.expected, result)
			}
		})
	}
}

// TestParseExchangeRate_WithVariousInputs_ParsesCorrectly tests exchange rate parsing.
func (s *ParserTestSuite) TestParseExchangeRate_WithVariousInputs_ParsesCorrectly() {
	tests := []struct {
		name     string
		input    string
		expected float64
		wantErr  bool
	}{
		{
			name:     "valid rate",
			input:    "1.18483",
			expected: 1.18483,
			wantErr:  false,
		},
		{
			name:     "one-to-one rate",
			input:    "1.00000",
			expected: 1.00000,
			wantErr:  false,
		},
		{
			name:     "zero rate",
			input:    "0.00000",
			expected: 0.0,
			wantErr:  false,
		},
		{
			name:     "empty string",
			input:    "",
			expected: 0.0,
			wantErr:  false,
		},
		{
			name:     "invalid format",
			input:    "invalid",
			expected: 0.0,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			// Act
			result, err := parseExchangeRate(tt.input)

			// Assert
			if tt.wantErr {
				s.Error(err)
			} else {
				s.NoError(err)
				s.Equal(tt.expected, result)
			}
		})
	}
}
