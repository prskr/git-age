package main

import (
	"log/slog"
	"os"

	"github.com/prskr/git-age/cmd"
)

func main() {
	app := new(cmd.App)

	if err := app.Execute(); err != nil {
		slog.Error("Error occurred while running app", slog.String("err", err.Error()))
		os.Exit(1)
	}
}
