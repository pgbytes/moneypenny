// Package ynab provides functionality to transform domain transactions into YNAB CSV format.
package ynab

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pgbytes/moneypenny/internal/domain"
)

const (
	// ynabDateFormat is the date format expected by YNAB CSV imports (DD-MM-YYYY).
	ynabDateFormat = "02-01-2006"

	// outputSuffix is appended to input filename for the output file.
	outputSuffix = "_ynab"

	// outputExtension is the file extension for output files.
	outputExtension = ".csv"
)

// csvHeaders defines the column headers for YNAB CSV format.
var csvHeaders = []string{"Date", "Payee", "Memo", "Amount"}

// TransformResult contains information about the transformation operation.
type TransformResult struct {
	// OutputPath is the path where the CSV file was written.
	OutputPath string

	// TransactionCount is the number of transactions written.
	TransactionCount int
}

// TransformToCSV transforms a slice of domain transactions into YNAB CSV format
// and writes the result to the specified output file path.
//
// The CSV format follows YNAB import requirements:
//   - Header row: Date,Payee,Memo,Amount
//   - Date format: DD-MM-YYYY
//   - Amount: 2 decimal places, preserves sign (negative for outflows)
//
// Context is respected for cancellation during the operation.
func TransformToCSV(ctx context.Context, transactions []domain.Transaction, outputPath string) (*TransformResult, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	// Check context before starting
	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("transform cancelled: %w", ctx.Err())
	default:
	}

	// Create output file
	file, err := os.Create(outputPath)
	if err != nil {
		return nil, fmt.Errorf("creating output file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header row
	if err := writer.Write(csvHeaders); err != nil {
		return nil, fmt.Errorf("writing header row: %w", err)
	}

	// Write transaction rows
	for i, tx := range transactions {
		// Check context periodically
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("transform cancelled at row %d: %w", i+1, ctx.Err())
		default:
		}

		row := transactionToRow(tx)
		if err := writer.Write(row); err != nil {
			return nil, fmt.Errorf("writing row %d: %w", i+1, err)
		}
	}

	// Ensure all data is flushed
	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, fmt.Errorf("flushing csv writer: %w", err)
	}

	return &TransformResult{
		OutputPath:       outputPath,
		TransactionCount: len(transactions),
	}, nil
}

// transactionToRow converts a single domain transaction to a CSV row.
func transactionToRow(tx domain.Transaction) []string {
	return []string{
		tx.Date.Format(ynabDateFormat),
		tx.Payee,
		tx.Memo,
		formatAmount(tx.Amount),
	}
}

// formatAmount formats the amount with 2 decimal places.
func formatAmount(amount float64) string {
	return fmt.Sprintf("%.2f", amount)
}

// GenerateOutputPath creates the output file path based on the input file path.
// The output path is the input path with "_ynab.csv" suffix.
//
// Examples:
//   - "statement.csv" → "statement_ynab.csv"
//   - "statement" → "statement_ynab.csv"
//   - "/path/to/file.txt" → "/path/to/file_ynab.csv"
func GenerateOutputPath(inputPath string) string {
	dir := filepath.Dir(inputPath)
	base := filepath.Base(inputPath)

	// Remove extension if present
	ext := filepath.Ext(base)
	nameWithoutExt := strings.TrimSuffix(base, ext)

	// Create output filename
	outputName := nameWithoutExt + outputSuffix + outputExtension

	return filepath.Join(dir, outputName)
}
