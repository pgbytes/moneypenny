package ynab

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/suite"
)

// TransactionsTestSuite groups all transaction-related API tests.
type TransactionsTestSuite struct {
	suite.Suite
	logger *mockLogger
	server *httptest.Server
	client *Client
}

func (s *TransactionsTestSuite) SetupSuite() {
	s.logger = &mockLogger{}
}

func (s *TransactionsTestSuite) TearDownTest() {
	if s.server != nil {
		s.server.Close()
	}
}

func TestTransactionsTestSuite(t *testing.T) {
	suite.Run(t, new(TransactionsTestSuite))
}

func (s *TransactionsTestSuite) setupServerAndClient(handler http.HandlerFunc) {
	s.server = httptest.NewServer(handler)

	cfg := Config{
		APIKey:   "test-api-key",
		BudgetID: "test-budget-id",
		BaseURL:  s.server.URL,
	}

	client, err := NewClient(cfg, s.logger)
	s.Require().NoError(err)
	s.client = client
}

func (s *TransactionsTestSuite) TestGetTransactionsByAccount_WithValidResponse_ReturnsTransactions() {
	// Arrange
	response := TransactionsResponse{
		Data: struct {
			Transactions    []Transaction `json:"transactions"`
			ServerKnowledge int64         `json:"server_knowledge"`
		}{
			Transactions: []Transaction{
				{ID: "tx-1", Date: "2026-01-15", Amount: -50000, PayeeName: "Grocery Store", AccountID: "acc-1"},
				{ID: "tx-2", Date: "2026-01-16", Amount: -25000, PayeeName: "Coffee Shop", AccountID: "acc-1"},
			},
			ServerKnowledge: 150,
		},
	}

	s.setupServerAndClient(func(w http.ResponseWriter, r *http.Request) {
		s.Equal("GET", r.Method)
		s.Contains(r.URL.Path, "/budgets/test-budget-id/accounts/acc-1/transactions")

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	})

	// Act
	transactions, err := s.client.GetTransactionsByAccount("acc-1", TransactionOptions{})

	// Assert
	s.NoError(err)
	s.Len(transactions, 2)
	s.Equal("tx-1", transactions[0].ID)
	s.Equal("Grocery Store", transactions[0].PayeeName)
}

func (s *TransactionsTestSuite) TestGetTransactionsByAccount_WithSinceDate_AppliesFilter() {
	// Arrange
	s.setupServerAndClient(func(w http.ResponseWriter, r *http.Request) {
		s.Equal("2026-01-01", r.URL.Query().Get("since_date"))

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(TransactionsResponse{})
	})

	// Act
	_, err := s.client.GetTransactionsByAccount("acc-1", TransactionOptions{
		SinceDate: "2026-01-01",
	})

	// Assert
	s.NoError(err)
}

func (s *TransactionsTestSuite) TestGetTransactionsByAccount_WithType_AppliesFilter() {
	// Arrange
	s.setupServerAndClient(func(w http.ResponseWriter, r *http.Request) {
		s.Equal("unapproved", r.URL.Query().Get("type"))

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(TransactionsResponse{})
	})

	// Act
	_, err := s.client.GetTransactionsByAccount("acc-1", TransactionOptions{
		Type: "unapproved",
	})

	// Assert
	s.NoError(err)
}

func (s *TransactionsTestSuite) TestGetTransactionsByAccount_WithNotFound_ReturnsError() {
	// Arrange
	s.setupServerAndClient(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(ErrorResponse{
			Error: APIError{ID: "404.2", Name: "resource_not_found", Detail: "Account not found"},
		})
	})

	// Act
	transactions, err := s.client.GetTransactionsByAccount("invalid-acc", TransactionOptions{})

	// Assert
	s.Error(err)
	s.Nil(transactions)
	s.ErrorIs(err, ErrNotFound)
}

func (s *TransactionsTestSuite) TestGetTransactionsByAccount_WithRateLimitExceeded_ReturnsError() {
	// Arrange
	callCount := 0
	s.setupServerAndClient(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusTooManyRequests)
		_ = json.NewEncoder(w).Encode(ErrorResponse{
			Error: APIError{ID: "429", Name: "too_many_requests", Detail: "Rate limit exceeded"},
		})
	})

	// Act
	transactions, err := s.client.GetTransactionsByAccount("acc-1", TransactionOptions{})

	// Assert
	s.Error(err)
	s.Nil(transactions)
	s.ErrorIs(err, ErrRateLimited)
	// Verify retries happened (default is 3 retries + 1 original = 4 calls)
	s.GreaterOrEqual(callCount, 1)
}

