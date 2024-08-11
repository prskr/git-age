package cli

import (
	"context"
	"fmt"
	"io/fs"
	"log/slog"

	"github.com/prskr/git-age/core/dto"
	"github.com/prskr/git-age/core/ports"
	"github.com/prskr/git-age/infrastructure"
)

type InitCliHandler struct {
	CommentFlag `embed:""`
	KeysFlag    `embed:""`
	RemoteFlag  `embed:""`

	Identities ports.IdentitiesStore `kong:"-"`
	Recipients ports.Recipients      `kong:"-"`
	RepoFS     ports.ReadWriteFS     `kong:"-"`
}

func (h *InitCliHandler) Run(ctx context.Context) (err error) {
	if _, err := fs.Stat(h.RepoFS, ports.RecipientsFileName); err == nil {
		slog.Info("Repository already initialized")
		return nil
	}

	cmd := dto.GenerateIdentityCommand{
		Comment: h.Comment,
		Remote:  h.Remote,
	}

	pubKey, err := h.Identities.Generate(ctx, cmd)
	if err != nil {
		return fmt.Errorf("failed to generate identity: %w", err)
	}

	if _, err := h.Recipients.Append(pubKey, h.Comment); err != nil {
		return fmt.Errorf("failed to append recipient: %w", err)
	}

	return nil
}

func (h *InitCliHandler) AfterApply(ctx context.Context, cwd ports.CWD, env ports.OSEnv) error {
	idStore, err := infrastructure.IdentitiesStore(
		ctx,
		infrastructure.NewAgentIdentitiesStoreSource(env),
		infrastructure.NewFileIdentityStoreSource(h.Keys),
	)
	if err != nil {
		return fmt.Errorf("failed to init identities store: %w", err)
	}

	h.Identities = idStore

	repoRootPath, err := infrastructure.FindRepoRootFrom(cwd)
	if err != nil {
		return err
	}

	h.RepoFS = infrastructure.NewReadWriteDirFS(repoRootPath)
	h.Recipients = infrastructure.NewRecipientsFile(h.RepoFS)

	return nil
}
