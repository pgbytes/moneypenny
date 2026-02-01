// Package ynab provides a client for the YNAB (You Need A Budget) API.
package ynab

import "time"

// Account represents a YNAB account.
type Account struct {
	ID                 string `json:"id"`
	Name               string `json:"name"`
	Type               string `json:"type"`
	OnBudget           bool   `json:"on_budget"`
	Closed             bool   `json:"closed"`
	Note               string `json:"note"`
	Balance            int64  `json:"balance"`
	ClearedBalance     int64  `json:"cleared_balance"`
	UnclearedBalance   int64  `json:"uncleared_balance"`
	TransferPayeeID    string `json:"transfer_payee_id"`
	DirectImportLinked bool   `json:"direct_import_linked"`
	DirectImportError  bool   `json:"direct_import_in_error"`
	Deleted            bool   `json:"deleted"`
}

// Transaction represents a YNAB transaction.
type Transaction struct {
	ID                    string           `json:"id"`
	Date                  string           `json:"date"`
	Amount                int64            `json:"amount"`
	Memo                  string           `json:"memo"`
	Cleared               ClearedStatus    `json:"cleared"`
	Approved              bool             `json:"approved"`
	FlagColor             string           `json:"flag_color"`
	FlagName              string           `json:"flag_name"`
	AccountID             string           `json:"account_id"`
	AccountName           string           `json:"account_name"`
	PayeeID               string           `json:"payee_id"`
	PayeeName             string           `json:"payee_name"`
	CategoryID            string           `json:"category_id"`
	CategoryName          string           `json:"category_name"`
	TransferAccountID     string           `json:"transfer_account_id"`
	TransferTransactionID string           `json:"transfer_transaction_id"`
	MatchedTransactionID  string           `json:"matched_transaction_id"`
	ImportID              string           `json:"import_id"`
	ImportPayeeName       string           `json:"import_payee_name"`
	ImportPayeeOriginal   string           `json:"import_payee_name_original"`
	DebtTransactionType   string           `json:"debt_transaction_type"`
	Deleted               bool             `json:"deleted"`
	Subtransactions       []SubTransaction `json:"subtransactions"`
}

// SubTransaction represents a split transaction component.
type SubTransaction struct {
	ID                    string `json:"id"`
	TransactionID         string `json:"transaction_id"`
	Amount                int64  `json:"amount"`
	Memo                  string `json:"memo"`
	PayeeID               string `json:"payee_id"`
	PayeeName             string `json:"payee_name"`
	CategoryID            string `json:"category_id"`
	CategoryName          string `json:"category_name"`
	TransferAccountID     string `json:"transfer_account_id"`
	TransferTransactionID string `json:"transfer_transaction_id"`
	Deleted               bool   `json:"deleted"`
}

// SaveTransaction represents a transaction to be created or updated.
type SaveTransaction struct {
	AccountID       string               `json:"account_id"`
	Date            string               `json:"date"`
	Amount          int64                `json:"amount"`
	PayeeID         string               `json:"payee_id,omitempty"`
	PayeeName       string               `json:"payee_name,omitempty"`
	CategoryID      string               `json:"category_id,omitempty"`
	Memo            string               `json:"memo,omitempty"`
	Cleared         ClearedStatus        `json:"cleared,omitempty"`
	Approved        bool                 `json:"approved,omitempty"`
	FlagColor       string               `json:"flag_color,omitempty"`
	ImportID        string               `json:"import_id,omitempty"`
	Subtransactions []SaveSubTransaction `json:"subtransactions,omitempty"`
}

// SaveSubTransaction represents a subtransaction to be created.
type SaveSubTransaction struct {
	Amount     int64  `json:"amount"`
	PayeeID    string `json:"payee_id,omitempty"`
	PayeeName  string `json:"payee_name,omitempty"`
	CategoryID string `json:"category_id,omitempty"`
	Memo       string `json:"memo,omitempty"`
}

// ClearedStatus represents the cleared state of a transaction.
type ClearedStatus string

const (
	// ClearedStatusCleared indicates the transaction is cleared.
	ClearedStatusCleared ClearedStatus = "cleared"
	// ClearedStatusUncleared indicates the transaction is not cleared.
	ClearedStatusUncleared ClearedStatus = "uncleared"
	// ClearedStatusReconciled indicates the transaction is reconciled.
	ClearedStatusReconciled ClearedStatus = "reconciled"
)