func (s *TransactionsTestSuite) TestCreateTransactions_WithValidData_CreatesTransactions() {
	// Arrange
	response := SaveTransactionsResponse{
		Data: struct {
			TransactionIDs     []string      `json:"transaction_ids"`
			Transaction        *Transaction  `json:"transaction,omitempty"`
			Transactions       []Transaction `json:"transactions,omitempty"`
			DuplicateImportIDs []string      `json:"duplicate_import_ids,omitempty"`
			ServerKnowledge    int64         `json:"server_knowledge"`
		}{
			TransactionIDs:  []string{"tx-new-1", "tx-new-2"},
			ServerKnowledge: 200,
		},
	}

	s.setupServerAndClient(func(w http.ResponseWriter, r *http.Request) {
		s.Equal("POST", r.Method)
		s.Contains(r.URL.Path, "/budgets/test-budget-id/transactions")

		var reqBody SaveTransactionsRequest
		_ = json.NewDecoder(r.Body).Decode(&reqBody)
		s.Len(reqBody.Transactions, 2)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(response)
	})

	// Act
	result, err := s.client.CreateTransactions([]SaveTransaction{
		{AccountID: "acc-1", Date: "2026-01-20", Amount: -10000, PayeeName: "Test Payee 1"},
		{AccountID: "acc-1", Date: "2026-01-21", Amount: -20000, PayeeName: "Test Payee 2"},
	})

	// Assert
	s.NoError(err)
	s.NotNil(result)
	s.Len(result.Data.TransactionIDs, 2)
}

func (s *TransactionsTestSuite) TestCreateTransactions_WithDuplicateImportID_Returns409Error() {
	// Arrange
	s.setupServerAndClient(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		_ = json.NewEncoder(w).Encode(ErrorResponse{
			Error: APIError{ID: "409", Name: "conflict", Detail: "A transaction with this import_id already exists"},
		})
	})

	// Act
	result, err := s.client.CreateTransactions([]SaveTransaction{
		{AccountID: "acc-1", Date: "2026-01-20", Amount: -10000, ImportID: "YNAB:-10000:2026-01-20:1"},
	})

	// Assert
	s.Error(err)
	s.Nil(result)
	s.ErrorIs(err, ErrConflict)
}

func (s *TransactionsTestSuite) TestCreateTransactions_WithEmptyList_ReturnsError() {
	// Arrange
	s.setupServerAndClient(func(w http.ResponseWriter, r *http.Request) {
		// Should not be called
		s.Fail("Server should not be called with empty transaction list")
	})

	// Act
	result, err := s.client.CreateTransactions([]SaveTransaction{})

	// Assert
	s.Error(err)
	s.Nil(result)
	s.Contains(err.Error(), "at least one transaction is required")
}

func (s *TransactionsTestSuite) TestCreateTransaction_WithSingleTransaction_CreatesTransaction() {
	// Arrange
	response := SaveTransactionsResponse{
		Data: struct {
			TransactionIDs     []string      `json:"transaction_ids"`
			Transaction        *Transaction  `json:"transaction,omitempty"`
			Transactions       []Transaction `json:"transactions,omitempty"`
			DuplicateImportIDs []string      `json:"duplicate_import_ids,omitempty"`
			ServerKnowledge    int64         `json:"server_knowledge"`
		}{
			TransactionIDs: []string{"tx-single"},
			Transaction: &Transaction{
				ID:        "tx-single",
				Date:      "2026-01-20",
				Amount:    -10000,
				PayeeName: "Single Payee",
			},
			ServerKnowledge: 201,
		},
	}

	s.setupServerAndClient(func(w http.ResponseWriter, r *http.Request) {
		var reqBody SaveTransactionsRequest
		_ = json.NewDecoder(r.Body).Decode(&reqBody)
		s.NotNil(reqBody.Transaction)
		s.Nil(reqBody.Transactions)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(response)
	})

	// Act
	result, err := s.client.CreateTransaction(SaveTransaction{
		AccountID: "acc-1",
		Date:      "2026-01-20",
		Amount:    -10000,
		PayeeName: "Single Payee",
	})

	// Assert
	s.NoError(err)
	s.NotNil(result)
	s.Len(result.Data.TransactionIDs, 1)
}

func (s *TransactionsTestSuite) TestGetTransactions_WithValidResponse_ReturnsAllTransactions() {
	// Arrange
	response := TransactionsResponse{
		Data: struct {
			Transactions    []Transaction `json:"transactions"`
			ServerKnowledge int64         `json:"server_knowledge"`
		}{
			Transactions: []Transaction{
				{ID: "tx-1", Date: "2026-01-15", Amount: -50000},
				{ID: "tx-2", Date: "2026-01-16", Amount: -25000},
				{ID: "tx-3", Date: "2026-01-17", Amount: -75000},
			},
			ServerKnowledge: 300,
		},
	}

	s.setupServerAndClient(func(w http.ResponseWriter, r *http.Request) {
		s.Contains(r.URL.Path, "/budgets/test-budget-id/transactions")
		s.NotContains(r.URL.Path, "/accounts/")

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	})

	// Act
	transactions, err := s.client.GetTransactions(TransactionOptions{})

	// Assert
	s.NoError(err)
	s.Len(transactions, 3)
}
