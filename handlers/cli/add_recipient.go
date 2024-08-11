package cli

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/alecthomas/kong"

	"github.com/prskr/git-age/core/dto"
	"github.com/prskr/git-age/core/ports"
	"github.com/prskr/git-age/core/services"
	"github.com/prskr/git-age/infrastructure"
)

type AddRecipientCliHandler struct {
	KeysFlag    `embed:""`
	CommentFlag `embed:""`
	Recipient   string `arg:"" help:"Recipient to add"`
	Message     string `help:"Message to be used for the commit" default:"chore: add recipient" short:"m"`
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

func (h *AddRecipientCliHandler) AfterApply(
	ctx context.Context,
	kongCtx *kong.Context,
	cwd ports.CWD,
	env ports.OSEnv,
) error {
	gitRepo, repoFS, err := infrastructure.NewGitRepositoryFromPath(cwd)
	if err != nil {
		return err
	}

	recipients := infrastructure.NewRecipientsFile(repoFS)

	idStore, err := infrastructure.IdentitiesStore(
		ctx,
		infrastructure.NewAgentIdentitiesStoreSource(env),
		infrastructure.NewFileIdentityStoreSource(h.Keys),
	)
	if err != nil {
		return fmt.Errorf("failed to init identities store: %w", err)
	}

	remotes, err := gitRepo.Remotes()
	if err != nil {
		return fmt.Errorf("failed to determine Git remotes: %w", err)
	}

	query := dto.IdentitiesQuery{
		Remotes: remotes,
	}

	ids, err := idStore.Identities(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to get identities: %w", err)
	}

	openSealer, err := services.NewAgeSealer(
		services.WithIdentities(ids...),
		services.WithRecipients(recipients),
	)
	if err != nil {
		return err
	}

	kongCtx.BindTo(repoFS, (*ports.ReadWriteFS)(nil))
	kongCtx.BindTo(gitRepo, (*ports.GitRepository)(nil))
	kongCtx.BindTo(recipients, (*ports.Recipients)(nil))
	kongCtx.BindTo(openSealer, (*ports.FileOpenSealer)(nil))

	return nil
}
