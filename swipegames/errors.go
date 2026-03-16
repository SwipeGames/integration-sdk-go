package swipegames

import "fmt"

// APIError represents an error response from the Swipe Games API.
type APIError struct {
	StatusCode int
	Message    string
	Code       string
	Details    string
}

func (e *APIError) Error() string {
	if e.Code != "" {
		return fmt.Sprintf("SwipeGamesApiError: %s (status=%d, code=%s)", e.Message, e.StatusCode, e.Code)
	}
	return fmt.Sprintf("SwipeGamesApiError: %s (status=%d)", e.Message, e.StatusCode)
}

// ValidationError represents a validation error for request parameters.
type ValidationError struct {
	Message string
	Field   string
}

func (e *ValidationError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("SwipeGamesValidationError: %s (field=%s)", e.Message, e.Field)
	}
	return fmt.Sprintf("SwipeGamesValidationError: %s", e.Message)
}
