// Package fetch provides the command for fetching transactions from YNAB.
package fetch

import (
	"fmt"

	"github.com/pgbytes/moneypenny/internal/client/ynab"
	"github.com/pgbytes/moneypenny/internal/config"
	"github.com/pgbytes/moneypenny/internal/log"
	"github.com/spf13/cobra"
)

// Flags for the fetch command - isolated to this package.
var (
	accountID string
	limit     int
	sinceDate string
)

// Cmd fetches transactions from YNAB.
var Cmd = &cobra.Command{
	Use:   "fetch",
	Short: "Fetch transactions from YNAB",
	Long: `Fetch transactions from a specific YNAB account.

Example:
  mp ynab transactions fetch -f config.json -a account-id -n 20
  mp ynab transactions fetch -f config.json -a account-id --since-date 2026-01-01`,
	RunE: run,
}

func init() {
	// Add flags - no prefix needed since they're isolated to this package
	Cmd.Flags().StringVarP(&accountID, "account-id", "a", "", "account ID to fetch transactions from")
	Cmd.Flags().IntVarP(&limit, "limit", "n", 10, "number of transactions to fetch")
	Cmd.Flags().StringVarP(&sinceDate, "since-date", "s", "", "fetch transactions since date (ISO format: YYYY-MM-DD)")

	// Mark required flags
	_ = Cmd.MarkFlagRequired("account-id")
}

func run(cmd *cobra.Command, args []string) error {
	logger := log.GetLogger()

	// Get config path from parent's persistent flag
	configPath, err := cmd.Flags().GetString("config")
	if err != nil {
		return fmt.Errorf("getting config flag: %w", err)
	}

	// Load configuration
	cfg, err := config.LoadFromFile(configPath)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	// Validate YNAB config
	if err := cfg.YNAB.Validate(); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	// Create YNAB client
	client, err := ynab.NewClient(ynab.Config{
		APIKey:   cfg.YNAB.APIKey,
		BudgetID: cfg.YNAB.BudgetID,
	}, logger)
	if err != nil {
		return fmt.Errorf("creating YNAB client: %w", err)
	}

	// Build transaction options
	opts := ynab.TransactionOptions{
		SinceDate: sinceDate,
	}

	// Fetch transactions
	transactions, err := client.GetTransactionsByAccount(accountID, opts)
	if err != nil {
		return fmt.Errorf("fetching transactions: %w", err)
	}

	// Limit transactions if requested
	transactions = ynab.LimitTransactions(transactions, limit)

	// Display transactions
	logger.Infof("Fetched %d transactions from account %s", len(transactions), accountID)

	for _, t := range transactions {
		amount := ynab.MilliunitsToFloat(t.Amount)
		logger.Infof("  %s | %10.2f | %-30s | %s",
			t.Date,
			amount,
			truncateString(t.PayeeName, 30),
			truncateString(t.Memo, 40),
		)
	}

	return nil
}

// truncateString truncates a string to the specified length.
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}
