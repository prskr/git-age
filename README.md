# git-age

[![Go Report Card](https://goreportcard.com/badge/github.com/prskr/git-age)](https://goreportcard.com/report/github.com/prskr/git-age)
[![Go build](https://github.com/prskr/git-age/actions/workflows/go.yaml/badge.svg)](https://github.com/prskr/git-age/actions/workflows/go.yaml)

## Introduction

## Install

See [INSTALL.md](INSTALL.md).

## Getting started

### Init repository to share secret files

```bash
git age init

# or if you want to add some comment to the generated key
git age init -c "My comment"
```

### Add another user to an already prepared repository

__Remarks:__ The repository has to be in a clean state i.e. no changes files.

Alice wants to share the secrets stored in her Git repository with Bob.

1. Bob installs `git-age` on his machine and configures his global git config

```bash
git age install
```

2. Bob generates a new key pair

```bash
git age gen-key

# or if you want to add some comment to the generated key
git age gen-key -c "My comment"
```

the generated private key will be stored automatically in your `keys.txt`

3. Bob sends his public key to Alice
4. Alice adds Bob's public key to her repository

```bash
git age add-recipient <public key>

# or if you want to add some comment to the added key

git age add-recipient -c "My comment" <public key>
```

`git age add-recipient` will:

1. add the public key to the repository (`.agerecipients` file)
2. re-encrypt all files with the new set of recipients
3. commit the changes

As soon as Alice pushed the changes to the remote repository, Bob can pull the changes and decrypt the files.

## Tips and tricks

### Diff of text files

Set the `diff.age.textconv` git config to `cat` to see plain text diffs of encrypted files.

```bash
git config --global diff.age.textconv cat
```

## Configuration

For now `git-age` is configured either via environment variables or CLI flags.
The most interesting part is where it reads and writes the private keys from.
This can be configured via the `GIT_AGE_KEYS` environment variable or the `--keys` flag.
By default `git-age` will store the private keys in `$XDG_CONFIG_HOME/git-age/keys.txt`.

| Platform | Config path                                                               |
|----------|---------------------------------------------------------------------------|
| Linux    | `$XDG_CONFIG_HOME/git-age/keys.txt` i.e. `$HOME/.config/git-age.keys.txt` |
| macOS    | `$HOME/Library/Application Support/git-age/keys.txt`                      |
| Windows  | `%LOCALAPPDATA%\git-age\keys.txt`                                         |

## Development

### Required Tools

- [`golangci-lint`](https://golangci-lint.run/)
- [`husky`](https://github.com/go-courier/husky)
- `goimports`
- `gofumpt`
- *optionally*: [`goreleaser`](https://goreleaser.com/)
- *optionally*: [`dlv`](https://github.com/go-delve/delve)

### Install husky

Ensure `golangci-lint` and other checks are executed before commit.

```bash
go install github.com/go-courier/husky/cmd/husky@latest

husky init
```