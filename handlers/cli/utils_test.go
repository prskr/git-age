package cli_test

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/adrg/xdg"
	"github.com/alecthomas/kong"
	"github.com/go-git/go-billy/v5/osfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/cache"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/storage/filesystem"

	"github.com/prskr/git-age/core/ports"
	"github.com/prskr/git-age/core/services"
	"github.com/prskr/git-age/infrastructure"
	"github.com/prskr/git-age/internal/fsx"
	"github.com/prskr/git-age/internal/testx"
)

func newKong(tb testing.TB, grammar any) *kong.Kong {
	tb.Helper()
	inst, err := kong.New(grammar, kong.Vars{
		"XDG_CONFIG_HOME":     xdg.ConfigHome,
		"file_path_separator": string(filepath.Separator),
	})
	if err != nil {
		tb.Fatalf("failed to create parser: %v", err)
	}

	return inst
}

type testSetup struct {
	root   string
	repo   *git.Repository
	repoFS ports.ReadWriteFS
}

func prepareTestRepo(tb testing.TB) (s *testSetup) {
	tb.Helper()

	s = new(testSetup)

	s.root = tb.TempDir()
	wd := testx.ResultOf(tb, os.Getwd)

	tb.Cleanup(func() {
		if err := os.Chdir(wd); err != nil {
			tb.Errorf("failed to restore working directory: %v", err)
		}
	})

	srcFS := infrastructure.NewReadWriteDirFS(filepath.Join(wd, "testdata"))
	s.repoFS = infrastructure.NewReadWriteDirFS(s.root)

	recipients := infrastructure.NewRecipientsFile(srcFS)

	sealer, err := services.NewAgeSealer(services.WithRecipients(recipients))
	if err != nil {
		tb.Fatalf("failed to create sealer: %v", err)
	}

	err = fsx.NewSyncer(srcFS, s.repoFS, fsx.SealingMiddleware(sealer, "*.env")).Sync()
	if err != nil {
		tb.Fatalf("failed to populate test directory: %v", err)
	}

	gitDirPath := filepath.Join(s.root, ".git")
	if err := os.MkdirAll(gitDirPath, 0o755); err != nil {
		tb.Fatalf("failed to create .git directory: %v", err)
	}

	gitfs := osfs.New(gitDirPath, osfs.WithBoundOS())
	objCache := cache.NewObjectLRU(4 * 1024)
	fsStorer := filesystem.NewStorage(gitfs, objCache)

	s.repo = testx.ResultOfA[*git.Repository](tb, git.Init, fsStorer, osfs.New(s.root, osfs.WithBoundOS()))
	wt := testx.ResultOf(tb, s.repo.Worktree)

	gitCfg, err := s.repo.Config()
	if err != nil {
		tb.Fatalf("failed to get git config: %v", err)
	}

	gitCfg.User.Name = tb.Name()
	gitCfg.User.Email = "ci@git-age.io"

	if err := s.repo.SetConfig(gitCfg); err != nil {
		tb.Fatalf("failed to set git config: %v", err)
	}

	err = fs.WalkDir(s.repoFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			if path == ".git" {
				return fs.SkipDir
			}
			return nil
		}

		if _, err := wt.Add(path); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		tb.Fatalf("failed to stage files: %v", err)
	}

	_ = &git.CommitOptions{
		Author: &object.Signature{
			Name:  tb.Name(),
			Email: "ci@git-age.io",
			When:  time.Now().UTC(),
		},
	}

	testx.ResultOfA[plumbing.Hash](tb, wt.Commit, "initial commit", new(git.CommitOptions))

	err = fsx.NewSyncer(srcFS, s.repoFS).Sync()
	if err != nil {
		tb.Fatalf("failed to populate test directory: %v", err)
	}

	if err := os.Chdir(s.root); err != nil {
		tb.Errorf("failed to change directory: %v", err)
	}

	return s
}
