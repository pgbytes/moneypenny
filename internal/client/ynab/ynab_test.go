package ynab

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/suite"
)

// mockLogger implements log.Logger interface for testing.
type mockLogger struct{}

func (m *mockLogger) Warn(args ...interface{})                    {}
func (m *mockLogger) Warnf(template string, args ...interface{})  {}
func (m *mockLogger) Info(args ...interface{})                    {}
func (m *mockLogger) Infof(template string, args ...interface{})  {}
func (m *mockLogger) Debug(args ...interface{})                   {}
func (m *mockLogger) Debugf(template string, args ...interface{}) {}
func (m *mockLogger) Error(args ...interface{})                   {}
func (m *mockLogger) Errorf(template string, args ...interface{}) {}
func (m *mockLogger) Fatal(args ...interface{})                   {}
func (m *mockLogger) Fatalf(template string, args ...interface{}) {}

// YNABClientTestSuite groups all YNAB client initialization tests.
type YNABClientTestSuite struct {
	suite.Suite
	logger *mockLogger
}

func (s *YNABClientTestSuite) SetupSuite() {
	s.logger = &mockLogger{}
}

func TestYNABClientTestSuite(t *testing.T) {
	suite.Run(t, new(YNABClientTestSuite))
}

func (s *YNABClientTestSuite) TestNewClient_WithValidConfig_CreatesClient() {
	// Arrange
	cfg := Config{
		APIKey:   "test-api-key",
		BudgetID: "test-budget-id",
	}

	// Act
	client, err := NewClient(cfg, s.logger)

	// Assert
	s.NoError(err)
	s.NotNil(client)
	s.Equal("test-budget-id", client.BudgetID())
}

func (s *YNABClientTestSuite) TestNewClient_WithEmptyAPIKey_ReturnsError() {
	// Arrange
	cfg := Config{
		APIKey:   "",
		BudgetID: "test-budget-id",
	}

	// Act
	client, err := NewClient(cfg, s.logger)

	// Assert
	s.Error(err)
	s.Nil(client)
	s.Contains(err.Error(), "api key is required")
}

func (s *YNABClientTestSuite) TestNewClient_WithEmptyBudgetID_ReturnsError() {
	// Arrange
	cfg := Config{
		APIKey:   "test-api-key",
		BudgetID: "",
	}

	// Act
	client, err := NewClient(cfg, s.logger)

	// Assert
	s.Error(err)
	s.Nil(client)
	s.Contains(err.Error(), "budget id is required")
}

func (s *YNABClientTestSuite) TestNewClient_WithCustomBaseURL_UsesCustomURL() {
	// Arrange
	cfg := Config{
		APIKey:   "test-api-key",
		BudgetID: "test-budget-id",
		BaseURL:  "https://custom.api.com/v1",
	}

	// Act
	client, err := NewClient(cfg, s.logger)

	// Assert
	s.NoError(err)
	s.NotNil(client)
	s.Equal("https://custom.api.com/v1", client.baseURL)
}

// AccountsTestSuite groups account-related API tests.
type AccountsTestSuite struct {
	suite.Suite
	logger *mockLogger
	server *httptest.Server
	client *Client
}

func (s *AccountsTestSuite) SetupSuite() {
	s.logger = &mockLogger{}
}

func (s *AccountsTestSuite) TearDownTest() {
	if s.server != nil {
		s.server.Close()
	}
}

func TestAccountsTestSuite(t *testing.T) {
	suite.Run(t, new(AccountsTestSuite))
}

func (s *AccountsTestSuite) setupServerAndClient(handler http.HandlerFunc) {
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

func (s *AccountsTestSuite) TestGetAccounts_WithValidResponse_ReturnsAccounts() {
	// Arrange
	response := AccountsResponse{
		Data: struct {
			Accounts        []Account `json:"accounts"`
			ServerKnowledge int64     `json:"server_knowledge"`
		}{
			Accounts: []Account{
				{ID: "acc-1", Name: "Checking", Type: "checking", Balance: 100000},
				{ID: "acc-2", Name: "Savings", Type: "savings", Balance: 500000},
			},
			ServerKnowledge: 100,
		},
	}

	s.setupServerAndClient(func(w http.ResponseWriter, r *http.Request) {
		s.Equal("GET", r.Method)
		s.Contains(r.URL.Path, "/budgets/test-budget-id/accounts")
		s.Equal("Bearer test-api-key", r.Header.Get("Authorization"))

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	})

	// Act
	accounts, err := s.client.GetAccounts()

	// Assert
	s.NoError(err)
	s.Len(accounts, 2)
	s.Equal("acc-1", accounts[0].ID)
	s.Equal("Checking", accounts[0].Name)
}

func (s *AccountsTestSuite) TestGetAccounts_WithUnauthorized_ReturnsError() {
	// Arrange
	s.setupServerAndClient(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(ErrorResponse{
			Error: APIError{ID: "401", Name: "not_authorized", Detail: "Invalid token"},
		})
	})

	// Act
	accounts, err := s.client.GetAccounts()

	// Assert
	s.Error(err)
	s.Nil(accounts)
	s.ErrorIs(err, ErrUnauthorized)
}

func (s *AccountsTestSuite) TestGetAccounts_WithNotFound_ReturnsError() {
	// Arrange
	s.setupServerAndClient(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(ErrorResponse{
			Error: APIError{ID: "404.2", Name: "resource_not_found", Detail: "Budget not found"},
		})
	})

	// Act
	accounts, err := s.client.GetAccounts()

	// Assert
	s.Error(err)
	s.Nil(accounts)
	s.ErrorIs(err, ErrNotFound)
}
