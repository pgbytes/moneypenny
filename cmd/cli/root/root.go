package root

import (
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"os"
)

var (
	csvFilePath string
)

func init() {
	rootCmd.PersistentFlags().StringVarP(&csvFilePath, "csv", "c", "sample/sample.csv", "relative path to csv file to process")
}

var rootCmd = &cobra.Command{
	Use:   "mp",
	Short: "MoneyPenny is my finance assistant",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		// Exit as success if called with no arguments (same behaviour as
		// docker and other cobra based cli)
		if len(os.Args[1:]) == 0 {
			os.Exit(0)
		}
		handleError(err)
	}
}

func handleError(err error) {
	if err != nil {
		zap.S().Errorf("error while executing command: %v", err)
		os.Exit(1)
	}
	os.Exit(0)
}
