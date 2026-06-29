## Contributing

Thank you for considering contributing to dcvix-launcher.

### Getting Started

1. Fork the repository.
2. Ensure you have the build requirements from the README.
3. Run `go mod tidy` and `make build-linux` to verify your environment works.

### Code Style

- Follow [Effective Go](https://go.dev/doc/effective_go) conventions.
- Keep dependencies minimal, prefer the standard library and Fyne's built-in widgets.
- Use `fyne.Do()` to schedule UI updates from goroutines.

### Structure

- `cmd/dcvix-launcher/main.go` - application entry point.
- `internal/` - all application logic, organized by concern:
  - `client/` - broker communication.
  - `config/` - INI configuration parsing.
  - `gui/` - Fyne UI and components.
  - `logger/` - logging setup.
  - `service/` - business logic glue.
  - `version/` - build version metadata.
- Shared logic must be extracted into the appropriate package, no duplication.

### Pull Requests

- Keep changes focused, one feature or fix per PR.
- Run `make build-linux` before submitting to confirm the project compiles.
- Write clear commit messages following the existing style.
- Ensure new code follows the package conventions.
