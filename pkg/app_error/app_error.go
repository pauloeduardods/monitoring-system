package app_error

import "fmt"

type ApiError struct {
	Message     string `json:"message,omitempty"`
	Description string `json:"description,omitempty"`
	StatusCode  int    `json:"-"`
}

func (e *ApiError) Error() string {
	return fmt.Sprintf("message: %s, description: %s, status_code: %d", e.Message, e.Description, e.StatusCode)
}

func NewApiError(statusCode int, message string, description ...string) *ApiError {
	apiError := &ApiError{
		Message:     message,
		StatusCode:  statusCode,
		Description: "",
	}
	if len(description) > 0 {
		apiError.Description = description[0]
	}
	return apiError
}
