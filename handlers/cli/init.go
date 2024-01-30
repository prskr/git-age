package cli

import (
	"fmt"
	"io/fs"
	"log/slog"
	"os"

	"github.com/prskr/git-age/core/ports"
	"github.com/prskr/git-age/infrastructure"
	"github.com/urfave/cli/v2"
)

type InitCliHandler struct {
	Identities ports.Identities
	Recipients ports.Recipients
	RepoFS     ports.ReadWriteFS
}

func (h *InitCliHandler) Init(ctx *cli.Context) (err error) {
	if _, err := fs.Stat(h.RepoFS, ports.RecipientsFileName); err == nil {
		slog.Info("Repository already initialized")
		return nil
	}

	pubKey, err := h.Identities.Generate(ctx.String("comment"))
	if err != nil {
		return fmt.Errorf("failed to generate identity: %w", err)
	}

	if err := h.Recipients.Append(pubKey, ctx.String("comment")); err != nil {
		return fmt.Errorf("failed to append recipient: %w", err)
	}

	return nil
}

func (h *InitCliHandler) Command() *cli.Command {
	return &cli.Command{
		Name:   "init",
		Usage:  ``,
		Action: h.Init,
		Flags: []cli.Flag{
			&commentFlag,
			&keysFlag,
		},
		Before: func(context *cli.Context) error {
			h.Identities = infrastructure.NewIdentities(context.String("keys"))

			wd, err := os.Getwd()
			if err != nil {
				return err
			}

			repoRootPath, err := repoRoot(wd)
			if err != nil {
				return err
			}

			h.RepoFS = infrastructure.NewReadWriteDirFS(repoRootPath)
			h.Recipients = infrastructure.NewRecipientsFile(h.RepoFS)

			return nil
		},
	}
}
