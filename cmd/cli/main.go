package main

import (
	"github.com/pgbytes/moneypenny/cmd/cli/root"
	"github.com/pgbytes/moneypenny/internal/log"
)

func main() {
	err := log.SetupLogging(log.NewLoggingConfig("info", "console"))
	if err != nil {
		panic(err)
	}

	log.GetLogger().Info("Welcome to MoneyPenny!")
	root.Execute()
}
