package cli

import (
	"fmt"
	"github.com/adrg/xdg"
	"github.com/go-git/go-git/v5"
	"github.com/prskr/git-age/core/ports"
	"github.com/prskr/git-age/core/services"
	"github.com/prskr/git-age/infrastructure"
	"github.com/urfave/cli/v2"
	"io/fs"
	"os"
	"path/filepath"
)

type FilesCliHandler struct {
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

func (h *FilesCliHandler) TrackFiles(*cli.Context) error {
	return nil
}

func (h *FilesCliHandler) ReEncryptFiles(*cli.Context) error {
	if clean, err := h.Repository.IsDirty(); err != nil {
		return fmt.Errorf("failed to check if repository is dirty: %w", err)
	} else if !clean {
		return cli.Exit("Repository is dirty", 1)
	}

	if err := h.Repository.WalkAgeFiles(ports.ReEncryptWalkFunc(h.Repository, h.RepoFS, h.Encryption)); err != nil {
		return err
	}

	return nil
}

func (h *FilesCliHandler) Command() *cli.Command {
	return &cli.Command{
		Name: "files",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "keys",
				DefaultText: "By default keys are read from $XDG_CONFIG_HOME/git-age/keys.txt i.e. $HOME/.config/git-age/keys.txt on most systems",
				EnvVars: []string{
					"GIT_AGE_KEYS",
				},
			},
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

			repoRootPath, err := repoRoot(wd)
			if err != nil {
				return err
			}

			gitRepo, err := git.PlainOpen(repoRootPath)
			if err != nil {
				return err
			}

			h.RepoFS = infrastructure.NewReadWriteDirFS(repoRootPath)

			h.Repository, err = infrastructure.NewGitRepository(h.RepoFS, gitRepo)
			if err != nil {
				return err
			}

			keysPath := filepath.Join(xdg.ConfigHome, "git-age", "keys.txt")
			if flagPath := context.String("keys"); flagPath != "" {
				keysPath = flagPath
			}

			h.Encryption, err = services.NewAgeSealer(
				services.WithIdentitiesFrom(keysPath),
				services.WithRecipientsFrom(wd),
			)

			return err
		},
	}
}
