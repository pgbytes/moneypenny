package ynab

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/pgbytes/moneypenny/internal/domain"
	"github.com/stretchr/testify/suite"
)

// TransformTestSuite groups all transformer tests.
type TransformTestSuite struct {
	suite.Suite
	tempDir string
}

func TestTransformTestSuite(t *testing.T) {
	suite.Run(t, new(TransformTestSuite))
}

func (s *TransformTestSuite) SetupTest() {
	tempDir, err := os.MkdirTemp("", "ynab_transform_test")
	s.Require().NoError(err)
	s.tempDir = tempDir
}

func (s *TransformTestSuite) TearDownTest() {
	if s.tempDir != "" {
		os.RemoveAll(s.tempDir)
	}
}

// TestTransformToCSV_WithValidTransactions_CreatesCSVFile tests successful transformation.
func (s *TransformTestSuite) TestTransformToCSV_WithValidTransactions_CreatesCSVFile() {
	// Arrange
	outputPath := filepath.Join(s.tempDir, "output.csv")
	transactions := []domain.Transaction{
		{
			Date:   time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC),
			Payee:  "Amazon",
			Memo:   "Office supplies",
			Amount: -25.50,
		},
		{
			Date:   time.Date(2026, 1, 20, 0, 0, 0, 0, time.UTC),
			Payee:  "Salary Deposit",
			Memo:   "",
			Amount: 3500.00,
		},
	}
	ctx := context.Background()

	// Act
	result, err := TransformToCSV(ctx, transactions, outputPath)

	// Assert
	s.NoError(err)
	s.NotNil(result)
	s.Equal(outputPath, result.OutputPath)
	s.Equal(2, result.TransactionCount)

	// Verify file contents
	content, err := os.ReadFile(outputPath)
	s.Require().NoError(err)

	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	s.Equal(3, len(lines), "should have header + 2 data rows")
	s.Equal("Date,Payee,Memo,Amount", lines[0])
	s.Equal("15-01-2026,Amazon,Office supplies,-25.50", lines[1])
	s.Equal("20-01-2026,Salary Deposit,,3500.00", lines[2])
}

// TestTransformToCSV_WithEmptySlice_CreatesHeaderOnlyCSV tests empty transaction handling.
func (s *TransformTestSuite) TestTransformToCSV_WithEmptySlice_CreatesHeaderOnlyCSV() {
	// Arrange
	outputPath := filepath.Join(s.tempDir, "empty_output.csv")
	transactions := []domain.Transaction{}
	ctx := context.Background()

	// Act
	result, err := TransformToCSV(ctx, transactions, outputPath)

	// Assert
	s.NoError(err)
	s.NotNil(result)
	s.Equal(0, result.TransactionCount)

	// Verify file contains only header
	content, err := os.ReadFile(outputPath)
	s.Require().NoError(err)

	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	s.Equal(1, len(lines), "should have only header row")
	s.Equal("Date,Payee,Memo,Amount", lines[0])
}

// TestTransformToCSV_WithCancelledContext_ReturnsError tests context cancellation.
func (s *TransformTestSuite) TestTransformToCSV_WithCancelledContext_ReturnsError() {
	// Arrange
	outputPath := filepath.Join(s.tempDir, "cancelled.csv")
	transactions := []domain.Transaction{
		{
			Date:   time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC),
			Payee:  "Test",
			Amount: -10.00,
		},
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// Act
	result, err := TransformToCSV(ctx, transactions, outputPath)

	// Assert
	s.Error(err)
	s.Nil(result)
	s.Contains(err.Error(), "cancelled")
}

// TestTransformToCSV_WithInvalidPath_ReturnsError tests error handling for invalid path.
func (s *TransformTestSuite) TestTransformToCSV_WithInvalidPath_ReturnsError() {
	// Arrange
	outputPath := "/nonexistent/directory/output.csv"
	transactions := []domain.Transaction{
		{
			Date:   time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC),
			Payee:  "Test",
			Amount: -10.00,
		},
	}
	ctx := context.Background()

	// Act
	result, err := TransformToCSV(ctx, transactions, outputPath)

	// Assert
	s.Error(err)
	s.Nil(result)
	s.Contains(err.Error(), "creating output file")
}

