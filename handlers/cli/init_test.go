package cli_test

import (
	"os"
	"path/filepath"
	"testing"

	"filippo.io/age"
	"github.com/alecthomas/kong"

	"github.com/prskr/git-age/core/ports"
	"github.com/prskr/git-age/handlers/cli"
)

func TestInitCliHandler_Run(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	setup := prepareTestRepo(t)

	if err := os.Remove(filepath.Join(setup.root, ports.RecipientsFileName)); err != nil {
		t.Errorf("failed to remove file: %v", err)
		return
	}

	parser := newKong(
		t,
		new(cli.InitCliHandler),
		kong.Bind(ports.CWD(setup.root)),
	)

	keysFilePath := filepath.Join(tmpDir, "keys.txt")
	args := []string{
		"-k", keysFilePath,
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

	keysFile, err := os.Open(keysFilePath)
	if err != nil {
		t.Errorf("failed to open file: %v", err)
		return
	}

	t.Cleanup(func() {
		_ = keysFile.Close()
	})

	if _, err := age.ParseIdentities(keysFile); err != nil {
		t.Errorf("failed to parse identities: %v", err)
		return
	}

	recipientsFile, err := os.Open(filepath.Join(setup.root, ports.RecipientsFileName))
	if err != nil {
		t.Errorf("failed to open file: %v", err)
		return
	}

	t.Cleanup(func() {
		_ = recipientsFile.Close()
	})

	if _, err := age.ParseRecipients(recipientsFile); err != nil {
		t.Errorf("failed to parse recipients: %v", err)
		return
	}
}
