package infrastructure_test

import (
	"strings"
	"testing"

	"filippo.io/age"
	"github.com/prskr/git-age/core/ports"
	"github.com/prskr/git-age/infrastructure"
	"github.com/prskr/git-age/internal/testfs"
)

func TestRecipientsFile_All(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		setup   func(tb testing.TB, fs testfs.TestFS)
		wantNum int
		wantErr bool
	}{
		{
			name: "empty recipients file",
		},
		{
			name: "garbage in file",
			setup: func(tb testing.TB, fs testfs.TestFS) {
				tb.Helper()
				if err := fs.Add(ports.RecipientsFileName, []byte("garbage")); err != nil {
					tb.Fatalf("failed to write garbage to file: %v", err)
				}
			},
			wantErr: true,
		},
		{
			name: "single recipient",
			setup: func(tb testing.TB, fs testfs.TestFS) {
				tb.Helper()
				id, err := age.GenerateX25519Identity()
				if err != nil {
					tb.Fatalf("failed to create age identity: %v", err)
				}

				if err := fs.Add(ports.RecipientsFileName, []byte(id.Recipient().String())); err != nil {
					tb.Fatalf("failed to write identity to file: %v", err)
				}
			},
			wantNum: 1,
		},
		{
			name: "multiple recipients",
			setup: func(tb testing.TB, fs testfs.TestFS) {
				tb.Helper()
				publicKeys := make([]string, 0, 5)
				for i := 0; i < 5; i++ {
					id, err := age.GenerateX25519Identity()
					if err != nil {
						tb.Fatalf("failed to create age identity: %v", err)
					}

					publicKeys = append(publicKeys, id.Recipient().String())
				}
				if err := fs.Add(ports.RecipientsFileName, []byte(strings.Join(publicKeys, "\n"))); err != nil {
					tb.Fatalf("failed to write identity to file: %v", err)
				}
			},
			wantNum: 5,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tfs := testfs.NewTestFS()

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