// TransactionOptions contains optional parameters for fetching transactions.
type TransactionOptions struct {
	// SinceDate filters transactions to those on or after this date (ISO format: YYYY-MM-DD).
	SinceDate string
	// Type filters transactions by type: "uncategorized" or "unapproved".
	Type string
	// LastKnowledgeOfServer is used for delta requests.
	LastKnowledgeOfServer int64
}

// API Response wrapper types

// AccountsResponse wraps the accounts list response.
type AccountsResponse struct {
	Data struct {
		Accounts        []Account `json:"accounts"`
		ServerKnowledge int64     `json:"server_knowledge"`
	} `json:"data"`
}

// TransactionsResponse wraps the transactions list response.
type TransactionsResponse struct {
	Data struct {
		Transactions    []Transaction `json:"transactions"`
		ServerKnowledge int64         `json:"server_knowledge"`
	} `json:"data"`
}

// SaveTransactionsRequest is the request body for creating transactions.
type SaveTransactionsRequest struct {
	Transaction  *SaveTransaction  `json:"transaction,omitempty"`
	Transactions []SaveTransaction `json:"transactions,omitempty"`
}

// SaveTransactionsResponse wraps the create transactions response.
type SaveTransactionsResponse struct {
	Data struct {
		TransactionIDs     []string      `json:"transaction_ids"`
		Transaction        *Transaction  `json:"transaction,omitempty"`
		Transactions       []Transaction `json:"transactions,omitempty"`
		DuplicateImportIDs []string      `json:"duplicate_import_ids,omitempty"`
		ServerKnowledge    int64         `json:"server_knowledge"`
	} `json:"data"`
}

// BudgetSummary represents a YNAB budget summary.
type BudgetSummary struct {
	ID             string          `json:"id"`
	Name           string          `json:"name"`
	LastModifiedOn string          `json:"last_modified_on"`
	FirstMonth     string          `json:"first_month"`
	LastMonth      string          `json:"last_month"`
	DateFormat     *DateFormat     `json:"date_format"`
	CurrencyFormat *CurrencyFormat `json:"currency_format"`
	Accounts       []Account       `json:"accounts,omitempty"`
}

// DateFormat represents the date format settings for a budget.
type DateFormat struct {
	Format string `json:"format"`
}

// CurrencyFormat represents the currency format settings for a budget.
type CurrencyFormat struct {
	ISOCode          string `json:"iso_code"`
	ExampleFormat    string `json:"example_format"`
	DecimalDigits    int    `json:"decimal_digits"`
	DecimalSeparator string `json:"decimal_separator"`
	SymbolFirst      bool   `json:"symbol_first"`
	GroupSeparator   string `json:"group_separator"`
	CurrencySymbol   string `json:"currency_symbol"`
	DisplaySymbol    bool   `json:"display_symbol"`
}

// BudgetSummaryResponse wraps the budgets list response.
type BudgetSummaryResponse struct {
	Data struct {
		Budgets       []BudgetSummary `json:"budgets"`
		DefaultBudget *BudgetSummary  `json:"default_budget,omitempty"`
	} `json:"data"`
}

// Milliunits conversion helpers

// MilliunitsToFloat converts YNAB milliunits to a float64 amount.
// YNAB stores amounts as milliunits (1/1000 of currency unit).
// Example: 123930 milliunits = $123.93
func MilliunitsToFloat(milliunits int64) float64 {
	return float64(milliunits) / 1000.0
}

// FloatToMilliunits converts a float64 amount to YNAB milliunits.
// Example: $123.93 = 123930 milliunits
func FloatToMilliunits(amount float64) int64 {
	return int64(amount * 1000)
}

// GenerateImportID creates a YNAB-compatible import ID for deduplication.
// Format: YNAB:[milliunit_amount]:[iso_date]:[occurrence]
// Example: YNAB:-294230:2015-12-30:1
func GenerateImportID(amount int64, date time.Time, occurrence int) string {
	return "YNAB:" + formatInt64(amount) + ":" + date.Format("2006-01-02") + ":" + formatInt64(int64(occurrence))
}

// formatInt64 converts an int64 to string without importing strconv.
func formatInt64(n int64) string {
	if n == 0 {
		return "0"
	}

	negative := n < 0
	if negative {
		n = -n
	}

	var digits []byte
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}

	if negative {
		digits = append([]byte{'-'}, digits...)
	}

	return string(digits)
}
