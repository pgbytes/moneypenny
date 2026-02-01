// Package milesmore provides a parser for Miles & More credit card CSV statements.
package milesmore

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/pgbytes/moneypenny/internal/domain"
)

const (
	// Expected column indices based on CSV format.
	colVoucherDate      = 0
	colReceiptDate      = 1
	colPayee            = 2
	colForeignCurrency  = 3
	colForeignAmount    = 4
	colExchangeRate     = 5
	colAmount           = 6
	colCurrency         = 7
	expectedColumnCount = 8

	// Date format used in the CSV: "1/29/2026".
	csvDateFormat = "1/2/2006"

	// Foreign transaction fee identifier.
	feeIdentifier = "AUSLANDSEINSATZENTGELT"
)

// ParseResult contains the parsed transactions, any non-fatal errors encountered,
// and summary information.
type ParseResult struct {
	// Transactions contains all successfully parsed transactions.
	Transactions []domain.Transaction

	// Errors contains non-fatal parsing errors for individual rows.
	Errors []ParseError

	// TotalRows is the total number of data rows processed (excluding headers).
	TotalRows int

	// SuccessfulRows is the number of successfully parsed rows.
	SuccessfulRows int
}

// ParseError represents a non-fatal error encountered while parsing a specific row.
type ParseError struct {
	// Line is the line number in the source file.
	Line int

	// Row is the raw CSV row data.
	Row []string

	// Error is the error encountered.
	Error error
}

// Parse reads a Miles & More credit card CSV statement and returns domain transactions.
// The parser is lenient: it skips invalid rows and collects errors for reporting.
//
// CSV Format:
//   - First 3-4 lines contain metadata (skipped)
//   - Transaction rows have 8 columns separated by semicolons
//   - Columns: Voucher date, Receipt date, Payee, Foreign currency, Foreign amount,
//     Exchange rate, Amount (EUR), Currency
//
// Context is respected for cancellation during long-running parses.
func Parse(ctx context.Context, reader io.Reader, sourceFile string) (*ParseResult, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	csvReader := csv.NewReader(reader)
	csvReader.Comma = ';'
	csvReader.LazyQuotes = true
	csvReader.TrimLeadingSpace = true
	csvReader.FieldsPerRecord = -1 // Allow variable column count for header rows

	result := &ParseResult{
		Transactions: make([]domain.Transaction, 0),
		Errors:       make([]ParseError, 0),
	}

	lineNumber := 0
	headerSkipped := false
	occurrenceMap := make(map[string]int)       // Track occurrences for import ID
	var previousTransaction *domain.Transaction // Track for fee association

	for {
		// Check context cancellation
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("parsing cancelled: %w", ctx.Err())
		default:
		}

		record, err := csvReader.Read()
		lineNumber++

		if err == io.EOF {
			break
		}

		if err != nil {
			// CSV parsing error - record and continue
			result.Errors = append(result.Errors, ParseError{
				Line:  lineNumber,
				Row:   record,
				Error: fmt.Errorf("csv read error: %w", err),
			})
			continue
		}

		// Skip metadata header rows (first 4 lines)
		if !headerSkipped {
			if lineNumber <= 4 {
				continue
			}
			// Check if this is the column header row
			if len(record) > 0 && strings.Contains(record[0], "Voucher date") {
				headerSkipped = true
				continue
			}
			// First data row
			headerSkipped = true
		}

		// Skip empty rows
		if len(record) == 0 || (len(record) == 1 && strings.TrimSpace(record[0]) == "") {
			continue
		}

		// Skip balance row (last line in Miles & More statements)
		if len(record) > 0 && strings.HasPrefix(strings.TrimSpace(record[0]), "Balance:") {
			continue
		}

		// Validate column count
		if len(record) < expectedColumnCount {
			result.Errors = append(result.Errors, ParseError{
				Line:  lineNumber,
				Row:   record,
				Error: fmt.Errorf("expected %d columns, got %d", expectedColumnCount, len(record)),
			})
			result.TotalRows++
			continue
		}

		// Parse the transaction
		transaction, err := parseTransaction(record, lineNumber, sourceFile)
		if err != nil {
			result.Errors = append(result.Errors, ParseError{
				Line:  lineNumber,
				Row:   record,
				Error: err,
			})
			result.TotalRows++
			continue
		}

		// Check if this is a foreign transaction fee
		if strings.Contains(transaction.Payee, feeIdentifier) && previousTransaction != nil {
			// Associate fee with previous foreign transaction if applicable
			if previousTransaction.ForeignCurrency != "" && previousTransaction.ForeignAmount != 0 {
				transaction.Memo = fmt.Sprintf("Fee for transaction: %s", previousTransaction.Payee)
			}
		}

		// Generate import ID
		transaction.ImportID = generateImportID(transaction, occurrenceMap)

		result.Transactions = append(result.Transactions, *transaction)
		result.TotalRows++
		result.SuccessfulRows++
		previousTransaction = transaction
	}

	return result, nil
}

