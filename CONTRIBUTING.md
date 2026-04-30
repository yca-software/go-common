# Contributing

Thanks for helping improve `go-common`.

## Development Setup

1. Ensure Go is installed:
   - Go `1.26.2` or newer
2. Run tests:
   ```bash
   go test ./... -count=1
   ```

Some integration tests use testcontainers; ensure Docker is available when running full test suites.

## Pull Request Guidelines

- Keep PRs focused and reviewable.
- Add or update tests for behavior changes.
- Update package README files when API or behavior changes.
- Avoid introducing breaking changes without migration notes.

## Code Style

- Follow existing package patterns.
- Prefer small, composable functions and explicit error handling.
