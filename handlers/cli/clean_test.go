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

func TestCleanCliHandler_Run(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name           string
		inFileModifier func(*os.File) error
	}{
		{
			name: "Unmodified file",
		},
		{
			name: "Modified file",
			inFileModifier: func(f *os.File) error {
				if _, err := f.Seek(0, io.SeekEnd); err != nil {
					return err
				}

				if _, err := f.Write([]byte("test")); err != nil {
					return err
				}

				if _, err := f.Seek(0, io.SeekStart); err != nil {
					return err
				}

				return nil
			},
		},
	}

	//nolint:paralleltest // not necessary anymore in Go 1.22
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			setup := prepareTestRepo(t)

			out := new(bytes.Buffer)

			inFile, err := os.OpenFile(filepath.Join(setup.root, ".env"), os.O_RDWR, 0o644)
			if err != nil {
				t.Errorf("failed to open file: %v", err)
				return
			}

			t.Cleanup(func() {
				_ = inFile.Close()
			})

			if tt.inFileModifier != nil {
				if err := tt.inFileModifier(inFile); err != nil {
					t.Errorf("failed to modify input file: %v", err)
					return
				}
			}

			parser := newKong(
				t,
				new(cli.CleanCliHandler),
				kong.Bind(ports.CWD(setup.root)),
				kong.BindTo(ports.STDIN(inFile), (*ports.STDIN)(nil)),
				kong.BindTo(ports.STDOUT(out), (*ports.STDOUT)(nil)),
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

			ids, err := age.ParseIdentities(bytes.NewReader(keys))
			if err != nil {
				t.Errorf("failed to parse identities: %v", err)
				return
			}

			reader, err := age.Decrypt(out, ids...)
			if err != nil {
				t.Errorf("failed to decrypt file: %v", err)
				return
			}

			if tt.inFileModifier != nil {
				return
			}

			hash := sha256.New()
			if _, err := io.Copy(hash, reader); err != nil {
				t.Errorf("failed to copy file to hash: %v", err)
				return
			}

			outHash := hash.Sum(nil)
			if !bytes.Equal(expectedHash, outHash) {
				t.Errorf("input and output hashes do not match")
			}
		})
	}
}
