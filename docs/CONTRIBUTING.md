# Contributing to portcheck

First off, thank you for considering contributing to portcheck!

## Development Setup

### Prerequisites

- Go 1.22+
- Make (optional, but recommended)
- golangci-lint (for linting)

### Getting Started

```bash
# Clone the repo
git clone https://github.com/uncaughtx/portcheck.git
cd portcheck

# Install dependencies
go mod download

# Build
make build

# Run tests
make test

# Run linter
make lint
```

## Making Changes

1. **Fork** the repository
2. **Create a branch** for your feature (`git checkout -b feature/amazing-feature`)
3. **Make your changes** with clear, descriptive commits
4. **Run tests** (`make test`)
5. **Run linter** (`make lint`)
6. **Push** to your fork (`git push origin feature/amazing-feature`)
7. **Open a Pull Request**

## Commit Messages

Would appreciate if you could follow meaningful commit messages:

```
feat: add UDP port scanning support
fix: handle permission denied on Linux gracefully
docs: add Windows installation instructions
```

## Code Style

- Run `gofmt` and `goimports` before committing
- Follow standard Go conventions
- Add comments for exported functions
- Keep functions focused and testable

## Testing

- Add tests for new features
- Ensure existing tests pass
- Test cross-platform when possible

```bash
# Run all tests
make test

# Run with coverage
make test-cover
```

## Pull Request Guidelines

- Fill out the PR template
- Link related issues
- Include screenshots for UI changes
- Request review from maintainers

## Questions?

Open a [Discussion](https://github.com/uncaughtx/portcheck/discussions) or reach out to [@uncaughtx](https://github.com/uncaughtx).

---

Thank you for contributing!
