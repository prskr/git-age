package infrastructure_test

import (
	"strings"
	"testing"

	"github.com/prskr/git-age/internal/fsx"

	"filippo.io/age"

	"github.com/prskr/git-age/core/ports"
	"github.com/prskr/git-age/infrastructure"
)

func TestRecipientsFile_All(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		setup   func(tb testing.TB, fs ports.ReadWriteFS)
		wantNum int
		wantErr bool
	}{
		{
			name: "empty recipients file",
		},
		{
			name: "garbage in file",
			setup: func(tb testing.TB, rwfs ports.ReadWriteFS) {
				tb.Helper()

				if err := fsx.WriteTo(rwfs, ports.RecipientsFileName, []byte("garbage")); err != nil {
					tb.Fatalf("failed to write garbage to file: %v", err)
				}
			},
			wantErr: true,
		},
		{
			name: "single recipient",
			setup: func(tb testing.TB, rwfs ports.ReadWriteFS) {
				tb.Helper()
				id, err := age.GenerateX25519Identity()
				if err != nil {
					tb.Fatalf("failed to create age identity: %v", err)
				}

				if err := fsx.WriteTo(rwfs, ports.RecipientsFileName, []byte(id.Recipient().String())); err != nil {
					tb.Fatalf("failed to write identity to file: %v", err)
				}
			},
			wantNum: 1,
		},
		{
			name: "multiple recipients",
			setup: func(tb testing.TB, rwfs ports.ReadWriteFS) {
				tb.Helper()
				publicKeys := make([]string, 0, 5)
				for range 5 {
					id, err := age.GenerateX25519Identity()
					if err != nil {
						tb.Fatalf("failed to create age identity: %v", err)
					}

					publicKeys = append(publicKeys, id.Recipient().String())
				}
				if err := fsx.WriteTo(rwfs, ports.RecipientsFileName, []byte(strings.Join(publicKeys, "\n"))); err != nil {
					tb.Fatalf("failed to write identity to file: %v", err)
				}
			},
			wantNum: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tfs := infrastructure.NewReadWriteDirFS(t.TempDir())

			if tt.setup != nil {
				tt.setup(t, tfs)
			}

			r := infrastructure.NewRecipientsFile(tfs)
			got, err := r.All()
			if (err != nil) != tt.wantErr {
				t.Errorf("All() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(got) != tt.wantNum {
				t.Errorf("All() got = %v, want %v", len(got), tt.wantNum)
			}
		})
	}
}
