# YNAB Transformer Package

The `internal/transform/ynab` package provides functionality to transform domain transactions into YNAB-compatible CSV format for import into the [You Need A Budget](https://www.youneedabudget.com/) application.

## Overview

This package is part of the MoneyPenny CLI tool and enables conversion of parsed bank statements (e.g., Miles & More credit card statements) into a standardized CSV format that YNAB can import.

## Package Structure

```
internal/transform/ynab/
├── ynab.go          # Core transformation logic
├── ynab_test.go     # Comprehensive unit tests
└── README.md        # This documentation
```

## Features

- **CSV Generation**: Creates YNAB-compatible CSV files from domain transactions
- **Context Support**: Respects context cancellation for long-running operations
- **Output Path Generation**: Automatically generates output file paths with `_ynab` suffix
- **Strict Error Handling**: Returns detailed errors with context for debugging

## CSV Output Format

The generated CSV follows YNAB import requirements:

| Column | Description | Format |
|--------|-------------|--------|
| Date | Transaction date | `DD-MM-YYYY` (e.g., `15-01-2026`) |
| Payee | Merchant or recipient | String (empty if not available) |
| Memo | Additional notes | String (empty if not available) |
| Amount | Transaction amount | Decimal with 2 places, sign preserved |

### Example Output

```csv
Date,Payee,Memo,Amount
15-01-2026,Amazon,Office supplies,-25.50
20-01-2026,Salary Deposit,,3500.00
02-01-2026,Coffee Shop,,-4.50
```

## API Reference

### TransformToCSV

```go
func TransformToCSV(ctx context.Context, transactions []domain.Transaction, outputPath string) (*TransformResult, error)
```

Transforms a slice of domain transactions into YNAB CSV format and writes to the specified output path.

**Parameters:**
- `ctx`: Context for cancellation support (nil defaults to `context.Background()`)
- `transactions`: Slice of domain transactions to transform
- `outputPath`: Absolute path where the CSV file will be written

**Returns:**
- `*TransformResult`: Contains output path and transaction count on success
- `error`: Wrapped error with context on failure

**Example:**
```go
transactions := []domain.Transaction{
    {
        Date:   time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC),
        Payee:  "Amazon",
        Memo:   "Office supplies",
        Amount: -25.50,
    },
}

result, err := ynab.TransformToCSV(context.Background(), transactions, "/path/to/output.csv")
if err != nil {
    log.Fatalf("Transform failed: %v", err)
}

fmt.Printf("Wrote %d transactions to %s\n", result.TransactionCount, result.OutputPath)
```

### GenerateOutputPath

```go
func GenerateOutputPath(inputPath string) string
```

Generates the output file path based on the input file path by adding `_ynab.csv` suffix.

**Parameters:**
- `inputPath`: Path to the input file

**Returns:**
- Output path with `_ynab.csv` suffix

**Examples:**
| Input | Output |
|-------|--------|
| `statement.csv` | `statement_ynab.csv` |
| `statement` | `statement_ynab.csv` |
| `/path/to/file.txt` | `/path/to/file_ynab.csv` |
| `data/2026/report.csv` | `data/2026/report_ynab.csv` |

### TransformResult

```go
type TransformResult struct {
    OutputPath       string  // Path where CSV was written
    TransactionCount int     // Number of transactions written
}
```

## Design Decisions

### 1. Date Format: DD-MM-YYYY

YNAB supports multiple date formats. We chose `DD-MM-YYYY` as it's commonly used in European locales where Miles & More cards are prevalent.

### 2. Empty Fields Preserved

Empty Payee and Memo fields are preserved as empty strings rather than using placeholders. This allows YNAB users to fill in details during import review.

### 3. Amount Sign Preservation

Transaction amounts preserve their original sign:
- **Negative** values represent outflows (expenses)
- **Positive** values represent inflows (income)

This matches YNAB's expected format and the domain transaction model.

### 4. Two Decimal Places

All amounts are formatted with exactly 2 decimal places for consistency and currency accuracy.

### 5. Strict Mode for Parsing

When used with the CLI command, parsing errors abort the transformation. This prevents partial or incorrect data from being imported into YNAB.

### 6. Context Cancellation

The transformer respects context cancellation, allowing graceful interruption of long-running operations.

## CLI Usage

The transformer is exposed via the CLI command:

```bash
mp ynab transform milesmore -i /path/to/statement.csv
```

This will:
1. Parse the Miles & More CSV statement
2. Validate all rows parsed successfully (strict mode)
3. Transform transactions to YNAB format
4. Write output to `/path/to/statement_ynab.csv`

### Command Structure

```
mp ynab transform milesmore
   │    │         │
   │    │         └── Action: Transform Miles & More statement
   │    └── Subcommand: Transform operations
   └── Command: YNAB operations
```

## Testing

The package includes comprehensive tests using `testify/suite`:

```bash
# Run all transformer tests
go test ./internal/transform/ynab/...

# Run with verbose output
go test -v ./internal/transform/ynab/...
```

### Test Coverage

| Test Suite | Coverage |
|------------|----------|
| `TransformTestSuite` | CSV generation, error handling, edge cases |
| `GenerateOutputPathTestSuite` | Path generation with various inputs |
| `FormatAmountTestSuite` | Amount formatting with decimals |

## Error Handling

All errors are wrapped with context using `fmt.Errorf("context: %w", err)` pattern:

```go
// Example error messages:
// "creating output file: permission denied"
// "transform cancelled: context canceled"
// "writing row 5: csv: error"
```

## Dependencies

- `github.com/pgbytes/moneypenny/internal/domain` - Transaction model
- Standard library: `context`, `encoding/csv`, `os`, `path/filepath`

## Future Enhancements

Potential improvements for future versions:

1. **Configurable date format** - Support different YNAB date format preferences
2. **Memo enrichment** - Option to include foreign currency details in memo
3. **Dry-run mode** - Preview output without writing file
4. **Multiple output formats** - Support YNAB4 vs nYNAB format differences
