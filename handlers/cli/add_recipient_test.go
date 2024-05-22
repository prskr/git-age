package cli_test

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/alecthomas/kong"

	"github.com/prskr/git-age/core/ports"
	"github.com/prskr/git-age/internal/testx"

	"filippo.io/age"

	"github.com/prskr/git-age/handlers/cli"
	"github.com/prskr/git-age/infrastructure"
)

func TestAddRecipientCliHandler_Run(t *testing.T) {
	t.Parallel()

	setup := prepareTestRepo(t)

	idToAdd, err := age.GenerateX25519Identity()
	if err != nil {
		t.Errorf("failed to generate identity: %v", err)
		return
	}

	parser := newKong(
		t,
		new(cli.AddRecipientCliHandler),
		kong.Bind(ports.CWD(setup.root)),
		kong.BindTo(testx.Context(t), (*context.Context)(nil)),
	)

	args := []string{
		"-k", fmt.Sprintf("file:///%s/keys.txt", filepath.ToSlash(setup.root)),
		"-c", "test comment",
		idToAdd.Recipient().String(),
	}

	ctx, err := parser.Parse(args)
	if err != nil {
		t.Errorf("failed to parse arguments: %v", err)
		return
	}

	if err := ctx.Run(); err != nil {
		t.Errorf("failed to run command: %v", err)
		return
	}

	repo, err := infrastructure.NewGitRepository(setup.repoFS, setup.repo)
	if err != nil {
		t.Errorf("failed to create repository: %v", err)
		return
	}

	obj, err := repo.OpenObjectAtHead(".env")
	if err != nil {
		t.Errorf("failed to open object: %v", err)
		return
	}

	objReader, err := obj.Reader()
	if err != nil {
		t.Errorf("failed to get reader: %v", err)
		return
	}

	t.Cleanup(func() {
		_ = objReader.Close()
	})

	_, err = age.Decrypt(objReader, idToAdd)
	if err != nil {
		t.Errorf("failed to decrypt file: %v", err)
	}
}
