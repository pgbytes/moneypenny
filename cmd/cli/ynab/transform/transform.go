// Package transform provides the parent command for YNAB transformation operations.
package transform

import (
	"github.com/pgbytes/moneypenny/cmd/cli/ynab/transform/milesmore"
	"github.com/spf13/cobra"
)

// Cmd is the parent command for YNAB transformation operations.
var Cmd = &cobra.Command{
	Use:   "transform",
	Short: "Transform bank statements to YNAB format",
	Long: `Commands for transforming various bank statement formats to YNAB-compatible CSV.

These commands read bank-specific CSV exports and convert them to a format
that can be imported directly into YNAB (You Need A Budget).

Output format: Date,Payee,Memo,Amount
Date format: DD-MM-YYYY`,
}

func init() {
	// Register subcommands
	Cmd.AddCommand(milesmore.Cmd)
}
