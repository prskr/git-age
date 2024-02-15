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

func (h *GenKeyCliHandler) Run() (err error) {
	pubKey, err := h.Identities.Generate(h.Comment)
	if err != nil {
		return fmt.Errorf("failed to generate identity: %w", err)
	}

	fmt.Println(pubKey)

	return nil
}

func (h *GenKeyCliHandler) AfterApply() error {
	h.Identities = infrastructure.NewIdentities(h.Keys)
	return nil
}
