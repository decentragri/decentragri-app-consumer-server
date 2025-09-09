package utils

import (
	"log"
	"runtime"

	"github.com/gofiber/fiber/v2"
)

// ErrorResponse represents a standardized error response structure
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

// HandleError logs the error internally and returns a sanitized error to the client
func HandleError(c *fiber.Ctx, err error, userMessage string, statusCode int) error {
	// Log internal error with context
	pc, file, line, _ := runtime.Caller(1)
	funcName := runtime.FuncForPC(pc).Name()
	log.Printf("Error in %s (%s:%d): %v", funcName, file, line, err)

	// Return sanitized error to client
	return c.Status(statusCode).JSON(ErrorResponse{
		Error:   userMessage,
		Message: "Please contact support if this issue persists",
	})
}

// HandleValidationError handles input validation errors
func HandleValidationError(c *fiber.Ctx, fieldName string) error {
	log.Printf("Validation error: invalid %s provided by client %s", fieldName, c.IP())

	return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
		Error: "Invalid input provided",
		Code:  "VALIDATION_ERROR",
	})
}

// HandleAuthError handles authentication and authorization errors
func HandleAuthError(c *fiber.Ctx, err error) error {
	log.Printf("Authentication error from client %s: %v", c.IP(), err)

	return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse{
		Error: "Authentication failed",
		Code:  "AUTH_ERROR",
	})
}

// HandleInternalError handles internal server errors
func HandleInternalError(c *fiber.Ctx, err error, operation string) error {
	pc, file, line, _ := runtime.Caller(1)
	funcName := runtime.FuncForPC(pc).Name()
	log.Printf("Internal error in %s (%s:%d) during %s: %v", funcName, file, line, operation, err)

	return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
		Error: "Internal server error",
		Code:  "INTERNAL_ERROR",
	})
}
