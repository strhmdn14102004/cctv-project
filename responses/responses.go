package responses

import (
	"encoding/json"
	"net/http"

	"github.com/go-playground/validator/v10"
)

// ErrorResponse represents a standard error response structure
type ErrorResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// ValidationError represents a single field validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidationErrorResponse represents a collection of validation errors
type ValidationErrorResponse struct {
	Success bool              `json:"success"`
	Errors  []ValidationError `json:"errors"`
}

// SendErrorResponse sends a standard error response
func SendErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(ErrorResponse{
		Success: false,
		Message: message,
	})
}

// SendSuccessResponse sends a standard success response with data
func SendSuccessResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    data,
	})
}

// SendValidationError sends a validation error response
func SendValidationError(w http.ResponseWriter, err error) {
	var errors []ValidationError
	for _, err := range err.(validator.ValidationErrors) {
		errors = append(errors, ValidationError{
			Field:   err.Field(),
			Message: err.Tag(),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(ValidationErrorResponse{
		Success: false,
		Errors:  errors,
	})
}
