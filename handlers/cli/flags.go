package cli

import "net/url"

//nolint:lll // flags require descriptions that aren't easily broken into multiple lines
type KeysFlag struct {
	Keys *url.URL `env:"GIT_AGE_KEYS" name:"keys" short:"k" default:"file://${XDG_CONFIG_HOME}${file_path_separator}git-age${file_path_separator}keys.txt"`
}

type CommentFlag struct {
	Comment string `short:"c" name:"comment" help:"Comment to add in file"`
}

type RemoteFlag struct {
	Remote string `short:"r" name:"remote" help:"Remote for which this key should be considered"`
}
