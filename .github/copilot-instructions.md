# MoneyPenny - Copilot Instructions

## Project Overview
MoneyPenny is a personal finance assistant CLI tool written in Go 1.24. It processes financial statements (e.g., Sparkasse CSV exports) to help with personal finance management.

## Architecture

### Directory Structure
- `cmd/cli/` - CLI entrypoint using Cobra command framework
- `cmd/cli/root/` - Root command and subcommand registration
- `internal/service/` - Business logic services (e.g., `sparkasse/` for bank-specific processing)
- `internal/log/` - Centralized logging using Zap with GCP-compatible formatting

### Key Patterns
- **Cobra CLI**: Commands are defined in `cmd/cli/root/`. New subcommands should be registered in `init()` functions using `rootCmd.AddCommand()`.
- **Logging**: Always use `log.GetLogger()` from `internal/log` - never instantiate loggers directly. Logger must be initialized via `log.SetupLogging()` in main.
- **Service layer**: Bank-specific logic lives in `internal/service/<bank>/`. Each bank processor should implement statement processing functions.

## Developer Workflow

### Essential Commands
```bash
make install-deps  # First-time setup: installs gotestsum, revive, staticcheck
make run           # Run the CLI directly
make build         # Full build with all checks (runs checks → builds binaries)
make test          # Run tests with race detection
make checks        # Lint + vet + staticcheck (runs before build/test)
```

### Code Quality
- **Linter**: `revive` with custom rules in [lintconfig.toml](lintconfig.toml)
- **Static analysis**: `staticcheck` configured in [staticcheck.conf](staticcheck.conf)
- Run `make checks` before committing; `make build` enforces this automatically.

## Conventions

### Logging
- Log levels: `debug`, `info`, `warn`, `error`
- Log formats: `json` (production), `console` (development)
- Use sugared logger methods: `Info()`, `Infof()`, `Error()`, `Errorf()`, etc.

### Error Handling
- CLI errors are handled centrally in `root.handleError()` which logs and exits with appropriate codes
- Return errors up the call stack; let the root command handle user-facing output

### Adding New Bank Support
1. Create new package under `internal/service/<bankname>/`
2. Implement CSV processing functions following `sparkasse.ProcessDebitStatement()` pattern
3. Add corresponding Cobra subcommand in `cmd/cli/root/`
