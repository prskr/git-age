package cli_test

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	"filippo.io/age"
	"github.com/alecthomas/kong"
	"github.com/minio/sha256-simd"

	"github.com/prskr/git-age/core/ports"
	"github.com/prskr/git-age/handlers/cli"
)

func TestSmudgeCliHandler_Run(t *testing.T) {
	t.Parallel()
	setup := prepareTestRepo(t)
	inBuf := encryptFileToBuffer(t, filepath.Join(setup.root, ".env"))
	outBuf := new(bytes.Buffer)

	parser := newKong(
		t,
		new(cli.SmudgeCliHandler),
		kong.Bind(ports.CWD(setup.root)),
		kong.BindTo(ports.STDIN(io.NopCloser(inBuf)), (*ports.STDIN)(nil)),
		kong.BindTo(ports.STDOUT(outBuf), (*ports.STDOUT)(nil)),
	)

	args := []string{
		"-k", filepath.Join(setup.root, "keys.txt"),
		".env",
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

	hash := sha256.New()
	if _, err := io.Copy(hash, outBuf); err != nil {
		t.Errorf("failed to copy file to hash: %v", err)
		return
	}

	outHash := hash.Sum(nil)
	if !bytes.Equal(expectedHash, outHash) {
		t.Errorf("input and output hashes do not match")
	}
}

func encryptFileToBuffer(tb testing.TB, filePath string) *bytes.Buffer {
	tb.Helper()

	file, err := os.Open(filePath)
	if err != nil {
		tb.Fatalf("failed to open file: %v", err)
	}

	defer func() {
		if err := file.Close(); err != nil {
			tb.Fatalf("failed to close file: %v", err)
		}
	}()

	r, err := age.ParseRecipients(bytes.NewReader(recipients))
	if err != nil {
		tb.Fatalf("failed to parse recipients: %v", err)
	}

	buf := new(bytes.Buffer)

	ageWriter, err := age.Encrypt(buf, r...)
	if err != nil {
		tb.Fatalf("failed to create encryptor: %v", err)
	}

	defer func() {
		if err := ageWriter.Close(); err != nil {
			tb.Fatalf("failed to close encryptor: %v", err)
		}
	}()

	_, err = io.Copy(ageWriter, file)
	if err != nil {
		tb.Fatalf("failed to copy file to buffer: %v", err)
	}

	return buf
}
