// Package budgets provides the parent command for budget operations.
package budgets

import (
	"github.com/pgbytes/moneypenny/cmd/cli/ynab/budgets/fetch"
	"github.com/spf13/cobra"
)

// Cmd is the parent command for budget operations.
var Cmd = &cobra.Command{
	Use:   "budgets",
	Short: "Budget management commands",
	Long:  `Commands for listing and managing YNAB budgets.`,
}

func init() {
	// Register subcommands
	Cmd.AddCommand(fetch.Cmd)
}
