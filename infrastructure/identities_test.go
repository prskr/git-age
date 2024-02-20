package infrastructure_test

import (
	"io"
	"os"
	"testing"

	"filippo.io/age"

	"github.com/prskr/git-age/infrastructure"
)

func TestIdentities_All(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		setup      func(tb testing.TB, writer io.StringWriter)
		wantNumber int
		wantErr    bool
	}{
		{
			name:    "empty keys file",
			wantErr: true,
		},
		{
			name:    "garbage in file",
			wantErr: true,
			setup: func(tb testing.TB, writer io.StringWriter) {
				tb.Helper()
				if _, err := writer.WriteString("garbage"); err != nil {
					tb.Fatalf("failed to write garbage to file: %v", err)
				}
			},
		},
		{
			name: "single key",
			setup: func(tb testing.TB, writer io.StringWriter) {
				tb.Helper()
				id, err := age.GenerateX25519Identity()
				if err != nil {
					tb.Fatalf("failed to create age identity: %v", err)
				}

				if _, err := writer.WriteString(id.String()); err != nil {
					tb.Fatalf("failed to write identity to file: %v", err)
				}
			},
			wantNumber: 1,
		},
		{
			name: "multiple keys",
			setup: func(tb testing.TB, writer io.StringWriter) {
				tb.Helper()
				for range 5 {
					id, err := age.GenerateX25519Identity()
					if err != nil {
						tb.Fatalf("failed to create age identity: %v", err)
					}

					if _, err := writer.WriteString(id.String() + "\n"); err != nil {
						tb.Fatalf("failed to write identity to file: %v", err)
					}
				}
			},
			wantNumber: 5,
		},
	}

	//nolint:paralleltest // not necessary anymore in Go 1.22
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			keysFile, err := os.CreateTemp(t.TempDir(), "keys*.txt")
			if err != nil {
				t.Errorf("failed to create temp file: %v", err)
				return
			}

			if tt.setup != nil {
				tt.setup(t, keysFile)
			}

			if err := keysFile.Close(); err != nil {
				t.Errorf("failed to close keys file after setup: %v", err)
				return
			}

			i := infrastructure.NewIdentities(keysFile.Name())

			got, err := i.All()
			if (err != nil) != tt.wantErr {
				t.Errorf("All() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(got) != tt.wantNumber {
				t.Errorf("expected %d identites, got %d", tt.wantNumber, len(got))
			}
		})
	}
}

func TestIdentities_Generate(t *testing.T) {
	t.Parallel()

	type args struct {
		comment string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "empty comment",
		},
		{
			name: "simple comment",
			args: args{
				comment: "some comment",
			},
		},
		{
			name: "multi-line comment",
			args: args{
				comment: `some comment
with multiple lines`,
			},
		},
	}

	//nolint:paralleltest // not necessary anymore in Go 1.22
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			keysFile, err := os.CreateTemp(t.TempDir(), "keys*.txt")
			if err != nil {
				t.Errorf("failed to create temp file: %v", err)
				return
			}

			if err := keysFile.Close(); err != nil {
				t.Errorf("failed to close keys file after setup: %v", err)
				return
			}

			i := infrastructure.NewIdentities(keysFile.Name())

			if _, err := i.Generate(tt.args.comment); (err != nil) != tt.wantErr {
				t.Errorf("Generate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			keysFile, err = os.Open(keysFile.Name())
			if err != nil {
				t.Errorf("failed to open keys file: %v", err)
				return
			}

			t.Cleanup(func() {
				_ = keysFile.Close()
			})

			if _, err := age.ParseIdentities(keysFile); (err != nil) != tt.wantErr {
				t.Errorf("failed to parse keys file: %v", err)
				return
			}
		})
	}
}