// TestTransformToCSV_DateFormat_UsesCorrectFormat tests DD-MM-YYYY date format.
func (s *TransformTestSuite) TestTransformToCSV_DateFormat_UsesCorrectFormat() {
	// Arrange
	outputPath := filepath.Join(s.tempDir, "date_format.csv")
	transactions := []domain.Transaction{
		{
			Date:   time.Date(2026, 12, 5, 0, 0, 0, 0, time.UTC),
			Payee:  "December Transaction",
			Amount: -100.00,
		},
	}
	ctx := context.Background()

	// Act
	result, err := TransformToCSV(ctx, transactions, outputPath)

	// Assert
	s.NoError(err)
	s.NotNil(result)

	content, err := os.ReadFile(outputPath)
	s.Require().NoError(err)

	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	s.Equal(2, len(lines))
	s.Contains(lines[1], "05-12-2026", "date should be in DD-MM-YYYY format")
}

// TestTransformToCSV_Amount_PreservesSign tests that positive and negative amounts are preserved.
func (s *TransformTestSuite) TestTransformToCSV_Amount_PreservesSign() {
	// Arrange
	outputPath := filepath.Join(s.tempDir, "amounts.csv")
	transactions := []domain.Transaction{
		{
			Date:   time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			Payee:  "Expense",
			Amount: -123.45,
		},
		{
			Date:   time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC),
			Payee:  "Income",
			Amount: 987.65,
		},
		{
			Date:   time.Date(2026, 1, 3, 0, 0, 0, 0, time.UTC),
			Payee:  "Zero",
			Amount: 0.00,
		},
	}
	ctx := context.Background()

	// Act
	result, err := TransformToCSV(ctx, transactions, outputPath)

	// Assert
	s.NoError(err)
	s.NotNil(result)

	content, err := os.ReadFile(outputPath)
	s.Require().NoError(err)

	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	s.Equal(4, len(lines))
	s.Contains(lines[1], "-123.45", "negative amount should be preserved")
	s.Contains(lines[2], "987.65", "positive amount should be preserved")
	s.Contains(lines[3], "0.00", "zero amount should be formatted correctly")
}

// TestTransformToCSV_WithEmptyPayee_LeavesFieldEmpty tests empty payee handling.
func (s *TransformTestSuite) TestTransformToCSV_WithEmptyPayee_LeavesFieldEmpty() {
	// Arrange
	outputPath := filepath.Join(s.tempDir, "empty_payee.csv")
	transactions := []domain.Transaction{
		{
			Date:   time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC),
			Payee:  "",
			Memo:   "No payee transaction",
			Amount: -50.00,
		},
	}
	ctx := context.Background()

	// Act
	result, err := TransformToCSV(ctx, transactions, outputPath)

	// Assert
	s.NoError(err)
	s.NotNil(result)

	content, err := os.ReadFile(outputPath)
	s.Require().NoError(err)

	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	s.Equal(2, len(lines))
	s.Equal("15-01-2026,,No payee transaction,-50.00", lines[1])
}

// TestTransformToCSV_WithTODOContext_Succeeds tests that context.TODO works as expected.
func (s *TransformTestSuite) TestTransformToCSV_WithTODOContext_Succeeds() {
	// Arrange
	outputPath := filepath.Join(s.tempDir, "todo_context.csv")
	transactions := []domain.Transaction{
		{
			Date:   time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC),
			Payee:  "Test",
			Amount: -10.00,
		},
	}

	// Act
	result, err := TransformToCSV(context.TODO(), transactions, outputPath)

	// Assert
	s.NoError(err)
	s.NotNil(result)
	s.Equal(1, result.TransactionCount)
}

