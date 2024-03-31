package cli

import (
	"context"
	"fmt"

	"github.com/prskr/git-age/core/dto"
	"github.com/prskr/git-age/core/ports"
	"github.com/prskr/git-age/infrastructure"
)

type GenKeyCliHandler struct {
	KeysFlag    `embed:""`
	CommentFlag `embed:""`
	RemoteFlag  `embed:""`

	Identities ports.IdentitiesStore `kong:"-"`
}

func (h *GenKeyCliHandler) Run(ctx context.Context, stdout ports.STDOUT) (err error) {
	cmd := dto.GenerateIdentityCommand{
		Comment: h.Comment,
		Remote:  h.Remote,
	}

	pubKey, err := h.Identities.Generate(ctx, cmd)
	if err != nil {
		return fmt.Errorf("failed to generate identity: %w", err)
	}

	_, err = fmt.Fprintln(stdout, pubKey)

	return err
}

func (h *GenKeyCliHandler) AfterApply(ctx context.Context) error {
	idStore, err := infrastructure.IdentitiesStore(
		ctx,
		infrastructure.NewAgentIdentitiesStoreSource(),
		infrastructure.NewFileIdentityStoreSource(h.Keys),
	)
	if err != nil {
		return fmt.Errorf("failed to init identities store: %w", err)
	}

	h.Identities = idStore
	return nil
}
