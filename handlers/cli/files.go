package cli

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/alecthomas/kong"

	"github.com/go-git/go-git/v5"

	"github.com/prskr/git-age/core/ports"
	"github.com/prskr/git-age/core/services"
	"github.com/prskr/git-age/infrastructure"
)

type ListFilesCliHandler struct{}

func (ListFilesCliHandler) Run(repo ports.GitRepository) error {
	return repo.WalkAgeFiles(func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		fmt.Println(path)

		return nil
	})
}

type TrackFilesCliHandler struct {
	Pattern string `arg:"" help:"Pattern to track"`

	// relative working directory within the repository
	WorkingDir string `kong:"-"`
}

func (h *TrackFilesCliHandler) Run(repoFS ports.ReadWriteFS) error {
	attributesFile, err := repoFS.OpenRW(filepath.Join(h.WorkingDir, ".gitattributes"))
	if err != nil {
		return fmt.Errorf("failed to open .gitattributes file: %w", err)
	}

	if _, err := attributesFile.Seek(0, io.SeekEnd); err != nil {
		return err
	}

	defer func() {
		err = errors.Join(err, attributesFile.Close())
	}()

	attributesLine := h.Pattern + " filter=age diff=age merge=age -text\n"
	if _, err := attributesFile.WriteString(attributesLine); err != nil {
		return fmt.Errorf("failed to write to .gitattributes file: %w", err)
	}

	return nil
}

func (h *TrackFilesCliHandler) AfterApply() (err error) {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	repoRootPath, err := infrastructure.FindRepoRootFrom(wd)
	if err != nil {
		return err
	}

	h.WorkingDir, err = filepath.Rel(repoRootPath, wd)

	return err
}

type ReEncryptFilesCliHandler struct{}

func (ReEncryptFilesCliHandler) Run(
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

	return nil
}

type FilesCliHandler struct {
	KeysFlag  `embed:""`
	List      ListFilesCliHandler      `cmd:"" help:"List files"`
	Track     TrackFilesCliHandler     `cmd:"" help:"Track files"`
	ReEncrypt ReEncryptFilesCliHandler `cmd:"" help:"Re-encrypt files tracked by git-age"`
}

func (h *FilesCliHandler) AfterApply(kctx *kong.Context) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	repoRootPath, err := infrastructure.FindRepoRootFrom(wd)
	if err != nil {
		return err
	}

	repo, err := git.PlainOpen(repoRootPath)
	if err != nil {
		return fmt.Errorf("failed to open git repository: %w", err)
	}

	repoFS := infrastructure.NewReadWriteDirFS(repoRootPath)

	gr, err := infrastructure.NewGitRepository(repoFS, repo)
	if err != nil {
		return err
	}

	sealer, err := services.NewAgeSealer(
		services.WithIdentities(infrastructure.NewIdentities(h.Keys)),
		services.WithRecipients(infrastructure.NewRecipientsFile(repoFS)),
	)
	if err != nil {
		return err
	}

	kctx.BindTo(repoFS, (*ports.ReadWriteFS)(nil))
	kctx.BindTo(gr, (*ports.GitRepository)(nil))
	kctx.BindTo(sealer, (*ports.FileOpenSealer)(nil))

	return nil
}