// GenerateOutputPathTestSuite groups all output path generation tests.
type GenerateOutputPathTestSuite struct {
	suite.Suite
}

func TestGenerateOutputPathTestSuite(t *testing.T) {
	suite.Run(t, new(GenerateOutputPathTestSuite))
}

// TestGenerateOutputPath_WithCSVExtension_ReplacesExtension tests CSV extension handling.
func (s *GenerateOutputPathTestSuite) TestGenerateOutputPath_WithCSVExtension_ReplacesExtension() {
	// Arrange
	input := "statement.csv"

	// Act
	result := GenerateOutputPath(input)

	// Assert
	s.Equal("statement_ynab.csv", result)
}

// TestGenerateOutputPath_WithNoExtension_AddsSuffix tests files without extension.
func (s *GenerateOutputPathTestSuite) TestGenerateOutputPath_WithNoExtension_AddsSuffix() {
	// Arrange
	input := "statement"

	// Act
	result := GenerateOutputPath(input)

	// Assert
	s.Equal("statement_ynab.csv", result)
}

// TestGenerateOutputPath_WithDifferentExtension_ReplacesExtension tests non-CSV extensions.
func (s *GenerateOutputPathTestSuite) TestGenerateOutputPath_WithDifferentExtension_ReplacesExtension() {
	// Arrange
	input := "/path/to/file.txt"

	// Act
	result := GenerateOutputPath(input)

	// Assert
	s.Equal("/path/to/file_ynab.csv", result)
}

// TestGenerateOutputPath_WithAbsolutePath_PreservesDirectory tests absolute paths.
func (s *GenerateOutputPathTestSuite) TestGenerateOutputPath_WithAbsolutePath_PreservesDirectory() {
	// Arrange
	input := "/home/user/documents/statement.csv"

	// Act
	result := GenerateOutputPath(input)

	// Assert
	s.Equal("/home/user/documents/statement_ynab.csv", result)
}

// TestGenerateOutputPath_WithRelativePath_PreservesDirectory tests relative paths.
func (s *GenerateOutputPathTestSuite) TestGenerateOutputPath_WithRelativePath_PreservesDirectory() {
	// Arrange
	input := "data/2026/statement.csv"

	// Act
	result := GenerateOutputPath(input)

	// Assert
	s.Equal("data/2026/statement_ynab.csv", result)
}

// TestGenerateOutputPath_WithDotsInFilename_HandlesCorrectly tests files with multiple dots.
func (s *GenerateOutputPathTestSuite) TestGenerateOutputPath_WithDotsInFilename_HandlesCorrectly() {
	// Arrange
	input := "statement.2026.01.csv"

	// Act
	result := GenerateOutputPath(input)

	// Assert
	s.Equal("statement.2026.01_ynab.csv", result)
}

// FormatAmountTestSuite groups amount formatting tests.
type FormatAmountTestSuite struct {
	suite.Suite
}

func TestFormatAmountTestSuite(t *testing.T) {
	suite.Run(t, new(FormatAmountTestSuite))
}

// TestFormatAmount_WithTwoDecimalPlaces_FormatsCorrectly tests standard formatting.
func (s *FormatAmountTestSuite) TestFormatAmount_WithTwoDecimalPlaces_FormatsCorrectly() {
	tests := []struct {
		name     string
		amount   float64
		expected string
	}{
		{"negative with decimals", -123.45, "-123.45"},
		{"positive with decimals", 987.65, "987.65"},
		{"zero", 0.0, "0.00"},
		{"negative whole number", -100.0, "-100.00"},
		{"positive whole number", 250.0, "250.00"},
		{"small negative", -0.01, "-0.01"},
		{"large number", 999999.99, "999999.99"},
		{"rounds to two decimals", 10.999, "11.00"},
		{"single decimal input", 5.5, "5.50"},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			result := formatAmount(tc.amount)
			s.Equal(tc.expected, result)
		})
	}
}
