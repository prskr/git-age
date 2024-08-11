package cli

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log/slog"

	"github.com/prskr/git-age/core/dto"
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

func (h *SmudgeCliHandler) AfterApply(ctx context.Context, cwd ports.CWD, env ports.OSEnv) (err error) {
	gitRepo, _, err := infrastructure.NewGitRepositoryFromPath(cwd)
	if err != nil {
		return fmt.Errorf("failed to init git repository: %w", err)
	}

	idStore, err := infrastructure.IdentitiesStore(
		ctx,
		infrastructure.NewAgentIdentitiesStoreSource(env),
		infrastructure.NewFileIdentityStoreSource(h.Keys),
	)
	if err != nil {
		return fmt.Errorf("failed to init identities store: %w", err)
	}

	remotes, err := gitRepo.Remotes()
	if err != nil {
		return fmt.Errorf("failed to determine Git remotes: %w", err)
	}

	query := dto.IdentitiesQuery{
		Remotes: remotes,
	}

	ids, err := idStore.Identities(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to get identities: %w", err)
	}

	h.Opener, err = services.NewAgeSealer(services.WithIdentities(ids...))
	return err
}
