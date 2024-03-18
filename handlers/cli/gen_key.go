package cli

import (
	"fmt"

	"github.com/prskr/git-age/core/ports"
	"github.com/prskr/git-age/infrastructure"
)

type GenKeyCliHandler struct {
	KeysFlag    `embed:""`
	CommentFlag `embed:""`

	Identities ports.Identities `kong:"-"`
}

func (h *GenKeyCliHandler) Run(stdout ports.STDOUT) (err error) {
	pubKey, err := h.Identities.Generate(h.Comment)
	if err != nil {
		return fmt.Errorf("failed to generate identity: %w", err)
	}

	_, err = fmt.Fprintln(stdout, pubKey)

	return err
}

func (h *GenKeyCliHandler) AfterApply() error {
	keysStore, err := infrastructure.KeysStoreFor(h.Keys)
	if err != nil {
		return fmt.Errorf("failed to create keys reader: %w", err)
	}

	h.Identities = infrastructure.NewIdentities(keysStore)
	return nil
}
