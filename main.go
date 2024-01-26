package main

import (
	"github.com/prskr/git-age/cmd"
	"log/slog"
	"os"
)

func main() {
	app := cmd.NewApp()

	if err := app.Run(); err != nil {
		slog.Error("Error occurred while running app", slog.String("err", err.Error()))
		os.Exit(1)
	}
}
