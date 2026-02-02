// Package milesmore provides the command for transforming Miles & More statements to YNAB format.
package milesmore

import (
	"context"
	"fmt"
	"os"

	"github.com/pgbytes/moneypenny/internal/log"
	"github.com/pgbytes/moneypenny/internal/parsers/milesmore"
	"github.com/pgbytes/moneypenny/internal/transform/ynab"
	"github.com/spf13/cobra"
)

// Flags for the milesmore command - isolated to this package.
var inputPath string

// Cmd transforms Miles & More statements to YNAB format.
var Cmd = &cobra.Command{
	Use:   "milesmore",
	Short: "Transform Miles & More statement to YNAB format",
	Long: `Transform a Miles & More credit card CSV statement to YNAB-compatible CSV format.

This command reads a Miles & More statement CSV file, parses all transactions,
and creates a new CSV file in YNAB import format at the same location with "_ynab" suffix.

The transformation is strict: if any parsing errors occur, the process aborts.

Example:
  mp ynab transform milesmore -i /path/to/statement.csv

Output will be created at: /path/to/statement_ynab.csv`,
	RunE: run,
}

func init() {
	Cmd.Flags().StringVarP(&inputPath, "input", "i", "", "path to Miles & More CSV statement file")

	_ = Cmd.MarkFlagRequired("input")
}

func run(cmd *cobra.Command, args []string) error {
	logger := log.GetLogger()
	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	logger.Infof("Starting Miles & More to YNAB transformation")
	logger.Debugf("Input file: %s", inputPath)

	// Validate input file exists
	if _, err := os.Stat(inputPath); os.IsNotExist(err) {
		return fmt.Errorf("input file not found: %s", inputPath)
	}

	// Open input file for parsing
	inputFile, err := os.Open(inputPath)
	if err != nil {
		return fmt.Errorf("opening input file: %w", err)
	}
	defer inputFile.Close()

	// Parse Miles & More CSV
	logger.Infof("Parsing Miles & More statement...")
	parseResult, err := milesmore.Parse(ctx, inputFile, inputPath)
	if err != nil {
		return fmt.Errorf("parsing Miles & More CSV: %w", err)
	}

	// Strict mode: abort if any parsing errors occurred
	if len(parseResult.Errors) > 0 {
		logger.Errorf("Parsing encountered %d errors (strict mode - aborting):", len(parseResult.Errors))
		for _, parseErr := range parseResult.Errors {
			logger.Errorf("  Line %d: %v", parseErr.Line, parseErr.Error)
		}
		return fmt.Errorf("parsing failed with %d errors, aborting transformation", len(parseResult.Errors))
	}

	logger.Infof("Successfully parsed %d transactions from %d rows",
		parseResult.SuccessfulRows, parseResult.TotalRows)

	// Check if there are any transactions to transform
	if len(parseResult.Transactions) == 0 {
		logger.Warnf("No transactions found in input file")
		return fmt.Errorf("no transactions to transform")
	}

	// Generate output path
	outputPath := ynab.GenerateOutputPath(inputPath)
	logger.Debugf("Output file: %s", outputPath)

	// Transform to YNAB format
	logger.Infof("Transforming to YNAB format...")
	transformResult, err := ynab.TransformToCSV(ctx, parseResult.Transactions, outputPath)
	if err != nil {
		return fmt.Errorf("transforming to YNAB format: %w", err)
	}

	logger.Infof("Transformation complete!")
	logger.Infof("  Transactions written: %d", transformResult.TransactionCount)
	logger.Infof("  Output file: %s", transformResult.OutputPath)

	return nil
}
