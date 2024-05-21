package infrastructure_test

import (
	"context"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"filippo.io/age"
	"github.com/stretchr/testify/assert"

	"github.com/prskr/git-age/core/dto"
	"github.com/prskr/git-age/infrastructure"
	"github.com/prskr/git-age/internal/testx"
)

const (
	singleIdentity = `# Tests
# public key: age1g5h29jjf0c69s7z86nrtd997un6z7zcq54x7l2a6j27745h5p5lqsmklq9
AGE-SECRET-KEY-1K2WD2SE8TUA0FYJ3768W9JLYVUM6M7KHMW2TKWMV6VMCH9ESG52QRAAYNW
`

	multipleIdentities = `# Test 1
# public key: age1g5h29jjf0c69s7z86nrtd997un6z7zcq54x7l2a6j27745h5p5lqsmklq9
AGE-SECRET-KEY-1K2WD2SE8TUA0FYJ3768W9JLYVUM6M7KHMW2TKWMV6VMCH9ESG52QRAAYNW
# Test 2
# public key: age1a975r8q6gylt6vu5jugert3faj3s5a0jwwlaa7zw033zhqg85clsu5u6kh
AGE-SECRET-KEY-10U5FFM4YVSAWHL4W2PZXJJKQULZU0TDMA8W3H79YMGL5DWDKN8DQZV77VU
`
)

func TestFileIdentityStoreSource_IsValid(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		url     *url.URL
		want    bool
		wantErr bool
	}{
		{
			name:    "Nil URL",
			url:     nil,
			want:    false,
			wantErr: false,
		},
		{
			name:    "Missing path",
			url:     new(url.URL),
			want:    false,
			wantErr: false,
		},
		{
			name: "Valid URL - without scheme",
			url: &url.URL{
				Path: "/home/prskr/.config/git-age/keys.txt",
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "Valid URL - with scheme",
			url: &url.URL{
				Scheme: "file:/",
				Path:   "/home/prskr/.config/git-age/keys.txt",
			},
			want:    true,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			src := infrastructure.NewFileIdentityStoreSource(tt.url)
			got, err := src.IsValid(testx.Context(t))
			if (err != nil) != tt.wantErr {
				t.Errorf("IsValid() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			assert.Equal(t, tt.want, got, "IsValid() got = %v, want %v", got, tt.want)
		})
	}
}

func TestFileIdentityStore_Identities(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		keysContent string
		want        func(tb testing.TB, ids []age.Identity)
		wantErr     bool
	}{
		{
			name:    "Non-existing file",
			wantErr: false,
		},
		{
			name:        "Existing file - single identity",
			keysContent: singleIdentity,
			want: func(tb testing.TB, ids []age.Identity) {
				tb.Helper()
				if len(ids) < 1 {
					t.Errorf("expected at least one identity, got %d", len(ids))
				}
			},
			wantErr: false,
		},
		{
			name:        "Existing file - multiple identities",
			keysContent: multipleIdentities,
			want: func(tb testing.TB, ids []age.Identity) {
				tb.Helper()
				if len(ids) != 2 {
					t.Errorf("expected at least one identity, got %d", len(ids))
				}
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tmpDir := t.TempDir()

			keysFilePath := filepath.Join(tmpDir, "keys.txt")
			if tt.keysContent != "" {
				assert.NoError(
					t,
					os.WriteFile(keysFilePath, []byte(tt.keysContent), 0o400),
					"failed to write keys file",
				)
			}

			store, err := infrastructure.NewFileIdentityStoreSource(&url.URL{Path: keysFilePath}).GetStore()
			assert.NoError(t, err, "failed to get store")

			got, err := store.Identities(testx.Context(t), dto.IdentitiesQuery{})
			if (err != nil) != tt.wantErr {
				t.Errorf("Identities() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.want != nil {
				tt.want(t, got)
			}
		})
	}
}

func TestFileIdentityStore_Generate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		keysContent string
		cmd         dto.GenerateIdentityCommand
		wantErr     bool
	}{
		{
			name:    "Generate new identity - empty file",
			cmd:     dto.GenerateIdentityCommand{},
			wantErr: false,
		},
		{
			name: "Generate new identity with comment - empty file",
			cmd: dto.GenerateIdentityCommand{
				Comment: "test",
			},
			wantErr: false,
		},
		{
			name: "Generate new identity with comment and remote - empty file",
			cmd: dto.GenerateIdentityCommand{
				Comment: "test",
				Remote:  "https://github.com/prskr/git-age",
			},
			wantErr: false,
		},
		{
			name:        "Generate new identity - non-empty file",
			cmd:         dto.GenerateIdentityCommand{},
			keysContent: multipleIdentities,
			wantErr:     false,
		},
		{
			name: "Generate new identity with comment - non-empty file",
			cmd: dto.GenerateIdentityCommand{
				Comment: "test",
			},
			keysContent: multipleIdentities,
			wantErr:     false,
		},
		{
			name: "Generate new identity with comment and remote - non-empty file",
			cmd: dto.GenerateIdentityCommand{
				Comment: "test",
				Remote:  "https://github.com/prskr/git-age",
			},
			keysContent: multipleIdentities,
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tmpDir := t.TempDir()

			keysFilePath := filepath.Join(tmpDir, "keys.txt")
			if tt.keysContent != "" {
				assert.NoError(
					t,
					os.WriteFile(keysFilePath, []byte(tt.keysContent), 0o600),
					"failed to write keys file",
				)
			}

			store, err := infrastructure.NewFileIdentityStoreSource(&url.URL{Path: keysFilePath}).GetStore()
			if err != nil {
				assert.NoError(t, err, "failed to get store")
			}

			pubKey, err := store.Generate(context.Background(), tt.cmd)
			if (err != nil) != tt.wantErr {
				t.Errorf("Generate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if _, err := age.ParseX25519Recipient(pubKey); err != nil {
				assert.NoError(t, err, "failed to parse public key")
			}

			keysFile, err := os.Open(keysFilePath)
			assert.NoError(t, err, "failed to open keys file")

			t.Cleanup(func() {
				_ = keysFile.Close()
			})

			if _, err := age.ParseIdentities(io.TeeReader(keysFile, testWriter(t))); err != nil {
				assert.NoError(t, err, "failed to parse identities")
			}
		})
	}
}

func testWriter(tb testing.TB) io.Writer {
	tb.Helper()
	return writerFunc(func(p []byte) (n int, err error) {
		tb.Helper()
		tb.Log(string(p))
		return len(p), nil
	})
}

type writerFunc func(p []byte) (n int, err error)

func (w writerFunc) Write(p []byte) (n int, err error) {
	return w(p)
}
