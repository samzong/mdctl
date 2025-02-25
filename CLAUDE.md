# CLAUDE.md - mdctl Project Guide

## Build Commands
- `make build` - Build the binary
- `make test` - Run all tests
- `make fmt` - Format code and tidy modules
- `make clean` - Clean up build artifacts
- `make all` - Run clean, fmt, build, and test

## Code Style Guidelines
- Use Go standard formatting (`gofmt`)
- Organize imports alphabetically with standard library first
- Use camelCase for variable names, PascalCase for exported functions/types
- Return errors rather than using panics for recoverable situations
- Document all exported functions and types with comments
- Keep functions small and focused on a single responsibility
- Use the Cobra library for CLI commands, following its standard patterns
- Error messages should be descriptive and start with lowercase
- Implement proper error handling with contextual information

## Structure
- Place core logic in the `internal/` directory
- Keep command definitions in the `cmd/` directory
- Minimize dependencies and prefer standard library when possible