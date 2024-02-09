package cli

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v5"

	"github.com/prskr/git-age/core/ports"
	"github.com/prskr/git-age/core/services"
	"github.com/prskr/git-age/infrastructure"
	"github.com/urfave/cli/v2"
)

type FilesCliHandler struct {
	// relative working directory within the repository
	WorkingDir string
	RepoFS     ports.ReadWriteFS
	Encryption ports.FileOpenSealer
	Repository ports.GitRepository
}

func (h *FilesCliHandler) ListFiles(*cli.Context) error {
	return h.Repository.WalkAgeFiles(func(path string, d fs.DirEntry, err error) error {
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

func (h *FilesCliHandler) TrackFiles(ctx *cli.Context) (err error) {
	if ctx.NArg() != 1 {
		return cli.Exit("Expected exactly one argument", 1)
	}

	attributesFile, err := h.RepoFS.OpenRW(filepath.Join(h.WorkingDir, ".gitattributes"))
	if err != nil {
		return fmt.Errorf("failed to open .gitattributes file: %w", err)
	}

	if _, err := attributesFile.Seek(0, io.SeekEnd); err != nil {
		return err
	}

	defer func() {
		err = errors.Join(err, attributesFile.Close())
	}()

	attributesLine := ctx.Args().First() + " filter=age diff=age merge=age -text\n"
	if _, err := attributesFile.WriteString(attributesLine); err != nil {
		return fmt.Errorf("failed to write to .gitattributes file: %w", err)
	}

	return nil
}

func (h *FilesCliHandler) ReEncryptFiles(*cli.Context) error {
	if clean, err := h.Repository.IsStagingDirty(); err != nil {
		return fmt.Errorf("failed to check if repository is dirty: %w", err)
	} else if !clean {
		return cli.Exit("Repository is dirty", 1)
	}

	if err := h.Repository.WalkAgeFiles(services.ReEncryptWalkFunc(h.Repository, h.RepoFS, h.Encryption)); err != nil {
		return err
	}

	return nil
}

func (h *FilesCliHandler) Command() *cli.Command {
	return &cli.Command{
		Name: "files",
		Flags: []cli.Flag{
			&keysFlag,
		},
		Subcommands: []*cli.Command{
			{
				Name:    "list",
				Aliases: []string{"ls"},
				Action:  h.ListFiles,
			},
			{
				Name:   "track",
				Action: h.TrackFiles,
				Args:   true,
			},
			{
				Name:   "re-encrypt",
				Action: h.ReEncryptFiles,
			},
		},
		Before: func(context *cli.Context) error {
			wd, err := os.Getwd()
			if err != nil {
				return err
			}

			repoRootPath, err := infrastructure.FindRepoRootFrom(wd)
			if err != nil {
				return err
			}

			h.WorkingDir, err = filepath.Rel(repoRootPath, wd)
			if err != nil {
				return err
			}

			repo, err := git.PlainOpen(repoRootPath)
			if err != nil {
				return fmt.Errorf("failed to open git repository: %w", err)
			}

			h.RepoFS = infrastructure.NewReadWriteDirFS(repoRootPath)

			h.Repository, err = infrastructure.NewGitRepository(h.RepoFS, repo)
			if err != nil {
				return err
			}

			h.Encryption, err = services.NewAgeSealer(
				services.WithIdentities(infrastructure.NewIdentities(context.String("keys"))),
				services.WithRecipients(infrastructure.NewRecipientsFile(h.RepoFS)),
			)

			return err
		},
	}
}
