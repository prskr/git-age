package cli_test

import (
	_ "embed"
	"encoding/hex"
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

var (
	expectedHash []byte

	//go:embed testdata/sampleRepo/keys.txt
	keys []byte

	//go:embed testdata/sampleRepo/.agerecipients
	recipients []byte
)

func init() {
	s, err := hex.DecodeString("8a7d8f4374d752b3a46ef521c4b39325d1172211c487cc4ada4cd587dcce2cd5")
	if err != nil {
		panic(err)
	}

	expectedHash = s
}

func newKong(tb testing.TB, grammar any, opts ...kong.Option) *kong.Kong {
	tb.Helper()

	opts = append(opts, kong.Vars{
		"XDG_CONFIG_HOME":     xdg.ConfigHome,
		"file_path_separator": string(filepath.Separator),
	})

	inst, err := kong.New(grammar, opts...)
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

	srcFS := infrastructure.NewReadWriteDirFS(filepath.Join(wd, "testdata", "sampleRepo"))
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
			When: time.Now().UTC(),
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
