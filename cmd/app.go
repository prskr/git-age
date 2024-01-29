package cmd

import (
	"context"
	"github.com/lmittmann/tint"
	clih "github.com/prskr/git-age/handlers/cli"
	"github.com/urfave/cli/v2"
	"log/slog"
	"os"
	"os/signal"
	"time"
)

func NewApp() *App {
	smudgeHandler := clih.SmudgeCliHandler{}
	cleanHandler := clih.CleanCliHandler{}
	filesHandler := clih.FilesCliHandler{}
	initHandler := clih.InitCliHandler{}
	genKeyHandler := clih.GenKeyCliHandler{}
	addRecipientHandler := clih.AddRecipientCliHandler{}
	installHandler := clih.InstallCliHandler{}

	a := &App{
		root: &cli.App{
			Name: "git-age",
			Usage: `
git-age is a Git filter to encrypt/decrypt files on push/pull operations.
`,
			Commands: []*cli.Command{
				smudgeHandler.Command(),
				cleanHandler.Command(),
				initHandler.Command(),
				genKeyHandler.Command(),
				addRecipientHandler.Command(),
				installHandler.Command(),
				filesHandler.Command(),
			},
		},
	}

	a.root.Before = a.setup

	return a
}

type App struct {
	root *cli.App
}

func (a *App) Run() error {
	ctx, _ := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)

	return a.root.RunContext(ctx, os.Args)
}

func (a *App) setup(*cli.Context) error {
	a.configureLogging()

	return nil
}

func (*App) configureLogging() {
	opts := &tint.Options{
		Level:      slog.LevelInfo,
		TimeFormat: time.RFC3339,
	}
	slog.SetDefault(slog.New(tint.NewHandler(os.Stderr, opts)))
}