// parseTransaction parses a single CSV row into a domain.Transaction.
func parseTransaction(record []string, lineNumber int, sourceFile string) (*domain.Transaction, error) {
	transaction := &domain.Transaction{
		SourceFile: sourceFile,
		SourceLine: lineNumber,
		Currency:   "EUR", // Default settlement currency
	}

	// Parse voucher date (primary transaction date)
	voucherDate, err := parseDate(strings.TrimSpace(record[colVoucherDate]))
	if err != nil {
		return nil, fmt.Errorf("invalid voucher date: %w", err)
	}
	transaction.Date = voucherDate

	// Parse receipt date (posting date)
	receiptDate, err := parseDate(strings.TrimSpace(record[colReceiptDate]))
	if err != nil {
		return nil, fmt.Errorf("invalid receipt date: %w", err)
	}
	transaction.PostingDate = receiptDate

	// Parse payee
	transaction.Payee = strings.TrimSpace(record[colPayee])
	if transaction.Payee == "" {
		return nil, fmt.Errorf("payee is required")
	}

	// Parse amount (EUR) - required
	amount, err := parseAmount(strings.TrimSpace(record[colAmount]))
	if err != nil {
		return nil, fmt.Errorf("invalid amount: %w", err)
	}
	transaction.Amount = amount

	// Parse currency
	currency := strings.TrimSpace(record[colCurrency])
	if currency != "" {
		transaction.Currency = currency
	}

	// Parse foreign currency fields (optional)
	foreignCurrency := strings.TrimSpace(record[colForeignCurrency])
	if foreignCurrency != "" && foreignCurrency != "EUR" {
		transaction.ForeignCurrency = foreignCurrency

		// Parse foreign amount
		foreignAmount, err := parseAmount(strings.TrimSpace(record[colForeignAmount]))
		if err == nil {
			transaction.ForeignAmount = foreignAmount
		}

		// Parse exchange rate
		exchangeRate, err := parseExchangeRate(strings.TrimSpace(record[colExchangeRate]))
		if err == nil && exchangeRate > 0 {
			transaction.ExchangeRate = exchangeRate
		}
	}

	return transaction, nil
}

// parseDate parses a date string in the format "1/29/2026".
func parseDate(dateStr string) (time.Time, error) {
	if dateStr == "" {
		return time.Time{}, fmt.Errorf("date is empty")
	}

	t, err := time.Parse(csvDateFormat, dateStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid date format (expected M/D/YYYY): %w", err)
	}

	return t, nil
}

// parseAmount parses an amount string, handling European number format (comma as decimal).
// Examples: "-330", "-8.44", "-0.16"
func parseAmount(amountStr string) (float64, error) {
	if amountStr == "" {
		return 0, fmt.Errorf("amount is empty")
	}

	// Remove any whitespace
	amountStr = strings.TrimSpace(amountStr)

	// Parse as float
	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid number format: %w", err)
	}

	return amount, nil
}

// parseExchangeRate parses an exchange rate string.
func parseExchangeRate(rateStr string) (float64, error) {
	if rateStr == "" {
		return 0, nil
	}

	rate, err := strconv.ParseFloat(strings.TrimSpace(rateStr), 64)
	if err != nil {
		return 0, fmt.Errorf("invalid exchange rate: %w", err)
	}

	return rate, nil
}

// generateImportID generates a YNAB-compatible import ID.
// Format: "YNAB:[milliunit_amount]:[iso_date]:[occurrence]"
// Example: "YNAB:-294230:2015-12-30:1"
func generateImportID(t *domain.Transaction, occurrenceMap map[string]int) string {
	// Convert amount to milliunits (multiply by 1000)
	milliunits := int64(t.Amount * 1000)

	// Format date as ISO (YYYY-MM-DD)
	isoDate := t.Date.Format("2006-01-02")

	// Create base key for occurrence tracking
	baseKey := fmt.Sprintf("%d:%s", milliunits, isoDate)

	// Increment occurrence counter
	occurrenceMap[baseKey]++
	occurrence := occurrenceMap[baseKey]

	return fmt.Sprintf("YNAB:%d:%s:%d", milliunits, isoDate, occurrence)
}
