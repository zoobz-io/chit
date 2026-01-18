# Contributing to chit

Thank you for your interest in contributing to chit! This guide will help you get started.

## Code of Conduct

By participating in this project, you agree to maintain a respectful and inclusive environment for all contributors.

## Getting Started

1. Fork the repository
2. Clone your fork: `git clone https://github.com/yourusername/chit.git`
3. Create a feature branch: `git checkout -b feature/your-feature-name`
4. Make your changes
5. Run tests: `make test`
6. Commit your changes with a descriptive message
7. Push to your fork: `git push origin feature/your-feature-name`
8. Create a Pull Request

## Development Guidelines

### Code Style

- Follow standard Go conventions
- Run `go fmt` before committing
- Add comments for exported functions and types
- Keep functions small and focused

### Testing

- Write tests for new functionality
- Ensure all tests pass: `make test`
- Include benchmarks for performance-critical code
- Aim for good test coverage

### Documentation

- Update documentation for API changes
- Add examples for new features
- Keep doc comments clear and concise

## Types of Contributions

### Bug Reports

- Use GitHub Issues
- Include minimal reproduction code
- Describe expected vs actual behavior
- Include Go version and OS

### Feature Requests

- Open an issue for discussion first
- Explain the use case
- Consider backwards compatibility

### Code Contributions

#### Adding Primitives

New chat primitives should:
- Implement `pipz.Chainable[*Chat]`
- Follow the existing pattern (builder methods, capitan signals)
- Emit appropriate signals for observability
- Include comprehensive tests
- Add documentation with examples

#### Examples

New examples should:
- Demonstrate a real-world chatbot scenario
- Include tests
- Have descriptive comments
- Follow the existing structure

## Pull Request Process

1. **Keep PRs focused** - One feature/fix per PR
2. **Write descriptive commit messages**
3. **Update tests and documentation**
4. **Ensure CI passes**
5. **Respond to review feedback**

## Testing

Run the full test suite:
```bash
make test
```

Run with race detection:
```bash
go test -race ./...
```

Run benchmarks:
```bash
make bench
```

## Project Structure

```
chit/
├── *.go              # Core library files (primitives, chat, emitter)
├── *_test.go         # Tests
├── testing/          # Test helpers and integration tests
└── docs/             # Documentation
```

## Commit Messages

Follow conventional commits:
- `feat:` New feature
- `fix:` Bug fix
- `docs:` Documentation changes
- `test:` Test additions/changes
- `refactor:` Code refactoring
- `perf:` Performance improvements
- `chore:` Maintenance tasks

## Questions?

- Open an issue for questions
- Check existing issues first
- Be patient and respectful

Thank you for contributing to chit!
