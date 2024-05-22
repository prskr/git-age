# Development

## Required Tools

- [`golangci-lint`](https://golangci-lint.run/)
- [`husky`](https://github.com/go-courier/husky)
- `goimports`
- `gofumpt`
- `asciidoctor`
- *optionally*: [`goreleaser`](https://goreleaser.com/)
- *optionally*: [`dlv`](https://github.com/go-delve/delve)

## Install husky

Ensure `golangci-lint` and other checks are executed before commit.

```bash
go install github.com/go-courier/husky/cmd/husky@latest

husky init
```