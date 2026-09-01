package client

import (
	"encoding/json"
	"errors"
	"fmt"
)

type APIError struct {
	StatusCode int
	Message    string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("API error (HTTP %d): %s", e.StatusCode, e.Message)
}

func IsNotFound(err error) bool {
	var apiErr *APIError
	return errors.As(err, &apiErr) && apiErr.StatusCode == 404
}

func IsConflict(err error) bool {
	var apiErr *APIError
	return errors.As(err, &apiErr) && apiErr.StatusCode == 409
}

func parseAPIError(statusCode int, body []byte) *APIError {
	apiErr := &APIError{StatusCode: statusCode}

	var objError struct {
		Error   string `json:"error"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(body, &objError); err == nil {
		if objError.Error != "" {
			apiErr.Message = objError.Error
			return apiErr
		}
		if objError.Message != "" {
			apiErr.Message = objError.Message
			return apiErr
		}
	}

	var strError string
	if err := json.Unmarshal(body, &strError); err == nil && strError != "" {
		apiErr.Message = strError
		return apiErr
	}

	if len(body) > 0 {
		apiErr.Message = string(body)
	} else {
		apiErr.Message = fmt.Sprintf("HTTP %d", statusCode)
	}
	return apiErr
}
