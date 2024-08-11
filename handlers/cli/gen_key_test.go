package cli_test

import (
	"bytes"
	"context"
	"os"
	"testing"

	"filippo.io/age"
	"github.com/alecthomas/kong"

	"github.com/prskr/git-age/core/ports"
	"github.com/prskr/git-age/handlers/cli"
	"github.com/prskr/git-age/internal/testx"
)

func TestGenKeyCliHandler_Run(t *testing.T) {
	t.Parallel()

	f, err := os.CreateTemp(t.TempDir(), "keys")
	if err != nil {
		t.Errorf("failed to create temp file: %v", err)
		return
	}

	if err := f.Close(); err != nil {
		t.Errorf("failed to close temp file: %v", err)
		return
	}

	outBuf := new(bytes.Buffer)

	parser := newKong(
		t,
		new(cli.GenKeyCliHandler),
		kong.BindTo(testx.Context(t), (*context.Context)(nil)),
		kong.BindTo(ports.STDOUT(outBuf), (*ports.STDOUT)(nil)),
		kong.Bind(ports.NewOSEnv()),
	)

	args := []string{
		"-k", "file:///" + f.Name(),
		"-c", "test",
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

	if _, err := age.ParseRecipients(outBuf); err != nil {
		t.Errorf("failed to parse identities: %v", err)
		return
	}

	f, err = os.OpenFile(f.Name(), os.O_RDONLY, 0o644)
	if err != nil {
		t.Errorf("failed to open file: %v", err)
		return
	}

	t.Cleanup(func() {
		_ = f.Close()
	})

	if _, err := age.ParseIdentities(f); err != nil {
		t.Errorf("failed to parse recipients: %v", err)
	}
}
