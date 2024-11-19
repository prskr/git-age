package cli_test

import (
	"bytes"
	"context"
	"os"
	"testing"

	"filippo.io/age"
	"github.com/alecthomas/kong"
	"github.com/stretchr/testify/assert"

	"github.com/prskr/git-age/core/ports"
	"github.com/prskr/git-age/handlers/cli"
	"github.com/prskr/git-age/internal/testx"
)

func TestListKeysCliHandler_Run(t *testing.T) {
	t.Parallel()

	outBuf := new(bytes.Buffer)
	keysFile, err := os.CreateTemp(t.TempDir(), "keys.txt")
	if err != nil {
		t.Errorf("failed to create temporary keys.txt file: %v", err)
		return
	}

	id := testx.ResultOf(t, age.GenerateX25519Identity)
	if _, err = keysFile.Write([]byte(id.String() + "\n")); err != nil {
		t.Errorf("failed to write identity to keys.txt file: %v", err)
		return
	}

	if !assert.NoError(t, keysFile.Close(), "failed to close keys.txt file") {
		return
	}

	parser := newKong(
		t,
		new(cli.ListKeysCliHandler),
		kong.BindTo(testx.Context(t), (*context.Context)(nil)),
		kong.BindTo(ports.STDOUT(outBuf), (*ports.STDOUT)(nil)),
		kong.Bind(ports.NewOSEnv()),
	)

	args := []string{
		"-k", "file:///" + keysFile.Name(),
	}

	kongCtx, err := parser.Parse(args)
	if !assert.NoError(t, err, "failed to parse arguments") {
		return
	}

	if !assert.NoError(t, kongCtx.Run(), "failed to run command") {
		return
	}

	assert.Contains(t, outBuf.String(), "Public Key")
	assert.Contains(t, outBuf.String(), id.Recipient().String())
}
