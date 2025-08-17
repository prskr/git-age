# AGENTS.md

## Build/Lint/Test Commands

- **Build**: `go build -o git-age main.go`
- **Lint**: `golangci-lint run`
- **Test**: `go test ./...` or `go tool gotestsum --format pkgname --junitfile out/junit.xml -- -race -shuffle=on -covermode=atomic ./...`
- **Run single test**: `go test -v ./path/to/package -run TestFunctionName`
- **Check for dependency updates**: `go list -u -f '{{if (and (not (or .Main .Indirect)) .Update)}}{{.Path}}: {{.Version}} -> {{.Update.Version}}{{end}}' -m all`

## Code Style Guidelines

- Go code follows standard idioms with tabs for indentation in Go files
- Imports are organized with `gci` tool following standard, default, local-prefix order
- All files must have a license header
- Error handling follows Go patterns with explicit error checking
- Function names use camelCase
- Package naming follows Go conventions (lowercase, no underscores)
- Struct fields and variables use camelCase
- Constants use UPPER_SNAKE_CASE
- Tests follow the `TestFunctionName` naming convention
- Use `//nolint` comments for specific linter suppressions
- Prefer explicit error messages over generic ones
- Use `github.com/stretchr/testify` for assertions in tests