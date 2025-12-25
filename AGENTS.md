# AGENTS.md

## Build/Test Commands
- **Build**: `go build ./cmd/lgo`
- **Test all**: `go test ./...`
- **Test single**: `go test -run TestName ./package/path`
- **Format**: `go fmt ./...`
- **Lint**: `go vet ./...`
- **Note**: If `go test` fails with permissions in `/home/jason/.cache/go-build`, use `GOCACHE=/home/jason/go/.cache/go-build`.
- **Note**: If `git commit` fails during pre-commit with `PermissionError` opening a PTY or writing `/home/jason/.cache/pre-commit`, document the failure and proceed with `git commit --no-verify`.

## Code Style Guidelines

### Imports
- Standard library imports first, then third-party, then local
- Group imports by blank lines between groups
- Use blank imports only when required for side effects

### Formatting
- Use `go fmt` for consistent formatting
- Go uses tabs for indentation (gofmt standard)
- Line length: reasonable, break long lines

### Naming
- **Exported**: PascalCase (NewParser, Transaction)
- **Unexported**: camelCase (cloverLeafParser, parseText)
- **Files**: snake_case with _test.go suffix for tests
- **Packages**: lowercase, single word when possible

### Types & Interfaces
- Define interfaces in dedicated files (interface.go)
- Use struct types with clear field names
- Embed interfaces when extending functionality

### Error Handling
- Return errors from functions that can fail
- Use slog for structured logging
- Check errors immediately after operations

### Testing
- Golden file testing for complex outputs
- Table-driven tests where appropriate
- Test files in same package (_test)
- Use t.Helper() for test helpers

### Logging
- Use slog for structured logging
- Include context in log messages
- Debug level for development, Info for important operations
