# Contributing to myshift-go

Thank you for your interest in contributing to myshift-go! This document provides guidelines and instructions for contributing to the project.

## Code of Conduct

By participating in this project, you agree to maintain a respectful and inclusive environment for everyone.

## Development Environment

### Prerequisites

- Go 1.21 or higher
- Git
- A PagerDuty account with API access

### Setting Up

1. Fork the repository
2. Clone your fork:
   ```bash
   git clone https://github.com/YOUR_USERNAME/myshift-go.git
   cd myshift-go
   ```
3. Download dependencies:
   ```bash
   go mod download
   ```
4. Build the project:
   ```bash
   go build -o myshift cmd/myshift/main.go
   ```

## Development Workflow

1. Create a new branch for your feature or bugfix:
   ```bash
   git checkout -b feature/your-feature-name
   # or
   git checkout -b fix/your-bugfix-name
   ```

2. Make your changes, following the coding standards below

3. Run tests:
   ```bash
   go test ./...
   ```

4. Run tests with coverage:
   ```bash
   go test -coverprofile=coverage.out ./...
   go tool cover -html=coverage.out
   ```

5. Format your code:
   ```bash
   go fmt ./...
   ```

6. Lint your code:
   ```bash
   go vet ./...
   ```

7. Commit your changes using semantic commit messages (see below)

8. Push to your fork:
   ```bash
   git push origin feature/your-feature-name
   ```

9. Create a Pull Request from your fork to the main repository

## Coding Standards

### Go Code Style

- Follow [Effective Go](https://golang.org/doc/effective_go.html) guidelines
- Use `gofmt` to format your code
- Use `go vet` to catch common mistakes
- Write clear, self-documenting code
- Use meaningful variable and function names
- Keep functions focused and small
- Add comments for exported functions and types
- Follow Go naming conventions (e.g., `camelCase` for unexported, `PascalCase` for exported)
- Use interfaces to define behavior contracts
- Handle errors explicitly; don't ignore them

### Package Structure

- Follow the project's existing package structure:
  - `cmd/`: Main application entry points
  - `internal/`: Private application code
  - `pkg/`: Public library code
  - Keep packages focused on a single responsibility

### Commit Messages

We use [Conventional Commits](https://www.conventionalcommits.org/) for our commit messages. This helps maintain a clear and consistent history of changes.

#### Format

```
<type>(<scope>): <description>

[optional body]

[optional footer(s)]
```

#### Types

- `feat`: A new feature
- `fix`: A bug fix
- `docs`: Documentation only changes
- `style`: Changes that do not affect the meaning of the code
- `refactor`: A code change that neither fixes a bug nor adds a feature
- `perf`: A code change that improves performance
- `test`: Adding missing tests or correcting existing tests
- `chore`: Changes to the build process or auxiliary tools
- `ci`: Changes to CI configuration files and scripts

#### Scopes

- `cli`: Command-line interface changes
- `config`: Configuration file handling
- `api`: PagerDuty API integration
- `commands`: Command implementations
- `formatters`: Output formatting changes
- `container`: Container-related changes
- `deps`: Dependency updates
- `docs`: Documentation changes

#### Examples

```
feat(commands): add iCal formatter support to plan and upcoming commands

Adds support for exporting schedule data in iCalendar format for import
into calendar applications. Both commands now support --format/-o option
with text and ical formats.

- Add PlanFormatter interface for extensible output formatting
- Implement TextFormatter and ICalFormatter
- Add -o short option support for format selection
- Update CLI help text and usage examples

Closes #123
```

```
fix(api): correct schedule_ids parameter format for PagerDuty API

PagerDuty API expects array parameters to have [] suffix. Changed
schedule_ids to schedule_ids[] in all API calls.

- Fix next command API calls
- Fix upcoming command API calls
- Fix plan command API calls
- Fix override command API calls

Fixes #456
```

```
docs(readme): update build and installation instructions

- Add Go version requirements
- Update build commands for Go modules
- Add examples for different platforms
- Include formatter usage examples
```

#### Guidelines

1. Use the present tense ("Add feature" not "Added feature")
2. Use the imperative mood ("Move cursor to..." not "Moves cursor to...")
3. Limit the first line to 72 characters or less
4. Reference issues and pull requests liberally after the first line
5. Consider starting the commit message with an applicable emoji:
   - 🎨 `:art:` when improving the format/structure of the code
   - 🐎 `:racehorse:` when improving performance
   - 🚱 `:non-potable_water:` when plugging memory leaks
   - 📝 `:memo:` when writing docs
   - 🐛 `:bug:` when fixing a bug
   - 🔥 `:fire:` when removing code or files
   - 💚 `:green_heart:` when fixing the CI build
   - ✅ `:white_check_mark:` when adding tests
   - 🔒 `:lock:` when dealing with security
   - ⬆️ `:arrow_up:` when upgrading dependencies
   - ⬇️ `:arrow_down:` when downgrading dependencies

### Documentation

- Update README.md if you add new features or change existing ones
- Document all new command-line options
- Add godoc comments for exported functions, types, and packages
- Include examples in documentation and comments
- Update TESTING.md if test procedures change

## Testing

- Write tests for new features using Go's testing package
- Ensure all tests pass before submitting a PR:
  ```bash
  go test ./...
  ```
- Include both unit tests and integration tests where appropriate
- Test edge cases and error conditions
- Use table-driven tests for multiple test cases
- Mock external dependencies (like PagerDuty API calls)
- Aim for good test coverage:
  ```bash
  go test -coverprofile=coverage.out ./...
  ```

### Test Structure

- Place tests in `*_test.go` files alongside the code they test
- Use descriptive test names that explain what is being tested
- Follow the pattern: `TestFunctionName_Scenario`
- Use `testify` assertions if needed for better error messages

## Pull Request Process

1. Update the README.md with details of changes if needed
2. Update the documentation with any new command-line options
3. Ensure all tests pass and code is properly formatted
4. The PR will be reviewed by maintainers
5. Address any feedback or requested changes
6. Once approved, your PR will be merged

## Adding New Features

When adding new features:

1. Discuss the feature in an issue first
2. Follow the existing code structure and patterns
3. Add appropriate tests
4. Update documentation
5. Add examples of usage
6. Consider backward compatibility
7. Follow Go interfaces and design patterns

### Adding New Commands

When adding new commands:

1. Create the command in `internal/commands/`
2. Add the command handler to `cmd/myshift/main.go`
3. Follow the existing command pattern with proper error handling
4. Add comprehensive tests
5. Update help text and documentation

### Adding New Formatters

When adding new output formatters:

1. Implement the `PlanFormatter` interface
2. Add the formatter to `GetFormatter()` function
3. Add tests for the new formatter
4. Update documentation with usage examples

## Reporting Bugs

When reporting bugs, please include:

1. A clear description of the bug
2. Steps to reproduce
3. Expected behavior
4. Actual behavior
5. Environment details (OS, Go version, etc.)
6. Any relevant error messages or logs
7. Configuration file content (with sensitive data redacted)

## Questions and Support

If you have questions or need help:

1. Check the existing documentation
2. Search existing issues
3. Create a new issue if needed

## Release Process

Releases follow semantic versioning (SemVer):

- `MAJOR.MINOR.PATCH`
- MAJOR: incompatible API changes
- MINOR: functionality added in backward-compatible manner
- PATCH: backward-compatible bug fixes

## License

By contributing to myshift-go, you agree that your contributions will be licensed under the project's Apache License 2.0. 