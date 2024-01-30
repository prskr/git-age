package cli

import (
	"io"
	"os"

	"github.com/prskr/git-age/core/ports"
	"github.com/prskr/git-age/core/services"
	"github.com/prskr/git-age/infrastructure"
	"github.com/urfave/cli/v2"
)

type SmudgeCliHandler struct {
	Opener ports.FileOpener
}

func (h *SmudgeCliHandler) SmudgeFile(*cli.Context) error {
	if err := requireStdin(); err != nil {
		return err
	}

	decryptedReader, err := h.Opener.OpenFile(os.Stdin)
	if err != nil {
		return err
	}

	_, err = io.Copy(os.Stdout, decryptedReader)

	return err
}

func (h *SmudgeCliHandler) Command() *cli.Command {
	return &cli.Command{
		Name:   "smudge",
		Usage:  "smudge should only be invoked by Git",
		Action: h.SmudgeFile,
		Args:   true,
		Hidden: true,
		Before: func(context *cli.Context) (err error) {
			h.Opener, err = services.NewAgeSealer(
				services.WithIdentities(infrastructure.NewIdentities(context.String("keys"))),
			)
			return err
		},
		Flags: []cli.Flag{
			&keysFlag,
		},
	}
}
