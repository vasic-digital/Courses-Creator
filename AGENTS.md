# Course Creator - Agent Guide

## Build/Lint/Test Commands

### Go Backend (core-processor/)
- **All tests**: `go test ./...`
- **Single test**: `go test -run TestFunctionName ./path/to/package`
- **Integration tests**: `go test ./tests/integration/`
- **Unit tests**: `go test ./tests/unit/`
- **Contract tests**: `go test ./tests/contract/`

### TypeScript Apps (creator-app/, mobile-player/)
- **Lint**: `npm run lint`
- **Format**: `npm run format`
- **All tests**: `npm test`
- **Single test**: `npm test -- --testNamePattern="test name"`
- **Watch mode**: `npm test -- --watch`

### Python (core-processor/)
- **Tests with coverage**: `pytest --cov=src --cov-report=term-missing`

## Code Style Guidelines

### Go
- **Naming**: PascalCase for exported types/functions, camelCase for unexported
- **Error handling**: Explicit error returns, no panics in production code
- **Imports**: Group standard library, then third-party, then local packages
- **Testing**: testify with require/assert, Given/When/Then descriptions
- **Formatting**: `gofmt` standard formatting

### TypeScript
- **Strict mode**: Enabled - no implicit any, strict null checks
- **Components**: Functional with PascalCase names, explicit return types
- **Variables/Functions**: camelCase, descriptive names
- **Imports**: Group external libraries, then internal modules, then types
- **Async**: Use async/await, avoid Promises directly
- **Error handling**: Try/catch with specific error types

### General
- **TDD**: Write tests BEFORE implementation (100% coverage required)
- **Comments**: No comments unless complex business logic
- **File paths**: Use absolute paths for file operations
- **Security**: Never log/expose secrets, validate all inputs