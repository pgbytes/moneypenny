// Package transactions provides the parent command for transaction operations.
package transactions

import (
	"github.com/pgbytes/moneypenny/cmd/cli/ynab/transactions/fetch"
	"github.com/spf13/cobra"
)

// Cmd is the parent command for transaction operations.
var Cmd = &cobra.Command{
	Use:   "transactions",
	Short: "Transaction management commands",
	Long:  `Commands for fetching and uploading transactions.`,
}

func init() {
	// Register subcommands
	Cmd.AddCommand(fetch.Cmd)
}
