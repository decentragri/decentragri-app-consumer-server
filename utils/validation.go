package utils

import (
	"net"
	"regexp"
	"strconv"
	"strings"
)

// ValidationRules contains common validation patterns
var ValidationRules = struct {
	// EthereumAddress matches valid Ethereum addresses
	EthereumAddress *regexp.Regexp
	// AlphaNumeric matches alphanumeric characters with spaces, hyphens, and underscores
	AlphaNumeric *regexp.Regexp
	// SafeString matches safe strings for general use
	SafeString *regexp.Regexp
}{
	EthereumAddress: regexp.MustCompile(`^0x[a-fA-F0-9]{40}$`),
	AlphaNumeric:    regexp.MustCompile(`^[a-zA-Z0-9\s\-_]+$`),
	SafeString:      regexp.MustCompile(`^[a-zA-Z0-9\s\-_.@]+$`),
}

// ValidateEthereumAddress validates if a string is a valid Ethereum address
func ValidateEthereumAddress(address string) bool {
	if address == "" {
		return false
	}
	return ValidationRules.EthereumAddress.MatchString(address)
}

// ValidateFarmName validates farm name input
func ValidateFarmName(farmName string) bool {
	if farmName == "" || len(farmName) > 100 {
		return false
	}
	return ValidationRules.AlphaNumeric.MatchString(farmName)
}

// ValidatePagination validates pagination parameters
func ValidatePagination(pageStr, limitStr string) (int, int, error) {
	page := 1
	limit := 10

	if pageStr != "" {
		p, err := strconv.Atoi(pageStr)
		if err != nil || p < 1 || p > 1000 {
			return 0, 0, NewValidationError("page", "must be between 1 and 1000")
		}
		page = p
	}

	if limitStr != "" {
		l, err := strconv.Atoi(limitStr)
		if err != nil || l < 1 || l > 100 {
			return 0, 0, NewValidationError("limit", "must be between 1 and 100")
		}
		limit = l
	}

	return page, limit, nil
}

// SanitizeInput removes potentially dangerous characters and trims whitespace
func SanitizeInput(input string) string {
	// Remove null bytes and control characters
	cleaned := strings.ReplaceAll(input, "\x00", "")
	cleaned = strings.TrimSpace(cleaned)

	// Remove potential SQL injection patterns
	dangerous := []string{
		"<script", "</script>", "javascript:", "data:", "vbscript:",
		"onload=", "onerror=", "onclick=", "eval(", "expression(",
	}

	lowerInput := strings.ToLower(cleaned)
	for _, pattern := range dangerous {
		if strings.Contains(lowerInput, pattern) {
			return ""
		}
	}

	return cleaned
}

// ValidateIPAddress validates if a string is a valid IP address
func ValidateIPAddress(ip string) bool {
	return net.ParseIP(ip) != nil
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return e.Field + ": " + e.Message
}

// NewValidationError creates a new validation error
func NewValidationError(field, message string) ValidationError {
	return ValidationError{
		Field:   field,
		Message: message,
	}
}

// ValidateTokenID validates token ID format
func ValidateTokenID(tokenID string) bool {
	if tokenID == "" || len(tokenID) > 20 {
		return false
	}
	// Token IDs should be numeric
	_, err := strconv.ParseUint(tokenID, 10, 64)
	return err == nil
}

// ValidateContractAddress validates smart contract address
func ValidateContractAddress(address string) bool {
	return ValidateEthereumAddress(address)
}

// RateLimiting validates rate limiting parameters
func ValidateRateLimit(requests, window int) bool {
	return requests > 0 && requests <= 1000 && window > 0 && window <= 3600
}
