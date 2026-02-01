// Package ynab provides a client for the YNAB (You Need A Budget) API.
//
// The client is designed to be long-running and reusable across multiple
// API calls. It handles authentication, request formatting, and error handling.
//
// Example usage:
//
//	cfg := ynab.Config{
//	    APIKey:   "your-api-key",
//	    BudgetID: "your-budget-id",
//	}
//	client, err := ynab.NewClient(cfg, logger)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	transactions, err := client.GetTransactionsByAccount("account-id", ynab.TransactionOptions{})
package ynab

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/pgbytes/moneypenny/internal/log"
)

const (
	// DefaultBaseURL is the YNAB API base URL.
	DefaultBaseURL = "https://api.ynab.com/v1"

	// DefaultTimeout is the default HTTP request timeout.
	DefaultTimeout = 30 * time.Second

	// DefaultRetryCount is the default number of retries for failed requests.
	DefaultRetryCount = 3

	// DefaultRetryWaitTime is the initial wait time between retries.
	DefaultRetryWaitTime = 1 * time.Second

	// DefaultRetryMaxWaitTime is the maximum wait time between retries.
	DefaultRetryMaxWaitTime = 5 * time.Second
)

// Config holds the configuration for the YNAB client.
type Config struct {
	// APIKey is the personal access token for authentication.
	APIKey string
	// BudgetID is the default budget ID for API operations.
	BudgetID string
	// BaseURL overrides the default API base URL (optional, for testing).
	BaseURL string
	// Timeout overrides the default request timeout (optional).
	Timeout time.Duration
}

// Client is a reusable YNAB API client.
type Client struct {
	httpClient *resty.Client
	baseURL    string
	apiKey     string
	budgetID   string
	logger     log.Logger
}

// NewClient creates a new YNAB API client with the given configuration.
func NewClient(cfg Config, logger log.Logger) (*Client, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("api key is required")
	}

	if cfg.BudgetID == "" {
		return nil, fmt.Errorf("budget id is required")
	}

	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}

	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = DefaultTimeout
	}

	httpClient := resty.New().
		SetBaseURL(baseURL).
		SetTimeout(timeout).
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept", "application/json").
		SetAuthToken(cfg.APIKey).
		SetRetryCount(DefaultRetryCount).
		SetRetryWaitTime(DefaultRetryWaitTime).
		SetRetryMaxWaitTime(DefaultRetryMaxWaitTime).
		AddRetryCondition(func(r *resty.Response, err error) bool {
			// Retry on network errors or 429 (rate limit) or 5xx errors
			if err != nil {
				return true
			}
			return r.StatusCode() == http.StatusTooManyRequests ||
				r.StatusCode() >= http.StatusInternalServerError
		})

	client := &Client{
		httpClient: httpClient,
		baseURL:    baseURL,
		apiKey:     cfg.APIKey,
		budgetID:   cfg.BudgetID,
		logger:     logger,
	}

	logger.Debugf("YNAB client initialized with base URL: %s", baseURL)

	return client, nil
}

// BudgetID returns the configured budget ID.
func (c *Client) BudgetID() string {
	return c.budgetID
}

// GetAccounts retrieves all accounts for the configured budget.
func (c *Client) GetAccounts() ([]Account, error) {
	c.logger.Debugf("Fetching accounts for budget: %s", c.budgetID)

	var result AccountsResponse
	var errResp ErrorResponse

	resp, err := c.httpClient.R().
		SetResult(&result).
		SetError(&errResp).
		Get(fmt.Sprintf("/budgets/%s/accounts", c.budgetID))

	if err != nil {
		return nil, fmt.Errorf("fetching accounts: %w", err)
	}

	if resp.IsError() {
		return nil, mapHTTPStatusToError(resp.StatusCode(), &errResp.Error)
	}

	c.logger.Debugf("Fetched %d accounts", len(result.Data.Accounts))

	return result.Data.Accounts, nil
}

// GetHTTPClient returns the underlying HTTP client for testing purposes.
func (c *Client) GetHTTPClient() *resty.Client {
	return c.httpClient
}
