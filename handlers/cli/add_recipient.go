package cli

import (
	"fmt"
	"os"

	"github.com/go-git/go-git/v5"
	"github.com/prskr/git-age/core/ports"
	"github.com/prskr/git-age/core/services"
	"github.com/prskr/git-age/infrastructure"
	"github.com/urfave/cli/v2"
)

type AddRecipientCliHandler struct {
	RepoFS     ports.ReadWriteFS
	Recipients ports.Recipients
	Encryption ports.FileOpenSealer
	Repository ports.GitRepository
}

func (h *AddRecipientCliHandler) AddRecipient(ctx *cli.Context) (err error) {
	if ctx.NArg() != 1 {
		return cli.Exit("Expected exactly one argument", 1)
	}

	if isDirty, err := h.Repository.IsDirty(); err != nil {
		return fmt.Errorf("failed to check if repository is dirty: %w", err)
	} else if isDirty {
		return cli.Exit("Repository is dirty", 1)
	}

	if err := h.Recipients.Append(ctx.Args().First(), ctx.String("comment")); err != nil {
		return fmt.Errorf("failed to append public key to recipients file: %w", err)
	}

	if err := h.Repository.StageFile(ports.RecipientsFileName); err != nil {
		return fmt.Errorf("failed to add recipients file to git index: %w", err)
	}

	if err := h.Repository.WalkAgeFiles(ports.ReEncryptWalkFunc(h.Repository, h.RepoFS, h.Encryption)); err != nil {
		return err
	}

	if err := h.Repository.Commit(ctx.String("message")); err != nil {
		return fmt.Errorf("failed to commit changes: %w", err)
	}

	return nil
}

func (h *AddRecipientCliHandler) Command() *cli.Command {
	return &cli.Command{
		Name:   "add-recipient",
		Action: h.AddRecipient,
		Args:   true,
		Flags: []cli.Flag{
			&commentFlag,
			&cli.StringFlag{
				Name:    "message",
				Aliases: []string{"m"},
				Usage:   "Message to be used for the commit",
				Value:   "chore: add recipient",
			},
			&keysFlag,
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

			h.Recipients = infrastructure.NewRecipientsFile(h.RepoFS)

			h.Repository, err = infrastructure.NewGitRepository(h.RepoFS, gitRepo)
			if err != nil {
				return err
			}

			keysPath := context.String("keys")

			h.Encryption, err = services.NewAgeSealer(
				services.WithIdentitiesFrom(keysPath),
				services.WithRecipientsFrom(wd),
			)

			return err
		},
	}
}
