package cli

import (
	"errors"
	"filippo.io/age"
	"fmt"
	"github.com/adrg/xdg"
	"github.com/go-git/go-git/v5"
	"github.com/prskr/git-age/core/ports"
	"github.com/prskr/git-age/core/services"
	"github.com/prskr/git-age/infrastructure"
	"github.com/urfave/cli/v2"
	"os"
	"path/filepath"
	"strings"
)

type AddRecipientCliHandler struct {
	RepoFS     ports.ReadWriteFS
	Encryption ports.FileOpenSealer
	Repository ports.GitRepository
}

func (h *AddRecipientCliHandler) AddRecipient(ctx *cli.Context) (err error) {
	if ctx.NArg() != 1 {
		return cli.Exit("Expected exactly one argument", 1)
	}

	pubKey := ctx.Args().First()
	if _, err := age.ParseRecipients(strings.NewReader(pubKey)); err != nil {
		return fmt.Errorf("failed to parse public key from argument: %w", err)
	}

	if clean, err := h.Repository.IsDirty(); err != nil {
		return fmt.Errorf("failed to check if repository is dirty: %w", err)
	} else if !clean {
		return cli.Exit("Repository is dirty", 1)
	}

	f, err := h.RepoFS.Append(ports.RecipientsFileName)
	if err != nil {
		return fmt.Errorf("failed to open recipients file: %w", err)
	}

	defer func() {
		err = errors.Join(err, f.Close())
	}()

	if comment := ctx.String("comment"); comment != "" {
		if _, err := f.WriteString(fmt.Sprintf("# %s\n", comment)); err != nil {
			return fmt.Errorf("failed to write comment to recipients file: %w", err)
		}
	}

	if _, err := f.WriteString(pubKey + "\n"); err != nil {
		return fmt.Errorf("failed to write public key to recipients file: %w", err)
	}

	if err := h.Repository.WalkAgeFiles(ports.ReEncryptWalkFunc(h.Repository, h.RepoFS, h.Encryption)); err != nil {
		return err
	}

	return nil
}

func (h *AddRecipientCliHandler) Command() *cli.Command {
	return &cli.Command{
		Name:   "add-recipient",
		Action: h.AddRecipient,
		Args:   true,
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
			wd, err := os.Getwd()
			if err != nil {
				return err
			}

			repoRootPath, err := repoRoot(wd)
			if err != nil {
				return err
			}

			gitRepo, err := git.PlainOpen(repoRootPath)
			if err != nil {
				return err
			}

			h.RepoFS = infrastructure.NewReadWriteDirFS(repoRootPath)

			h.Repository, err = infrastructure.NewGitRepository(h.RepoFS, gitRepo)
			if err != nil {
				return err
			}

			keysPath := filepath.Join(xdg.ConfigHome, "git-age", "keys.txt")
			if flagPath := context.String("keys"); flagPath != "" {
				keysPath = flagPath
			}

			h.Encryption, err = services.NewAgeSealer(
				services.WithIdentitiesFrom(keysPath),
				services.WithRecipientsFrom(wd),
			)

			return err
		},
	}
}
