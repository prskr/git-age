package cli

func KeysFlagOf(keys string) KeysFlag {
	return KeysFlag{Keys: keys}
}

//nolint:lll // flags require descriptions that aren't easily broken into multiple lines
type KeysFlag struct {
	Keys string `env:"GIT_AGE_KEYS" type:"existingfile" short:"k" default:"${XDG_CONFIG_HOME}${file_path_separator}git-age${file_path_separator}keys.txt"`
}

func CommentFlagOf(comment string) CommentFlag {
	return CommentFlag{Comment: comment}
}

type CommentFlag struct {
	Comment string `short:"c" help:"Comment to add in file"`
}
