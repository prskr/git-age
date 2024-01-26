package cli

import (
	"filippo.io/age"
	"github.com/adrg/xdg"
	"github.com/urfave/cli/v2"
	"io"
	"os"
	"path/filepath"
)

type SmudgeCliHandler struct {
	baseHandler
}

func (h *SmudgeCliHandler) SmudgeFile(*cli.Context) error {
	if err := requireStdin(); err != nil {
		return err
	}

	decryptedReader, err := age.Decrypt(os.Stdin, h.Identities...)
	if err != nil {
		return err
	}

	_, err = io.Copy(os.Stdout, decryptedReader)

	return err
}

func (h *SmudgeCliHandler) Command() *cli.Command {
	return &cli.Command{
		Name:   "smudge",
		Usage:  "smudge should only be invoked by Git",
		Action: h.SmudgeFile,
		Args:   true,
		Hidden: true,
		Before: func(context *cli.Context) error {
			keysPath := filepath.Join(xdg.ConfigHome, "git-age", "keys.txt")
			if flagPath := context.String("keys"); flagPath != "" {
				keysPath = flagPath
			}

			return h.AddIdentitiesFromPath(keysPath)
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "keys",
				DefaultText: "By default keys are read from $XDG_CONFIG_HOME/git-age/keys.txt i.e. $HOME/.config/git-age/keys.txt on most systems",
				EnvVars: []string{
					"GIT_AGE_KEYS",
				},
			},
		},
	}
}
