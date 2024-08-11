package cli_test

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"filippo.io/age"
	"github.com/alecthomas/kong"
	"github.com/go-git/go-git/v5"
	"github.com/minio/sha256-simd"

	"github.com/prskr/git-age/core/ports"
	"github.com/prskr/git-age/handlers/cli"
	"github.com/prskr/git-age/infrastructure"
	"github.com/prskr/git-age/internal/testx"
)

func TestListFilesCliHandler_Run(t *testing.T) {
	t.Parallel()
	setup := prepareTestRepo(t)

	out := new(bytes.Buffer)
	parser := newKong(
		t,
		new(cli.FilesCliHandler),
		kong.Bind(ports.CWD(setup.root)),
		kong.BindTo(testx.Context(t), (*context.Context)(nil)),
		kong.BindTo(ports.STDOUT(out), (*ports.STDOUT)(nil)),
		kong.Bind(ports.NewOSEnv()),
	)

	ctx, err := parser.Parse([]string{"ls"})
	if err != nil {
		t.Errorf("failed to parse arguments: %v", err)
		return
	}

	if err := ctx.Run(); err != nil {
		t.Errorf("failed to run command: %v", err)
		return
	}

	expected := map[string]struct{}{
		".env": {},
	}

	scanner := bufio.NewScanner(out)
	for scanner.Scan() {
		delete(expected, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		t.Errorf("failed to scan output: %v", err)
		return
	}

	if len(expected) > 0 {
		t.Errorf("missing files: %v", expected)
		return
	}
}

func TestTrackFilesCliHandler_Run(t *testing.T) {
	t.Parallel()
	setup := prepareTestRepo(t)

	parser := newKong(
		t,
		new(cli.FilesCliHandler),
		kong.Bind(ports.CWD(setup.root)),
		kong.BindTo(testx.Context(t), (*context.Context)(nil)),
		kong.Bind(ports.NewOSEnv()),
	)

	args := []string{
		"track",
		"appsettings.*.json",
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

	attributesFile, err := os.Open(filepath.Join(setup.root, ports.GitAttributesFileName))
	if err != nil {
		t.Errorf("failed to open .gitattributes file: %v", err)
		return
	}

	t.Cleanup(func() {
		_ = attributesFile.Close()
	})

	expected := "appsettings.*.json filter=age diff=age merge=age -text"
	scanner := bufio.NewScanner(attributesFile)
	for scanner.Scan() {
		if scanner.Text() == expected {
			return
		}
	}

	t.Error("expected line not found in .gitattributes file")
}

func TestReEncryptFilesCliHandler_Run(t *testing.T) {
	t.Parallel()
	setup := prepareTestRepo(t)

	// add a recipient to the repository before re-encrypting
	newId, err := age.GenerateX25519Identity()
	if err != nil {
		t.Errorf("failed to generate identity: %v", err)
		return
	}

	recp := infrastructure.NewRecipientsFile(setup.repoFS)
	if _, err := recp.Append(newId.Recipient().String(), ""); err != nil {
		t.Errorf("failed to append recipient: %v", err)
		return
	}

	wt, err := setup.repo.Worktree()
	if err != nil {
		t.Errorf("failed to get worktree: %v", err)
		return
	}

	if _, err := wt.Add(ports.RecipientsFileName); err != nil {
		t.Errorf("failed to add file: %v", err)
		return
	}

	if _, err := wt.Commit("chore: add recipient", new(git.CommitOptions)); err != nil {
		t.Errorf("failed to commit: %v", err)
		return
	}

	// re-encrypt the .env file
	parser := newKong(
		t,
		new(cli.FilesCliHandler),
		kong.Bind(ports.CWD(setup.root)),
		kong.BindTo(testx.Context(t), (*context.Context)(nil)),
		kong.Bind(ports.NewOSEnv()),
	)

	args := []string{
		"-k", fmt.Sprintf("file:///%s/keys.txt", filepath.ToSlash(setup.root)),
		"re-encrypt",
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

	// check if the .env file was re-encrypted and can be opened with new identity
	repo, err := infrastructure.NewGitRepository(setup.repoFS, setup.repo)
	if err != nil {
		t.Errorf("failed to create git repository: %v", err)
		return
	}

	obj, err := repo.OpenObjectAtHead(".env")
	if err != nil {
		t.Errorf("failed to open object: %v", err)
		return
	}

	objReader, err := obj.Reader()
	if err != nil {
		t.Errorf("failed to open object reader: %v", err)
		return
	}

	t.Cleanup(func() {
		_ = objReader.Close()
	})

	r, err := age.Decrypt(objReader, newId)
	if err != nil {
		t.Errorf("failed to decrypt .env file: %v", err)
		return
	}

	hash := sha256.New()
	if _, err := io.Copy(hash, r); err != nil {
		t.Errorf("failed to hash decrypted .env file: %v", err)
		return
	}

	if !bytes.Equal(hash.Sum(nil), expectedHash) {
		t.Errorf("decrypted .env file hash does not match expected")
	}
}
