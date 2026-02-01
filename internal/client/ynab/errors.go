package ynab

import (
	"errors"
	"fmt"
)

// Sentinel errors for common error conditions.
var (
	// ErrNotFound indicates the requested resource was not found.
	ErrNotFound = errors.New("resource not found")
	// ErrUnauthorized indicates invalid or missing authentication.
	ErrUnauthorized = errors.New("unauthorized")
	// ErrRateLimited indicates the API rate limit has been exceeded.
	ErrRateLimited = errors.New("rate limit exceeded")
	// ErrBadRequest indicates the request was malformed or invalid.
	ErrBadRequest = errors.New("bad request")
	// ErrConflict indicates a conflict, such as duplicate import_id.
	ErrConflict = errors.New("conflict")
)

// APIError represents an error response from the YNAB API.
type APIError struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Detail string `json:"detail"`
}

// Error implements the error interface.
func (e *APIError) Error() string {
	return fmt.Sprintf("ynab api error [%s]: %s - %s", e.ID, e.Name, e.Detail)
}

// ErrorResponse wraps the API error response structure.
type ErrorResponse struct {
	Error APIError `json:"error"`
}

// mapHTTPStatusToError maps HTTP status codes to sentinel errors.
func mapHTTPStatusToError(statusCode int, apiErr *APIError) error {
	var baseErr error

	switch statusCode {
	case 400:
		baseErr = ErrBadRequest
	case 401:
		baseErr = ErrUnauthorized
	case 403:
		// Could be subscription_lapsed, trial_expired, unauthorized_scope, or data_limit_reached
		baseErr = ErrUnauthorized
	case 404:
		baseErr = ErrNotFound
	case 409:
		baseErr = ErrConflict
	case 429:
		baseErr = ErrRateLimited
	default:
		if apiErr != nil {
			return apiErr
		}
		return fmt.Errorf("unexpected status code: %d", statusCode)
	}

	if apiErr != nil {
		return fmt.Errorf("%w: %s", baseErr, apiErr.Detail)
	}

	return baseErr
}
