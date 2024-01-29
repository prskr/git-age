package cli

import (
	"errors"
	"filippo.io/age"
	"fmt"
	"github.com/adrg/xdg"
	"github.com/prskr/git-age/core/ports"
	"github.com/prskr/git-age/infrastructure"
	"github.com/urfave/cli/v2"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
)

type InitCliHandler struct {
	IdentitiesFile string
	RepoFS         ports.ReadWriteFS
}

func (h *InitCliHandler) Init(ctx *cli.Context) (err error) {
	if _, err := fs.Stat(h.RepoFS, ports.RecipientsFileName); err == nil {
		slog.Info("Repository already initialized")
		return nil
	}

	identity, err := age.GenerateX25519Identity()
	if err != nil {
		return fmt.Errorf("failed to generate identity: %w", err)
	}

	identitiesDir, _ := filepath.Split(h.IdentitiesFile)
	if err := os.MkdirAll(identitiesDir, 0700); err != nil {
		return fmt.Errorf("failed to create identities directory: %w", err)
	}

	identitiesFile, err := os.OpenFile(h.IdentitiesFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("failed to open identities file: %w", err)
	}

	defer func() {
		err = errors.Join(err, identitiesFile.Close())
	}()

	if comment := ctx.String("comment"); comment != "" {
		if _, err := identitiesFile.WriteString(fmt.Sprintf("# %s\n", comment)); err != nil {
			return fmt.Errorf("failed to write comment to identities file: %w", err)
		}
	}

	if _, err := identitiesFile.WriteString(fmt.Sprintf("# public key: %s\n", identity.Recipient().String())); err != nil {
		return fmt.Errorf("failed to write public key to identities file: %w", err)
	}

	if _, err := identitiesFile.WriteString(identity.String() + "\n"); err != nil {
		return fmt.Errorf("failed to write public key to identities file: %w", err)
	}

	recipientsFile, err := h.RepoFS.Create(ports.RecipientsFileName)
	if err != nil {
		return fmt.Errorf("failed to create recipients file: %w", err)
	}

	defer func() {
		err = errors.Join(err, recipientsFile.Close())
	}()

	if comment := ctx.String("comment"); comment != "" {
		if _, err := recipientsFile.WriteString(fmt.Sprintf("# %s\n", comment)); err != nil {
			return fmt.Errorf("failed to write comment to recipients file: %w", err)
		}
	}

	if _, err := recipientsFile.WriteString(identity.Recipient().String() + "\n"); err != nil {
		return fmt.Errorf("failed to write public key to recipients file: %w", err)
	}

	return nil
}

func (h *InitCliHandler) Command() *cli.Command {
	return &cli.Command{
		Name:   "init",
		Usage:  ``,
		Action: h.Init,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "comment",
				Aliases: []string{"c"},
				Usage:   "Comment for the recipient",
			},
			&cli.StringFlag{
				Name:        "keys",
				DefaultText: "By default keys are read from $XDG_CONFIG_HOME/git-age/keys.txt i.e. $HOME/.config/git-age/keys.txt on most systems",
				EnvVars: []string{
					"GIT_AGE_KEYS",
				},
			},
		},
		Before: func(context *cli.Context) error {
			keysPath := filepath.Join(xdg.ConfigHome, "git-age", "keys.txt")
			if flagPath := context.String("keys"); flagPath != "" {
				keysPath = flagPath
			}

			h.IdentitiesFile = keysPath

			wd, err := os.Getwd()
			if err != nil {
				return err
			}

			repoRootPath, err := repoRoot(wd)
			if err != nil {
				return err
			}

			h.RepoFS = infrastructure.NewReadWriteDirFS(repoRootPath)

			return nil
		},
	}
}
