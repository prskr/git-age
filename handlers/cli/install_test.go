package cli_test

import (
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"testing"

	"gopkg.in/ini.v1"

	"github.com/prskr/git-age/handlers/cli"
	"github.com/prskr/git-age/internal/fsx"
)

func TestInstallCliHandler_Run(t *testing.T) {
	tests := []struct {
		name           string
		srcFile        string
		expectedValues map[string]string
	}{
		{
			name:    "install git-age filter in base git config",
			srcFile: "base_git_config.ini",
			expectedValues: map[string]string{
				"clean":    "git-age clean -- %f",
				"smudge":   "git-age smudge -- %f",
				"required": strconv.FormatBool(true),
			},
		},
		{
			name:    "Don't touch already configured git config",
			srcFile: "pre_configured_git_config.ini",
			expectedValues: map[string]string{
				"clean":    "age cat",
				"smudge":   "age smudge",
				"required": strconv.FormatBool(false),
			},
		},
	}
	// pre_configured_git_config.ini
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testHome := t.TempDir()
			t.Setenv(userHomeEnvVariable(), testHome)

			err := fsx.CopyFile(filepath.Join("testdata", tt.srcFile), filepath.Join(testHome, ".gitconfig"))
			if err != nil {
				t.Errorf("failed to copy file: %v", err)
				return
			}

			parser := newKong(t, new(cli.InstallCliHandler))
			ctx, err := parser.Parse(nil)
			if err != nil {
				t.Errorf("failed to parse arguments: %v", err)
				return
			}

			if err := ctx.Run(); err != nil {
				t.Errorf("failed to run command: %v", err)
				return
			}

			updatedGitConfig, err := os.ReadFile(filepath.Join(testHome, ".gitconfig"))
			if err != nil {
				t.Errorf("failed to read file: %v", err)
				return
			}

			cfg, err := ini.Load(updatedGitConfig)
			if err != nil {
				t.Errorf("failed to parse ini: %v", err)
				return
			}

			filterSection := cfg.Section(`filter "age"`)

			if filterSection == nil {
				t.Errorf("expected section to exist")
				return
			}

			for key, expectedValue := range tt.expectedValues {
				actualValue := getKey(t, filterSection, key)
				if actualValue != expectedValue {
					t.Errorf("expected %q, got %q", expectedValue, actualValue)
				}
			}
		})
	}
}

func getKey(tb testing.TB, section *ini.Section, keyName string) string {
	tb.Helper()

	key, err := section.GetKey(keyName)
	if err != nil {
		tb.Fatalf("failed to get key: %v", err)
	}

	return key.String()
}

func userHomeEnvVariable() string {
	switch runtime.GOOS {
	case "windows":
		return "USERPROFILE"
	default:
		return "HOME"
	}
}
