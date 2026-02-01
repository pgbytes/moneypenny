# MoneyPenny - Copilot Instructions

## Project Overview
MoneyPenny is a personal finance assistant CLI tool written in Go 1.25. It processes financial statements (e.g., Sparkasse CSV exports) and integrates with YNAB (You Need A Budget) for personal finance management.

## Architecture

### Directory Structure
- `cmd/cli/` - CLI entrypoint using Cobra command framework
- `cmd/cli/root/` - Root command and top-level command registration
- `cmd/cli/<command>/` - Command packages (e.g., `ynab/`)
- `cmd/cli/<command>/<subcommand>/` - Subcommand packages (e.g., `ynab/transactions/`)
- `cmd/cli/<command>/<subcommand>/<action>/` - Action packages (e.g., `ynab/transactions/fetch/`)
- `internal/config/` - Application configuration loading (JSON-based)
- `internal/client/` - External API clients (e.g., `ynab/` for YNAB API)
- `internal/service/` - Business logic services (e.g., `sparkasse/` for bank-specific processing)
- `internal/log/` - Centralized logging using Zap with GCP-compatible formatting

### Key Patterns
- **Cobra CLI**: Commands use a nested package structure for isolation:
  - Each command lives in its own package under `cmd/cli/<command>/`
  - Subcommands are nested packages: `cmd/cli/<command>/<subcommand>/`
  - Each package exports a `Cmd` variable that parent packages import and register
  - Flags are defined in the package's `init()` function, isolated from other commands
  - Root command imports and registers top-level commands in `cmd/cli/root/root.go`
- **Logging**: Always use `log.GetLogger()` from `internal/log` - never instantiate loggers directly. Logger must be initialized via `log.SetupLogging()` in main. External packages should accept `log.Logger` interface to avoid zap dependency leakage.
- **Service layer**: Bank-specific logic lives in `internal/service/<bank>/`. Each bank processor should implement statement processing functions.
- **Client layer**: External API clients live in `internal/client/<service>/`. Clients should be long-running and reusable, accepting configuration and logger at initialization.

### Configuration
- Configuration uses JSON format (no external dependencies)
- Config file structure supports multiple services:
```json
{
  "ynab": {
    "api_key": "your-personal-access-token",
    "budget_id": "your-budget-id"
  }
}
```
- Load config via `config.LoadFromFile(path)` from `internal/config`

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
- Use Go 1.13+ error wrapping: `fmt.Errorf("context: %w", err)`

### Testing
- Use `testify/suite` for grouping related tests
- Use table-driven tests for multiple input scenarios
- Use `httptest.Server` for mocking HTTP APIs
- Follow AAA pattern (Arrange-Act-Assert)

### Adding New Bank Support
1. Create new package under `internal/service/<bankname>/`
2. Implement CSV processing functions following `sparkasse.ProcessDebitStatement()` pattern
3. Add corresponding Cobra command package in `cmd/cli/<bankname>/`

### Adding New API Client
1. Create new package under `internal/client/<servicename>/`
2. Define types in `types.go`, errors in `errors.go`, client in `<servicename>.go`
3. Accept `log.Logger` interface for logging (not zap directly)
4. Use `go-resty/resty/v2` for HTTP client
5. Add corresponding Cobra command package in `cmd/cli/<servicename>/`

### Adding New CLI Command
1. Create package at `cmd/cli/<command>/<command>.go`
2. Export a `Cmd` variable of type `*cobra.Command`
3. Register subcommands in `init()` by importing child packages and calling `Cmd.AddCommand(child.Cmd)`
4. Define flags in `init()` - they are isolated to the package namespace
5. Import and register the top-level command in `cmd/cli/root/root.go`

Example structure for `mp foo bar baz`:
```
cmd/cli/foo/
  foo.go           # exports Cmd, imports bar package
  bar/
    bar.go         # exports Cmd, imports baz package  
    baz/
      baz.go       # exports Cmd, defines run function and flags
```

## YNAB Integration

### Client Usage
```go
import (
    "github.com/pgbytes/moneypenny/internal/client/ynab"
    "github.com/pgbytes/moneypenny/internal/config"
    "github.com/pgbytes/moneypenny/internal/log"
)

cfg, _ := config.LoadFromFile("config.json")
client, _ := ynab.NewClient(ynab.Config{
    APIKey:   cfg.YNAB.APIKey,
    BudgetID: cfg.YNAB.BudgetID,
}, log.GetLogger())

transactions, _ := client.GetTransactionsByAccount("account-id", ynab.TransactionOptions{})
```

### CLI Commands
```bash
# Fetch transactions from YNAB
mp ynab transactions fetch -f config.json -a <account-id> -n 20

# With date filter
mp ynab transactions fetch -f config.json -a <account-id> --since-date 2026-01-01
```

### Milliunits
YNAB uses milliunits (1/1000 of currency unit). Use helper functions:
- `ynab.MilliunitsToFloat(123930)` → `123.93`
- `ynab.FloatToMilliunits(123.93)` → `123930`
