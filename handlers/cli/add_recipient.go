package cli

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/alecthomas/kong"

	"github.com/prskr/git-age/core/ports"
	"github.com/prskr/git-age/core/services"
	"github.com/prskr/git-age/infrastructure"
)

type AddRecipientCliHandler struct {
	KeysFlag    `embed:""`
	CommentFlag `embed:""`
	Recipient   string `arg:"" help:"Recipient to add"`
	Message     string `help:"Message to be used for the commit" default:"chore: add recipient" aliases:"m"`
}

func (h *AddRecipientCliHandler) Run(
	repoFS ports.ReadWriteFS,
	recipients ports.Recipients,
	openSealer ports.FileOpenSealer,
	repo ports.GitRepository,
) (err error) {
	if isDirty, err := repo.IsStagingDirty(); err != nil {
		return fmt.Errorf("failed to check if repository is dirty: %w", err)
	} else if isDirty {
		slog.Warn("Repository is dirty")
		os.Exit(1)
	}

	slog.Info("Adding recipient", slog.String("recipient", h.Recipient))
	appendedRecipients, err := recipients.Append(h.Recipient, h.Comment)
	if err != nil {
		return fmt.Errorf("failed to append public key to recipients file: %w", err)
	}
	openSealer.AddRecipients(appendedRecipients...)

	slog.Info("Staging recipients file")
	if err := repo.StageFile(ports.RecipientsFileName); err != nil {
		return fmt.Errorf("failed to add recipients file to git index: %w", err)
	}

	if err := repo.WalkAgeFiles(services.ReEncryptWalkFunc(repo, repoFS, openSealer)); err != nil {
		return err
	}

	slog.Info("Committing changes")
	if err := repo.Commit(h.Message); err != nil {
		return fmt.Errorf("failed to commit changes: %w", err)
	}

	return nil
}

func (h *AddRecipientCliHandler) AfterApply(kctx *kong.Context) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	repository, repoFS, err := infrastructure.NewGitRepositoryFromPath(wd)
	if err != nil {
		return err
	}

	recipients := infrastructure.NewRecipientsFile(repoFS)

	openSealer, err := services.NewAgeSealer(
		services.WithIdentities(infrastructure.NewIdentities(h.Keys)),
		services.WithRecipients(recipients),
	)
	if err != nil {
		return err
	}

	kctx.BindTo(repoFS, (*ports.ReadWriteFS)(nil))
	kctx.BindTo(repository, (*ports.GitRepository)(nil))
	kctx.BindTo(recipients, (*ports.Recipients)(nil))
	kctx.BindTo(openSealer, (*ports.FileOpenSealer)(nil))

	return nil
}
