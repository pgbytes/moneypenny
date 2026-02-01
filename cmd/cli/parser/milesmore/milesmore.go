// Package milesmore provides the command for parsing Miles & More credit card statements.
package milesmore

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"github.com/pgbytes/moneypenny/internal/log"
	"github.com/pgbytes/moneypenny/internal/parsers/milesmore"
	"github.com/spf13/cobra"
)

var (
	filePath string
	verbose  bool
)

// Cmd parses a Miles & More credit card CSV statement.
var Cmd = &cobra.Command{
	Use:   "milesmore",
	Short: "Parse Miles & More credit card CSV statement",
	Long: `Parse a Miles & More credit card statement from a CSV file.

This command validates the CSV format, parses all transactions, and displays them
in a formatted table. Any parsing errors are reported at the end.

Example:
  mp parser milesmore --file statement.csv
  mp parser milesmore -f statement.csv --verbose`,
	RunE: run,
}

func init() {
	Cmd.Flags().StringVarP(&filePath, "file", "f", "", "path to CSV file")
	Cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "show detailed output")
	_ = Cmd.MarkFlagRequired("file")
}

func run(cmd *cobra.Command, args []string) error {
	logger := log.GetLogger()

	if err := validateFilePath(filePath); err != nil {
		return err
	}

	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("opening file: %w", err)
	}
	defer file.Close()

	logger.Infof("Parsing Miles & More statement: %s", filePath)
	ctx := context.Background()
	result, err := milesmore.Parse(ctx, file, filepath.Base(filePath))
	if err != nil {
		return fmt.Errorf("parsing CSV: %w", err)
	}

	displayTable(result, logger)

	if len(result.Errors) > 0 {
		displayErrors(result, logger)
	}

	return nil
}

func validateFilePath(path string) error {
	if path == "" {
		return fmt.Errorf("file path is required")
	}

	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("file does not exist: %s", path)
		}
		return fmt.Errorf("checking file: %w", err)
	}

	if info.IsDir() {
		return fmt.Errorf("path is a directory, not a file: %s", path)
	}

	ext := strings.ToLower(filepath.Ext(path))
	if ext != ".csv" {
		return fmt.Errorf("file must have .csv extension, got: %s", ext)
	}

	if info.Size() == 0 {
		return fmt.Errorf("file is empty: %s", path)
	}

	return nil
}

func displayTable(result *milesmore.ParseResult, logger log.Logger) {
	if len(result.Transactions) == 0 {
		logger.Warn("No transactions found in CSV")
		return
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	fmt.Fprintln(w, "\nDATE\tPAYEE\tAMOUNT (EUR)")
	fmt.Fprintln(w, strings.Repeat("═", 12)+"\t"+strings.Repeat("═", 45)+"\t"+strings.Repeat("═", 12))

	totalAmount := 0.0
	for _, tx := range result.Transactions {
		dateStr := tx.Date.Format("2006-01-02")
		payee := truncateString(tx.Payee, 45)
		amountStr := fmt.Sprintf("%.2f", tx.Amount)
		fmt.Fprintf(w, "%s\t%s\t%s\n", dateStr, payee, amountStr)
		totalAmount += tx.Amount
	}

	fmt.Fprintln(w, strings.Repeat("─", 12)+"\t"+strings.Repeat("─", 45)+"\t"+strings.Repeat("─", 12))

	fmt.Fprintf(w, "\nSummary:\n")
	fmt.Fprintf(w, "  Total Transactions:\t%d\n", len(result.Transactions))
	fmt.Fprintf(w, "  Total Amount:\t%.2f EUR\n", totalAmount)

	if len(result.Transactions) > 0 {
		firstDate := result.Transactions[len(result.Transactions)-1].Date.Format("2006-01-02")
		lastDate := result.Transactions[0].Date.Format("2006-01-02")
		fmt.Fprintf(w, "  Date Range:\t%s to %s\n", firstDate, lastDate)
	}

	fmt.Fprintf(w, "  Parsing Errors:\t%d\n", len(result.Errors))
	w.Flush()

	if verbose && len(result.Transactions) > 0 {
		displayVerboseDetails(result)
	}
}

func displayVerboseDetails(result *milesmore.ParseResult) {
	fmt.Println("\n" + strings.Repeat("═", 80))
	fmt.Println("VERBOSE TRANSACTION DETAILS")
	fmt.Println(strings.Repeat("═", 80))

	for i, tx := range result.Transactions {
		fmt.Printf("\nTransaction #%d:\n", i+1)
		fmt.Printf("  Date:           %s\n", tx.Date.Format("2006-01-02"))
		fmt.Printf("  Posting Date:   %s\n", tx.PostingDate.Format("2006-01-02"))
		fmt.Printf("  Payee:          %s\n", tx.Payee)
		fmt.Printf("  Amount:         %.2f %s\n", tx.Amount, tx.Currency)

		if tx.ForeignCurrency != "" {
			fmt.Printf("  Foreign Amount: %.2f %s\n", tx.ForeignAmount, tx.ForeignCurrency)
			fmt.Printf("  Exchange Rate:  %.5f\n", tx.ExchangeRate)
		}

		if tx.Memo != "" {
			fmt.Printf("  Memo:           %s\n", tx.Memo)
		}

		fmt.Printf("  Import ID:      %s\n", tx.ImportID)
	}

	fmt.Println()
}

func displayErrors(result *milesmore.ParseResult, logger log.Logger) {
	fmt.Println("\n" + strings.Repeat("═", 80))
	fmt.Printf("PARSING ERRORS (%d)\n", len(result.Errors))
	fmt.Println(strings.Repeat("═", 80))

	for i, parseErr := range result.Errors {
		fmt.Printf("\nError #%d (Line %d):\n", i+1, parseErr.Line)
		fmt.Printf("  Error:   %s\n", parseErr.Error.Error())
		if len(parseErr.Row) > 0 {
			fmt.Printf("  Raw Row: %s\n", strings.Join(parseErr.Row, " | "))
		}
	}

	fmt.Println()
	logger.Warnf("Found %d parsing errors. Please review the CSV file.", len(result.Errors))
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}
