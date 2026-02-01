package ynab

import (
	"fmt"
)

// GetTransactions retrieves all transactions for the configured budget.
func (c *Client) GetTransactions(opts TransactionOptions) ([]Transaction, error) {
	c.logger.Debugf("Fetching transactions for budget: %s", c.budgetID)

	var result TransactionsResponse
	var errResp ErrorResponse

	req := c.httpClient.R().
		SetResult(&result).
		SetError(&errResp)

	// Apply optional query parameters
	if opts.SinceDate != "" {
		req.SetQueryParam("since_date", opts.SinceDate)
	}
	if opts.Type != "" {
		req.SetQueryParam("type", opts.Type)
	}
	if opts.LastKnowledgeOfServer > 0 {
		req.SetQueryParam("last_knowledge_of_server", fmt.Sprintf("%d", opts.LastKnowledgeOfServer))
	}

	resp, err := req.Get(fmt.Sprintf("/budgets/%s/transactions", c.budgetID))
	if err != nil {
		return nil, fmt.Errorf("fetching transactions: %w", err)
	}

	if resp.IsError() {
		return nil, mapHTTPStatusToError(resp.StatusCode(), &errResp.Error)
	}

	c.logger.Debugf("Fetched %d transactions", len(result.Data.Transactions))

	return result.Data.Transactions, nil
}

// GetTransactionsByAccount retrieves transactions for a specific account.
func (c *Client) GetTransactionsByAccount(accountID string, opts TransactionOptions) ([]Transaction, error) {
	c.logger.Debugf("Fetching transactions for account: %s in budget: %s", accountID, c.budgetID)

	var result TransactionsResponse
	var errResp ErrorResponse

	req := c.httpClient.R().
		SetResult(&result).
		SetError(&errResp)

	// Apply optional query parameters
	if opts.SinceDate != "" {
		req.SetQueryParam("since_date", opts.SinceDate)
	}
	if opts.Type != "" {
		req.SetQueryParam("type", opts.Type)
	}
	if opts.LastKnowledgeOfServer > 0 {
		req.SetQueryParam("last_knowledge_of_server", fmt.Sprintf("%d", opts.LastKnowledgeOfServer))
	}

	resp, err := req.Get(fmt.Sprintf("/budgets/%s/accounts/%s/transactions", c.budgetID, accountID))
	if err != nil {
		return nil, fmt.Errorf("fetching account transactions: %w", err)
	}

	if resp.IsError() {
		return nil, mapHTTPStatusToError(resp.StatusCode(), &errResp.Error)
	}

	c.logger.Debugf("Fetched %d transactions for account %s", len(result.Data.Transactions), accountID)

	return result.Data.Transactions, nil
}

// CreateTransaction creates a single transaction.
func (c *Client) CreateTransaction(transaction SaveTransaction) (*SaveTransactionsResponse, error) {
	return c.createTransactionsInternal(&SaveTransactionsRequest{
		Transaction: &transaction,
	})
}

// CreateTransactions creates multiple transactions in a single request.
func (c *Client) CreateTransactions(transactions []SaveTransaction) (*SaveTransactionsResponse, error) {
	if len(transactions) == 0 {
		return nil, fmt.Errorf("at least one transaction is required")
	}

	return c.createTransactionsInternal(&SaveTransactionsRequest{
		Transactions: transactions,
	})
}

// createTransactionsInternal handles the API call for creating transactions.
func (c *Client) createTransactionsInternal(reqBody *SaveTransactionsRequest) (*SaveTransactionsResponse, error) {
	c.logger.Debugf("Creating transactions in budget: %s", c.budgetID)

	var result SaveTransactionsResponse
	var errResp ErrorResponse

	resp, err := c.httpClient.R().
		SetBody(reqBody).
		SetResult(&result).
		SetError(&errResp).
		Post(fmt.Sprintf("/budgets/%s/transactions", c.budgetID))

	if err != nil {
		return nil, fmt.Errorf("creating transactions: %w", err)
	}

	if resp.IsError() {
		return nil, mapHTTPStatusToError(resp.StatusCode(), &errResp.Error)
	}

	c.logger.Debugf("Created %d transactions", len(result.Data.TransactionIDs))

	if len(result.Data.DuplicateImportIDs) > 0 {
		c.logger.Debugf("Skipped %d duplicate transactions", len(result.Data.DuplicateImportIDs))
	}

	return &result, nil
}

// LimitTransactions returns the last n transactions from a slice.
// If n is greater than the slice length, returns all transactions.
func LimitTransactions(transactions []Transaction, n int) []Transaction {
	if n <= 0 || n >= len(transactions) {
		return transactions
	}

	// Return the last n transactions (most recent)
	start := len(transactions) - n
	return transactions[start:]
}
