// Package domain provides common domain models used across the application.
package domain

import "time"

// Transaction represents a financial transaction that can be used across
// different banking services and import sources.
type Transaction struct {
	// Date is the primary transaction date (voucher date/booking date).
	Date time.Time

	// PostingDate is when the transaction was posted/received.
	PostingDate time.Time

	// Payee is the merchant, recipient, or reason for payment.
	Payee string

	// Memo contains additional transaction description or notes.
	Memo string

	// Amount is the transaction amount in the settlement currency.
	// Negative values indicate outflows (expenses).
	Amount float64

	// Currency is the settlement currency code (e.g., "EUR", "USD").
	Currency string

	// ForeignAmount is the original amount in foreign currency (if applicable).
	// Zero value indicates no foreign currency conversion.
	ForeignAmount float64

	// ForeignCurrency is the foreign currency code (e.g., "USD", "GBP").
	// Empty string indicates no foreign currency conversion.
	ForeignCurrency string

	// ExchangeRate is the rate used for currency conversion.
	// Zero value indicates no conversion or rate not provided.
	ExchangeRate float64

	// ImportID is a unique identifier for duplicate detection.
	// Format: "YNAB:[milliunit_amount]:[iso_date]:[occurrence]"
	// Example: "YNAB:-294230:2015-12-30:1"
	ImportID string

	// SourceFile is the name of the file this transaction was parsed from.
	SourceFile string

	// SourceLine is the line number in the source file (for debugging).
	SourceLine int
}
