package cli

//nolint:lll // flags require descriptions that aren't easily broken into multiple lines
type KeysFlag struct {
	Keys string `env:"GIT_AGE_KEYS" type:"path" short:"k" default:"${XDG_CONFIG_HOME}${file_path_separator}git-age${file_path_separator}keys.txt"`
}

type CommentFlag struct {
	Comment string `short:"c" help:"Comment to add in file"`
}
