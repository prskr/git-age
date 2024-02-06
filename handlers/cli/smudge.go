package cli

import (
	"bufio"
	"io"
	"log/slog"
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

	reader := bufio.NewReader(os.Stdin)

	if isEncrypted, err := h.Opener.IsEncrypted(reader); err != nil {
		return err
	} else if !isEncrypted {
		slog.Warn("expected age-encrypted file, but got plaintext. Copying to stdout.")
		_, err = io.Copy(os.Stdout, reader)
		return err
	}

	decryptedReader, err := h.Opener.OpenFile(reader)
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
