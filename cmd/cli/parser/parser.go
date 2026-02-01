// Package parser provides commands for parsing financial statements from various sources.
package parser

import (
	"github.com/pgbytes/moneypenny/cmd/cli/parser/milesmore"
	"github.com/spf13/cobra"
)

// Cmd is the parent command for parser operations.
var Cmd = &cobra.Command{
	Use:   "parser",
	Short: "Parse financial statements from various sources",
	Long: `Commands for parsing financial statements from banks and credit card providers.

Supported formats:
  - milesmore: Miles & More credit card statements (CSV)

These commands validate and display parsed transactions before importing to external services.`,
}

func init() {
	// Register subcommands
	Cmd.AddCommand(milesmore.Cmd)
}
