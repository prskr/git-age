package cli

import (
	"bufio"
	"fmt"
	"io"
	"log/slog"

	"github.com/prskr/git-age/core/ports"
	"github.com/prskr/git-age/core/services"
	"github.com/prskr/git-age/infrastructure"
)

type SmudgeCliHandler struct {
	KeysFlag        `embed:""`
	Opener          ports.FileOpener `kong:"-"`
	FileToCleanPath string           `arg:"" name:"file" help:"Path to the file to clean"`
}

func (h *SmudgeCliHandler) Run(stdin ports.STDIN, stdout ports.STDOUT) error {
	if err := requireStdin(stdin); err != nil {
		return err
	}

	reader := bufio.NewReader(stdin)

	if isEncrypted, err := h.Opener.IsEncrypted(reader); err != nil {
		return err
	} else if !isEncrypted {
		slog.Warn("expected age-encrypted file, but got plaintext. Copying to stdout.")
		_, err = io.Copy(stdout, reader)
		return err
	}

	decryptedReader, err := h.Opener.OpenFile(reader)
	if err != nil {
		return err
	}

	_, err = io.Copy(stdout, decryptedReader)

	return err
}

func (h *SmudgeCliHandler) AfterApply() (err error) {
	keysStore, err := infrastructure.KeysStoreFor(h.Keys)
	if err != nil {
		return fmt.Errorf("failed to create keys reader: %w", err)
	}

	h.Opener, err = services.NewAgeSealer(services.WithIdentities(infrastructure.NewIdentities(keysStore)))
	return err
}
