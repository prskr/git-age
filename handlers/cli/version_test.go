package cli_test

import (
	"bytes"
	"fmt"
	"net/http"
	"testing"
	"testing/fstest"

	"github.com/alecthomas/kong"
	"github.com/prskr/git-age/core/ports"
	"github.com/prskr/git-age/handlers/cli"
	"github.com/stretchr/testify/assert"
)

//nolint:paralleltest // requires global state and cannot be executed in parallel
func TestVersionCliHandler_Run(t *testing.T) {
	type args struct {
		short            bool
		skipVersionCheck bool
	}
	tests := []struct {
		name           string
		args           args
		programVersion string
		latestVersion  string
		wantOut        string
	}{
		{
			name:    "Development version",
			wantOut: "Version is not set. This is a development build.\n",
		},
		{
			name: "Short version",
			args: args{
				short: true,
			},
			programVersion: "v0.1.0",
			wantOut:        "v0.1.0\n",
		},
		{
			name: "Skip version check",
			args: args{
				skipVersionCheck: true,
			},
			programVersion: "v0.1.0",
			wantOut:        "v0.1.0\nCommit: \nBuilt at \n",
		},
		{
			name:           "Run full - expect new version",
			programVersion: "v0.0.1",
			latestVersion:  "v0.2.0",
			wantOut:        "v0.0.1\nCommit: \nBuilt at \nA new version is available: 0.2.0 👉\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.programVersion != "" {
				cli.Version = tt.programVersion
				t.Cleanup(func() {
					cli.Version = cli.DefaultVersion
				})
			}

			handler := &cli.VersionCliHandler{
				Client: mockHttpClient(map[string][]byte{
					"repos/prskr/git-age/releases/latest": []byte(fmt.Sprintf(`{"tag_name": "%s"}`, tt.latestVersion)),
				}),
			}

			outBuf := new(bytes.Buffer)
			parser := newKong(
				t,
				handler,
				kong.BindTo(ports.STDOUT(outBuf), (*ports.STDOUT)(nil)),
			)

			var cliArgs []string
			if tt.args.short {
				cliArgs = append(cliArgs, "--short")
			}

			if tt.args.skipVersionCheck {
				cliArgs = append(cliArgs, "--skip-version-check")
			}

			kongCtx, err := parser.Parse(cliArgs)
			assert.NoError(t, err, "Failed to parse arguments: %v", err)

			assert.NoError(t, kongCtx.Run(), "failed to run command")

			gotOut := outBuf.String()
			assert.Equal(t, tt.wantOut, gotOut)
		})
	}
}

func mockHttpClient(files map[string][]byte) *http.Client {
	transportFS := make(fstest.MapFS)

	for k, v := range files {
		transportFS[k] = &fstest.MapFile{
			Data: v,
		}
	}

	return &http.Client{
		Transport: http.NewFileTransportFS(transportFS),
	}
}
