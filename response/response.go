package response

import (
	"net/http"

	"github.com/Masharah-Advisory/common/i18n"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// ErrorItem represents a structured error item
type ErrorItem struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// ApiResponse represents the generic API response structure
type ApiResponse[T any] struct {
	Success bool        `json:"success"`
	Data    *T          `json:"data,omitempty"`
	Errors  []ErrorItem `json:"errors,omitempty"`
	Message string      `json:"message"`
}

// Helper function to create pointer from string
func Ptr(s string) *string {
	return &s
}

// Helper function to create ErrorItem slice from single error
func Err(key, value string) []ErrorItem {
	return []ErrorItem{{Key: key, Value: value}}
}

// Helper function to create ErrorItems from map
func Errs(errors map[string]string) []ErrorItem {
	var items []ErrorItem
	for key, value := range errors {
		items = append(items, ErrorItem{Key: key, Value: value})
	}
	return items
}

// ValidationErrors converts validator.ValidationErrors to localized error items
func ValidationErrors(c *gin.Context, errs validator.ValidationErrors) []ErrorItem {
	var items []ErrorItem

	for _, e := range errs {
		// Construct i18n key based on validator tag, e.g., "validation.required"
		key := "validation." + e.Tag()

		// Template data can include field name and param
		data := gin.H{
			"Field": e.Field(), // Struct field name
			"Param": e.Param(), // Tag param, e.g., max=10 -> Param="10"
		}

		// Translate using i18n T() function
		localizedMessage := i18n.T(c, key, data)

		items = append(items, ErrorItem{
			Key:   e.Field(),
			Value: localizedMessage,
		})
	}

	return items
}

// Simple success response functions (most common use cases)

// OK sends a 200 OK response
func OK[T any](c *gin.Context, data T, message ...string) {
	msg := "Success"
	if len(message) > 0 {
		msg = message[0]
	}
	c.JSON(http.StatusOK, ApiResponse[T]{
		Success: true,
		Data:    &data,
		Message: msg,
	})
}

// OKMessage sends a 200 OK response with just a message
func OKMessage(c *gin.Context, message ...string) {
	msg := "Success"
	if len(message) > 0 {
		msg = message[0]
	}
	c.JSON(http.StatusOK, ApiResponse[any]{
		Success: true,
		Message: msg,
	})
}

func Accepted[T any](c *gin.Context, data T, message ...string) {
	msg := "Request accepted successfully"
	if len(message) > 0 {
		msg = message[0]
	}
	c.JSON(http.StatusAccepted, ApiResponse[T]{
		Success: true,
		Data:    &data,
		Message: msg,
	})
}

// Created sends a 201 Created response
func Created[T any](c *gin.Context, data T, message ...string) {
	msg := "Resource created successfully"
	if len(message) > 0 {
		msg = message[0]
	}
	c.JSON(http.StatusCreated, ApiResponse[T]{
		Success: true,
		Data:    &data,
		Message: msg,
	})
}

// NoContent sends a 204 No Content response
func NoContent(c *gin.Context, message ...string) {
	msg := "Success"
	if len(message) > 0 {
		msg = message[0]
	}
	c.JSON(http.StatusNoContent, ApiResponse[any]{
		Success: true,
		Message: msg,
	})
}

// Simple error response functions (most common use cases)

// BadRequest sends a 400 Bad Request response
func BadRequest(c *gin.Context, message string, errors ...[]ErrorItem) {
	response := ApiResponse[any]{
		Success: false,
		Message: message,
	}
	if len(errors) > 0 {
		response.Errors = errors[0]
	}
	c.JSON(http.StatusBadRequest, response)
}

// Unauthorized sends a 401 Unauthorized response
func Unauthorized(c *gin.Context, message ...string) {
	msg := "Unauthorized"
	if len(message) > 0 {
		msg = message[0]
	}
	c.JSON(http.StatusUnauthorized, ApiResponse[any]{
		Success: false,
		Message: msg,
	})
}

// Forbidden sends a 403 Forbidden response
func Forbidden(c *gin.Context, message ...string) {
	msg := "Forbidden"
	if len(message) > 0 {
		msg = message[0]
	}
	c.JSON(http.StatusForbidden, ApiResponse[any]{
		Success: false,
		Message: msg,
	})
}

// NotFound sends a 404 Not Found response
func NotFound(c *gin.Context, message ...string) {
	msg := "Not found"
	if len(message) > 0 {
		msg = message[0]
	}
	c.JSON(http.StatusNotFound, ApiResponse[any]{
		Success: false,
		Message: msg,
	})
}

// Conflict sends a 409 Conflict response
func Conflict(c *gin.Context, message string, errors ...[]ErrorItem) {
	response := ApiResponse[any]{
		Success: false,
		Message: message,
	}
	if len(errors) > 0 {
		response.Errors = errors[0]
	}
	c.JSON(http.StatusConflict, response)
}

// ValidationFailed sends a 422 Unprocessable Entity response
func ValidationFailed(c *gin.Context, message string, errors ...[]ErrorItem) {
	response := ApiResponse[any]{
		Success: false,
		Message: message,
	}
	if len(errors) > 0 {
		response.Errors = errors[0]
	}
	c.JSON(http.StatusUnprocessableEntity, response)
}

// InternalError sends a 500 Internal Server Error response
func InternalError(c *gin.Context, message ...string) {
	msg := "Internal server error"
	if len(message) > 0 {
		msg = message[0]
	}
	c.JSON(http.StatusInternalServerError, ApiResponse[any]{
		Success: false,
		Message: msg,
	})
}

// Advanced functions for custom use cases

// Success sends a custom success response
func Success[T any](c *gin.Context, statusCode int, data T, message string) {
	c.JSON(statusCode, ApiResponse[T]{
		Success: true,
		Data:    &data,
		Message: message,
	})
}

// Error sends a custom error response
func Error(c *gin.Context, statusCode int, message string, errors ...[]ErrorItem) {
	response := ApiResponse[any]{
		Success: false,
		Message: message,
	}
	if len(errors) > 0 {
		response.Errors = errors[0]
	}
	c.JSON(statusCode, response)
}

// JSON sends a custom JSON response (for complete control)
func JSON[T any](c *gin.Context, statusCode int, success bool, data *T, message string, errors []ErrorItem) {
	c.JSON(statusCode, ApiResponse[T]{
		Success: success,
		Data:    data,
		Message: message,
		Errors:  errors,
	})
}
