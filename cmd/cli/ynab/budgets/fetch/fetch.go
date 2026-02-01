// Package fetch provides the command for fetching budgets from YNAB.
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
	includeAccounts bool
	verbose         bool
)

// Cmd fetches budgets from YNAB.
var Cmd = &cobra.Command{
	Use:   "fetch",
	Short: "Fetch budgets from YNAB",
	Long: `Fetch all budgets for the authenticated YNAB user.

Example:
  mp ynab budgets fetch -f config.json
  mp ynab budgets fetch -f config.json --include-accounts
  mp ynab budgets fetch -f config.json --include-accounts --verbose`,
	RunE: run,
}

func init() {
	// Add flags - no prefix needed since they're isolated to this package
	Cmd.Flags().BoolVarP(&includeAccounts, "include-accounts", "a", false, "include accounts in output")
	Cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "display all details (default: short format)")
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

	// Validate API key is present
	if cfg.YNAB.APIKey == "" {
		return fmt.Errorf("invalid config: api_key is required")
	}

	// Create YNAB client - budget_id not required for listing budgets
	client, err := ynab.NewClient(ynab.Config{
		APIKey:   cfg.YNAB.APIKey,
		BudgetID: cfg.YNAB.BudgetID,
	}, logger)
	if err != nil {
		return fmt.Errorf("creating YNAB client: %w", err)
	}

	// Fetch budgets
	budgets, err := client.GetBudgets(includeAccounts)
	if err != nil {
		return fmt.Errorf("fetching budgets: %w", err)
	}

	// Display budgets
	logger.Infof("Fetched %d budgets", len(budgets))

	for _, b := range budgets {
		if verbose {
			printBudgetVerbose(logger, b, includeAccounts)
		} else {
			printBudgetShort(logger, b, includeAccounts)
		}
	}

	return nil
}

// printBudgetShort prints budget in short format (id, name, last_modified_on).
func printBudgetShort(logger log.Logger, b ynab.BudgetSummary, showAccounts bool) {
	logger.Infof("  Budget: %s | %s | Last Modified: %s", b.ID, b.Name, b.LastModifiedOn)

	if showAccounts && len(b.Accounts) > 0 {
		for _, acc := range b.Accounts {
			logger.Infof("    Account: %s | %s", acc.ID, acc.Name)
		}
	}
}

// printBudgetVerbose prints budget with all details.
func printBudgetVerbose(logger log.Logger, b ynab.BudgetSummary, showAccounts bool) {
	logger.Infof("  Budget:")
	logger.Infof("    ID:            %s", b.ID)
	logger.Infof("    Name:          %s", b.Name)
	logger.Infof("    Last Modified: %s", b.LastModifiedOn)
	logger.Infof("    First Month:   %s", b.FirstMonth)
	logger.Infof("    Last Month:    %s", b.LastMonth)

	if b.DateFormat != nil {
		logger.Infof("    Date Format:   %s", b.DateFormat.Format)
	}

	if b.CurrencyFormat != nil {
		logger.Infof("    Currency:      %s (%s)", b.CurrencyFormat.CurrencySymbol, b.CurrencyFormat.ISOCode)
	}

	if showAccounts && len(b.Accounts) > 0 {
		logger.Infof("    Accounts (%d):", len(b.Accounts))
		for _, acc := range b.Accounts {
			printAccountVerbose(logger, acc)
		}
	}
}

// printAccountVerbose prints account with all details.
func printAccountVerbose(logger log.Logger, acc ynab.Account) {
	logger.Infof("      Account:")
	logger.Infof("        ID:       %s", acc.ID)
	logger.Infof("        Name:     %s", acc.Name)
	logger.Infof("        Type:     %s", acc.Type)
	logger.Infof("        On Budget: %t", acc.OnBudget)
	logger.Infof("        Closed:   %t", acc.Closed)
	logger.Infof("        Balance:  %.2f", ynab.MilliunitsToFloat(acc.Balance))

	if acc.Note != "" {
		logger.Infof("        Note:     %s", acc.Note)
	}
}
