// Package ynab provides the parent command for YNAB operations.
package ynab

import (
	"github.com/pgbytes/moneypenny/cmd/cli/ynab/budgets"
	"github.com/pgbytes/moneypenny/cmd/cli/ynab/transactions"
	"github.com/pgbytes/moneypenny/cmd/cli/ynab/transform"
	"github.com/spf13/cobra"
)

// configPath holds the path to the config file for all YNAB subcommands.
var configPath string

// Cmd is the parent command for YNAB operations.
var Cmd = &cobra.Command{
	Use:   "ynab",
	Short: "YNAB budget management commands",
	Long: `Commands for interacting with YNAB (You Need A Budget) API.

These commands allow you to fetch and manage budgets and transactions in your YNAB account.`,
}

func init() {
	// Add persistent flags available to all subcommands
	Cmd.PersistentFlags().StringVarP(&configPath, "config", "f", "", "path to config file (JSON)")
	_ = Cmd.MarkPersistentFlagRequired("config")

	// Register subcommands
	Cmd.AddCommand(budgets.Cmd)
	Cmd.AddCommand(transactions.Cmd)
	Cmd.AddCommand(transform.Cmd)
}
