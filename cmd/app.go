package cmd

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"time"

	"github.com/prskr/git-age/core/ports"

	"github.com/adrg/xdg"
	"github.com/alecthomas/kong"

	"github.com/lmittmann/tint"

	clih "github.com/prskr/git-age/handlers/cli"
)

type App struct {
	Logging struct {
		Level slog.Level `env:"GIT_AGE_LOG_LEVEL" help:"Log level" default:"warn"`
	} `embed:""`

	Clean        clih.CleanCliHandler        `cmd:"" hidden:"" help:"clean should only be invoked by Git"`
	Smudge       clih.SmudgeCliHandler       `cmd:"" hidden:"" help:"smudge should only be invoked by Git"`
	Files        clih.FilesCliHandler        `cmd:"" help:"Interact with repo files"`
	AddRecipient clih.AddRecipientCliHandler `cmd:"" help:"Add a recipient to the list of recipients"`
	GenKey       clih.GenKeyCliHandler       `cmd:"" help:"Generate a new key pair"`
	Init         clih.InitCliHandler         `cmd:"" help:"Initialize a repository"`
	Install      clih.InstallCliHandler      `cmd:"" help:"Install git-age hooks in global git config"`
	Version      clih.VersionCliHandler      `cmd:"" help:"Print version information"`
}

func (a *App) Execute() error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	ctx, _ := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	cliCtx := kong.Parse(a,
		kong.Name("git-age"),
		kong.BindTo(ctx, (*context.Context)(nil)),
		kong.BindTo(os.Stdout, (*ports.STDOUT)(nil)),
		kong.BindTo(os.Stdin, (*ports.STDIN)(nil)),
		kong.Bind(ports.CWD(wd)),
		kong.Vars{
			"XDG_CONFIG_HOME":     xdg.ConfigHome,
			"file_path_separator": string(filepath.Separator),
		})

	return cliCtx.Run()
}

func (a *App) AfterApply() error {
	opts := &tint.Options{
		Level:      a.Logging.Level,
		TimeFormat: time.RFC3339,
	}
	slog.SetDefault(slog.New(tint.NewHandler(os.Stderr, opts)))

	return nil
}
