// Package ynab provides the parent command for YNAB operations.
package ynab

import (
	"github.com/pgbytes/moneypenny/cmd/cli/ynab/transactions"
	"github.com/spf13/cobra"
)

// Cmd is the parent command for YNAB operations.
var Cmd = &cobra.Command{
	Use:   "ynab",
	Short: "YNAB budget management commands",
	Long: `Commands for interacting with YNAB (You Need A Budget) API.

These commands allow you to fetch and manage transactions in your YNAB budget.`,
}

func init() {
	// Register subcommands
	Cmd.AddCommand(transactions.Cmd)
}
