package services_test

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"time"

	"filippo.io/age"
	"github.com/go-git/go-billy/v5/osfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/cache"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/storage/filesystem"
	"github.com/minio/sha256-simd"
	"github.com/prskr/git-age/core/ports"
	"github.com/prskr/git-age/core/services"
	"github.com/prskr/git-age/infrastructure"
	"github.com/prskr/git-age/internal/fsx"
	"github.com/prskr/git-age/internal/testx"
)

func TestReEncryptWalkFunc(t *testing.T) {
	t.Parallel()
	setup := prepareRepo(t)

	additionalId, err := age.GenerateX25519Identity()
	if err != nil {
		t.Fatalf("failed to generate identity: %v", err)
	}

	g, err := infrastructure.NewGitRepository(setup.repoFS, setup.repo)
	if err != nil {
		t.Errorf("failed to create git repository: %v", err)
		return
	}

	sealer, _ := services.NewAgeSealer()
	sealer.AddRecipients(setup.id.Recipient(), additionalId.Recipient())
	sealer.AddIdentities(setup.id, additionalId)

	err = g.WalkAgeFiles(services.ReEncryptWalkFunc(g, setup.repoFS, sealer))
	if err != nil {
		t.Errorf("failed to re-encrypt files: %v", err)
		return
	}

	if err := g.Commit("re-encrypt files"); err != nil {
		t.Errorf("failed to commit changes: %v", err)
		return
	}

	sealer, _ = services.NewAgeSealer()
	sealer.AddIdentities(additionalId)

	err = g.WalkAgeFiles(func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() || filepath.Ext(path) == ".gitattributes" {
			return nil
		}

		obj, err := g.OpenObjectAtHead(path)
		if err != nil {
			return err
		}

		objReader, err := obj.Reader()
		if err != nil {
			return err
		}

		defer func() {
			_ = objReader.Close()
		}()

		h := sha256.New()

		plainReader, err := sealer.OpenFile(objReader)
		if err != nil {
			return err
		}

		_, err = io.Copy(h, plainReader)
		if err != nil {
			return err
		}

		origHash, err := fsx.HashFSFile(setup.repoFS, path)
		if err != nil {
			return err
		}

		if !bytes.Equal(h.Sum(nil), origHash) {
			return fmt.Errorf("hash mismatch for %s", path)
		}

		return nil
	})
	if err != nil {
		t.Errorf("failed to verify re-encrypted files: %v", err)
	}
}

type testSetup struct {
	root   string
	id     *age.X25519Identity
	repo   *git.Repository
	repoFS ports.ReadWriteFS
}

func prepareRepo(tb testing.TB) (s *testSetup) {
	tb.Helper()

	s = new(testSetup)

	var err error
	s.id, err = age.GenerateX25519Identity()
	if err != nil {
		tb.Fatalf("failed to generate identity: %v", err)
	}

	sealer, _ := services.NewAgeSealer()
	sealer.AddRecipients(s.id.Recipient())

	s.root = tb.TempDir()
	wd := testx.ResultOf(tb, os.Getwd)

	srcFS := os.DirFS(filepath.Join(wd, "testdata"))
	s.repoFS = infrastructure.NewReadWriteDirFS(s.root)

	sealingOpts := []fsx.MiddlewareProvider{
		fsx.SealingMiddleware(sealer, "*.yaml"),
		fsx.SealingMiddleware(sealer, "*/*.json"),
	}

	if err := fsx.NewSyncer(srcFS, s.repoFS, sealingOpts...).Sync(); err != nil {
		tb.Fatalf("failed to populate test directory: %v", err)
	}

	if err := fsx.WriteTo(s.repoFS, ports.RecipientsFileName, []byte(s.id.Recipient().String())); err != nil {
		tb.Fatalf("failed to write recipients file: %v", err)
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

	for _, p := range []string{"values.yaml", filepath.Join("config", "secrets.json")} {
		if _, err := wt.Add(p); err != nil {
			tb.Fatalf("failed to add file: %v", err)
		}
	}

	_ = &git.CommitOptions{
		Author: &object.Signature{
			Name:  tb.Name(),
			Email: "ci@git-age.io",
			When:  time.Now().UTC(),
		},
	}

	testx.ResultOfA[plumbing.Hash](tb, wt.Commit, "initial commit", new(git.CommitOptions))

	return s
}
