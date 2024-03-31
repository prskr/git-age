package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/alecthomas/kong"

	"github.com/go-git/go-git/v5"

	"github.com/prskr/git-age/core/dto"
	"github.com/prskr/git-age/core/ports"
	"github.com/prskr/git-age/core/services"
	"github.com/prskr/git-age/infrastructure"
)

type ListFilesCliHandler struct{}

func (ListFilesCliHandler) Run(repo ports.GitRepository, stdout ports.STDOUT) error {
	return repo.WalkAgeFiles(func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		if _, err := fmt.Fprintln(stdout, path); err != nil {
			return fmt.Errorf("failed to write to stdout: %w", err)
		}

		return nil
	})
}

type TrackFilesCliHandler struct {
	Pattern string `arg:"" help:"Pattern to track"`

	// relative working directory within the repository
	WorkingDir string `kong:"-"`
}

func (h *TrackFilesCliHandler) Run(repoFS ports.ReadWriteFS) error {
	attributesFile, err := repoFS.Create(filepath.Join(h.WorkingDir, ports.GitAttributesFileName))
	if err != nil {
		return fmt.Errorf("failed to open %s file: %w", ports.GitAttributesFileName, err)
	}

	if _, err := attributesFile.Seek(0, io.SeekEnd); err != nil {
		return err
	}

	defer func() {
		err = errors.Join(err, attributesFile.Close())
	}()

	attributesLine := h.Pattern + " filter=age diff=age merge=age -text\n"
	if _, err := attributesFile.WriteString(attributesLine); err != nil {
		return fmt.Errorf("failed to write to %s file: %w", ports.GitAttributesFileName, err)
	}

	return nil
}

func (h *TrackFilesCliHandler) AfterApply(cwd ports.CWD) (err error) {
	repoRootPath, err := infrastructure.FindRepoRootFrom(cwd)
	if err != nil {
		return err
	}

	h.WorkingDir, err = filepath.Rel(repoRootPath, cwd.Value())

	return err
}

type ReEncryptFilesCliHandler struct {
	Message string `help:"Message to be used for the commit" default:"chore: re-encrypt secret files" short:"m"`
}

func (h ReEncryptFilesCliHandler) Run(
	repo ports.GitRepository,
	repoFS ports.ReadWriteFS,
	sealer ports.FileOpenSealer,
) error {
	if dirty, err := repo.IsStagingDirty(); err != nil {
		return fmt.Errorf("failed to check if repository is dirty: %w", err)
	} else if dirty {
		slog.Warn("Repository is dirty")
		os.Exit(1)
	}

	if err := repo.WalkAgeFiles(services.ReEncryptWalkFunc(repo, repoFS, sealer)); err != nil {
		return err
	}

	slog.Info("Committing changes")
	if err := repo.Commit(h.Message); err != nil {
		return fmt.Errorf("failed to commit changes: %w", err)
	}

	return nil
}

type FilesCliHandler struct {
	KeysFlag  `embed:""`
	List      ListFilesCliHandler      `cmd:"" name:"list" help:"List files" aliases:"ls"`
	Track     TrackFilesCliHandler     `cmd:"" name:"track" help:"Track files"`
	ReEncrypt ReEncryptFilesCliHandler `cmd:"" name:"re-encrypt" help:"Re-encrypt files tracked by git-age"`
}

func (h *FilesCliHandler) AfterApply(ctx context.Context, kongCtx *kong.Context, cwd ports.CWD) error {
	repoRootPath, err := infrastructure.FindRepoRootFrom(cwd)
	if err != nil {
		return err
	}

	repo, err := git.PlainOpen(repoRootPath)
	if err != nil {
		return fmt.Errorf("failed to open git repository: %w", err)
	}

	repoFS := infrastructure.NewReadWriteDirFS(repoRootPath)

	gitRepo, err := infrastructure.NewGitRepository(repoFS, repo)
	if err != nil {
		return err
	}

	idStore, err := infrastructure.IdentitiesStore(
		ctx,
		infrastructure.NewAgentIdentitiesStoreSource(),
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

	sealer, err := services.NewAgeSealer(
		services.WithIdentities(ids...),
		services.WithRecipients(infrastructure.NewRecipientsFile(repoFS)),
	)
	if err != nil {
		return err
	}

	kongCtx.BindTo(repoFS, (*ports.ReadWriteFS)(nil))
	kongCtx.BindTo(gitRepo, (*ports.GitRepository)(nil))
	kongCtx.BindTo(sealer, (*ports.FileOpenSealer)(nil))

	return nil
}
