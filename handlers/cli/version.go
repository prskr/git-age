package cli

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/urfave/cli/v2"
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
	Client *http.Client
}

func (h VersionCliHandler) Version(ctx *cli.Context) error {
	if ctx.Bool("short") {
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

	if ctx.Bool("skip-version-check") {
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

func (h VersionCliHandler) Command() *cli.Command {
	return &cli.Command{
		Name:   "version",
		Action: h.Version,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name: "short",
			},
			&cli.BoolFlag{
				Name: "skip-version-check",
			},
		},
	}
}

type releaseInfo struct {
	TagName string `json:"tag_name"`
}
