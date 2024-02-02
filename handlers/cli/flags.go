package cli

import (
	"path/filepath"

	"github.com/adrg/xdg"
	"github.com/urfave/cli/v2"
)

//nolint:gochecknoglobals // flags are used across multiple commands
var (
	//nolint:lll // flags require descriptions that aren't easily broken into multiple lines
	keysFlag = cli.StringFlag{
		Name:        "keys",
		DefaultText: "By default keys are read from $XDG_CONFIG_HOME/git-age/keys.txt i.e. $HOME/.config/git-age/keys.txt on most systems",
		Value:       filepath.Join(xdg.ConfigHome, "git-age", "keys.txt"),
		EnvVars: []string{
			"GIT_AGE_KEYS",
		},
	}

	commentFlag = cli.StringFlag{
		Name:    "comment",
		Aliases: []string{"c"},
		Usage:   "Comment to add in file",
	}
)
