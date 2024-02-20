package cli

import (
	"fmt"
	"io/fs"
	"log/slog"

	"github.com/prskr/git-age/core/ports"
	"github.com/prskr/git-age/infrastructure"
)

type InitCliHandler struct {
	CommentFlag `embed:""`
	KeysFlag    `embed:""`

	Identities ports.Identities  `kong:"-"`
	Recipients ports.Recipients  `kong:"-"`
	RepoFS     ports.ReadWriteFS `kong:"-"`
}

func (h *InitCliHandler) Run() (err error) {
	if _, err := fs.Stat(h.RepoFS, ports.RecipientsFileName); err == nil {
		slog.Info("Repository already initialized")
		return nil
	}

	pubKey, err := h.Identities.Generate(h.Comment)
	if err != nil {
		return fmt.Errorf("failed to generate identity: %w", err)
	}

	if _, err := h.Recipients.Append(pubKey, h.Comment); err != nil {
		return fmt.Errorf("failed to append recipient: %w", err)
	}

	return nil
}

func (h *InitCliHandler) AfterApply(cwd ports.CWD) error {
	h.Identities = infrastructure.NewIdentities(h.Keys)

	repoRootPath, err := infrastructure.FindRepoRootFrom(cwd)
	if err != nil {
		return err
	}

	h.RepoFS = infrastructure.NewReadWriteDirFS(repoRootPath)
	h.Recipients = infrastructure.NewRecipientsFile(h.RepoFS)

	return nil
}
