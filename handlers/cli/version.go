package cli

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Masterminds/semver/v3"

	"github.com/prskr/git-age/core/ports"
)

const DefaultVersion = "v0.0.0"

//nolint:gochecknoglobals // are set during build
var (
	Version = DefaultVersion
	Commit  = ""
	Date    = ""
)

type VersionCliHandler struct {
	Short            bool         `help:"Print only the version"`
	SkipVersionCheck bool         `help:"Skip version check"`
	Client           *http.Client `kong:"-"`
}

func (h VersionCliHandler) Run(stdout ports.STDOUT) error {
	if h.Short {
		_, _ = fmt.Fprintln(stdout, Version)
		return nil
	}

	if Version == DefaultVersion {
		_, _ = fmt.Fprintln(stdout, "Version is not set. This is a development build.")

		return nil
	}

	_, _ = fmt.Fprintf(stdout,
		`%s
Commit: %s
Built at %s
`, Version, Commit, Date)

	if h.SkipVersionCheck {
		return nil
	}

	currentVersion, err := semver.NewVersion(Version)
	if err != nil {
		return fmt.Errorf("failed to parse current version: %w", err)
	}

	release, err := h.checkLatestVersion()
	if err != nil {
		return err
	}

	if result := release.TagName.Compare(currentVersion); result > 0 {
		_, _ = fmt.Fprintf(stdout, "A new version is available: %s 👉\n", release.TagName)
	} else if result == 0 {
		_, _ = fmt.Fprintln(stdout, "You are using the latest version 🎉")
	} else {
		_, _ = fmt.Fprintln(stdout, "You're using a version that is somewhat newer than the latest release! 👻")
	}

	return nil
}

func (h VersionCliHandler) checkLatestVersion() (*releaseInfo, error) {
	client := h.Client
	if client == nil {
		client = http.DefaultClient
	}

	resp, err := client.Get("https://api.github.com/repos/prskr/git-age/releases/latest")
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	//nolint:err113 // no need to wrap - there's no point to wrap this error
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	decoder := json.NewDecoder(resp.Body)
	var release releaseInfo

	if err := decoder.Decode(&release); err != nil {
		return nil, err
	}

	return &release, nil
}

type releaseInfo struct {
	TagName semver.Version `json:"tag_name"`
}
