# Configuration

For now _git-age_ is configured either via environment variables or CLI flags.
The most interesting part is where it reads and writes the private keys from.
This can be configured via the `GIT_AGE_KEYS` environment variable or the `--keys` flag.
By default, _git-age_ will store the private keys in `$XDG_CONFIG_HOME/git-age/keys.txt`.

| Platform | Config path                                                               |
|----------|---------------------------------------------------------------------------|
| Linux    | `$XDG_CONFIG_HOME/git-age/keys.txt` i.e. `$HOME/.config/git-age.keys.txt` |
| macOS    | `$HOME/Library/Application Support/git-age/keys.txt`                      |
| Windows  | `%\LOCALAPPDATA%\git-age\keys.txt`                                        |

Additionally, _git-age_ can also look up identities with the help of an agent.
To use an agent set the `GIT_AGE_AGENT_HOST` environment variable to the corresponding endpoint.
The agent of your choice should tell you the value of this variable.