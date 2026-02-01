# Contributing to JellyWolProxy

Thank you for your interest in contributing! 🎉

## Development Setup

1. Fork and clone the repository
2. Install Go 1.23.5 or later
3. Install pre-commit hooks:
   ```bash
   pre-commit install
   ```
4. Run tests:
   ```bash
   go test ./...
   ```

## Commit Convention

We use [Conventional Commits](https://www.conventionalcommits.org/) for clear and structured commit messages.

### Format

```
<type>(<scope>): <subject>

<body>

<footer>
```

### Types

- **feat**: New feature
- **fix**: Bug fix
- **docs**: Documentation changes
- **style**: Code style changes (formatting, no logic change)
- **refactor**: Code refactoring (no functional change)
- **perf**: Performance improvements
- **test**: Adding or updating tests
- **build**: Build system changes
- **ci**: CI/CD changes
- **chore**: Other changes (dependencies, tooling)

### Examples

```bash
feat(wol): add retry mechanism for wake-on-lan packets
fix(proxy): handle connection timeout correctly
docs(readme): update docker installation instructions
ci(actions): add codecov integration
```

## Pull Request Process

1. Update tests and documentation
2. Ensure all tests pass: `go test ./...`
3. Run pre-commit checks: `pre-commit run --all-files`
4. Create a PR with a clear description
5. Link any related issues

## Code Style

- Follow Go conventions (gofmt, golangci-lint)
- Add tests for new features
- Keep functions small and focused
- Document exported functions

## Testing

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific test
go test -run TestFunctionName ./...
```

## Questions?

Open an issue or discussion if you need help!
