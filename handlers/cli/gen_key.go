package cli

import (
	"fmt"

	"github.com/prskr/git-age/core/ports"
	"github.com/prskr/git-age/infrastructure"
	"github.com/urfave/cli/v2"
)

type GenKeyCliHandler struct {
	Identities ports.Identities
}

// GenKey generates a new key
// adds it to the keys.txt file
// and prints the public key to STDOUT for sharing
func (h *GenKeyCliHandler) GenKey(ctx *cli.Context) (err error) {
	pubKey, err := h.Identities.Generate(ctx.String("comment"))
	if err != nil {
		return fmt.Errorf("failed to generate identity: %w", err)
	}

	fmt.Println(pubKey)

	return nil
}

func (h *GenKeyCliHandler) Command() *cli.Command {
	return &cli.Command{
		Name:   "gen-key",
		Action: h.GenKey,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "comment",
				Aliases: []string{"c"},
				Usage:   "Comment for the recipient",
			},
			&keysFlag,
		},
		Before: func(context *cli.Context) error {
			h.Identities = infrastructure.NewIdentities(context.String("keys"))
			return nil
		},
	}
}
