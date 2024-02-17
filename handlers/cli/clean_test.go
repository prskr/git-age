package cli_test

import (
	"bytes"
	_ "embed"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"testing"

	"filippo.io/age"
	"github.com/minio/sha256-simd"
	"github.com/prskr/git-age/handlers/cli"
)

var (
	expectedhash []byte

	//go:embed testdata/keys.txt
	keys []byte
)

func init() {
	s, err := hex.DecodeString("8a7d8f4374d752b3a46ef521c4b39325d1172211c487cc4ada4cd587dcce2cd5")
	if err != nil {
		panic(err)
	}

	expectedhash = s
}

func TestCleanCliHandler_Run(t *testing.T) {
	setup := prepareTestRepo(t)

	if err := os.Chdir(setup.root); err != nil {
		t.Errorf("failed to change directory: %v", err)
		return
	}

	outFile, err := os.CreateTemp(t.TempDir(), ".env")
	if err != nil {
		t.Errorf("failed to create temp file: %v", err)
		return
	}

	os.Stdout = outFile

	parser := newKong(t, new(cli.CleanCliHandler))

	args := []string{
		"-k", filepath.Join(setup.root, "keys.txt"),
		".env",
	}

	ctx, err := parser.Parse(args)
	if err != nil {
		t.Errorf("failed to parse arguments: %v", err)
		return
	}

	f, err := os.Open(filepath.Join(setup.root, ".env"))
	if err != nil {
		t.Errorf("failed to open file: %v", err)
		return
	}

	os.Stdin = f
	t.Cleanup(func() {
		_ = f.Close()
	})

	if err := ctx.Run(); err != nil {
		t.Errorf("failed to run command: %v", err)
		return
	}

	if _, err := outFile.Seek(0, io.SeekStart); err != nil {
		t.Errorf("failed to seek to start: %v", err)
		return
	}

	ids, err := age.ParseIdentities(bytes.NewReader(keys))
	if err != nil {
		t.Errorf("failed to parse identities: %v", err)
		return
	}

	reader, err := age.Decrypt(outFile, ids...)
	if err != nil {
		t.Errorf("failed to decrypt file: %v", err)
		return
	}

	hash := sha256.New()
	if _, err := io.Copy(hash, reader); err != nil {
		t.Errorf("failed to copy file to hash: %v", err)
		return
	}

	outHash := hash.Sum(nil)
	if !bytes.Equal(expectedhash, outHash) {
		t.Errorf("input and output hashes do not match")
	}
}
