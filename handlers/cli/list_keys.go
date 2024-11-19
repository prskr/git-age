package cli

import (
	"context"
	"fmt"
	"log/slog"
	"text/tabwriter"

	"filippo.io/age"

	"github.com/prskr/git-age/core/dto"
	"github.com/prskr/git-age/core/ports"
	"github.com/prskr/git-age/infrastructure"
)

type ListKeysCliHandler struct {
	KeysFlag `embed:""`

	Identities ports.IdentitiesStore `kong:"-"`
}

func (h *ListKeysCliHandler) Run(ctx context.Context, stdout ports.STDOUT) error {
	identities, err := h.Identities.Identities(ctx, dto.IdentitiesQuery{Remotes: []string{""}})
	if err != nil {
		return fmt.Errorf("failed to list identities: %w", err)
	}

	writer := tabwriter.NewWriter(stdout, 0, 0, 3, ' ', 0)

	_, _ = fmt.Fprintln(writer, "Public Key\t")

	for _, id := range identities {
		identity, ok := id.(*age.X25519Identity)
		if !ok {
			slog.Warn("uknown identity type", slog.String("type", fmt.Sprintf("%t", id)))
			continue
		}
		_, _ = fmt.Fprintf(writer, "%s\t\n", identity.Recipient().String())
	}

	return writer.Flush()
}

func (h *ListKeysCliHandler) AfterApply(ctx context.Context, env ports.OSEnv) error {
	idStore, err := infrastructure.IdentitiesStore(
		ctx,
		infrastructure.NewAgentIdentitiesStoreSource(env),
		infrastructure.NewFileIdentityStoreSource(h.Keys),
	)
	if err != nil {
		return fmt.Errorf("failed to init identities store: %w", err)
	}

	h.Identities = idStore
	return nil
}
