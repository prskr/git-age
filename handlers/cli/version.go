package cli

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"golang.org/x/mod/semver"
)

const defaultVersion = "v0.0.0"

//nolint:gochecknoglobals // are set during build
var (
	version = defaultVersion
	commit  = ""
	date    = time.Now().UTC().Format(time.RFC3339)
)

type VersionCliHandler struct {
	Short            bool         `help:"Print only the version"`
	SkipVersionCheck bool         `help:"Skip version check"`
	Client           *http.Client `kong:"-"`
}

func (h VersionCliHandler) Run() error {
	if h.Short {
		fmt.Println(version)
		return nil
	}

	if version == defaultVersion {
		fmt.Println("Version is not set. This is a development build.")

		return nil
	}

	fmt.Printf(`%s
Commit: %s
Built at %s
`, version, commit, date)

	if h.SkipVersionCheck {
		return nil
	}

	client := h.Client
	if client == nil {
		client = http.DefaultClient
	}

	resp, err := client.Get("https://api.github.com/repos/prskr/git-age/releases/latest")
	if err != nil {
		return err
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	//nolint:goerr113 // no need to wrap - there's no point to wrap this error
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	decoder := json.NewDecoder(resp.Body)
	var release releaseInfo

	if err := decoder.Decode(&release); err != nil {
		return err
	}

	if result := semver.Compare(release.TagName, version); result > 0 {
		fmt.Printf("A new version is available: %s 👉\n", release.TagName)
	} else if result == 0 {
		fmt.Println("You are using the latest version 🎉")
	} else {
		fmt.Println("You're using a version that is somewhat newer than the latest release! 👻")
	}

	return nil
}

type releaseInfo struct {
	TagName string `json:"tag_name"`
}
