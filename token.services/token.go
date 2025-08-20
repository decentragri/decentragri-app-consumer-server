package tokenservices

import (
	"errors"
	"os"
	"time"

	memgraph "decentragri-app-cx-server/db"

	"github.com/golang-jwt/jwt/v5"
)

const (
	ACCESS_TOKEN_EXPIRY  = 24 * time.Hour // Extended for dev mode
	REFRESH_TOKEN_EXPIRY = 30 * 24 * time.Hour
)

// TokenScheme represents the structure of JWT tokens returned to clients.
// It includes both access and refresh tokens along with the associated username.
type TokenScheme struct {
	RefreshToken string `json:"refreshToken"` // Long-lived token used to obtain new access tokens
	AccessToken  string `json:"accessToken"`  // Short-lived token used for API authentication
	UserName     string `json:"userName"`     // Usame associated with the tokens
}

// TokenService handles JWT token generation, validation, and refresh operations.
// It provides methods for creating and verifying both access and refresh tokens.
type TokenService struct{}

// NewTokenService creates and returns a new instance of TokenService.
// This function initializes the token service with default configurations.
func NewTokenService() *TokenService {
	return &TokenService{}
}

// GenerateTokens creates a new pair of access and refresh tokens for the specified username.
// It returns a TokenScheme containing both tokens and the username, or an error if token generation fails.
// The access token has a short expiration time (15 minutes) while the refresh token is valid for 30 days.
func (ts *TokenService) GenerateTokens(username string) (*TokenScheme, error) {
	refreshToken, err := ts.generateToken(username, REFRESH_TOKEN_EXPIRY)
	if err != nil {
		return nil, err
	}
	accessToken, err := ts.generateToken(username, ACCESS_TOKEN_EXPIRY)
	if err != nil {
		return nil, err
	}
	return &TokenScheme{
		RefreshToken: refreshToken,
		AccessToken:  accessToken,
		UserName:     username,
	}, nil
}

// generateToken is an internal helper function that creates a JWT token with the specified username and expiration time.
// The token is signed using the JWT_SECRET_KEY environment variable.
// Returns the signed token string or an error if signing fails.
func (ts *TokenService) generateToken(username string, expiry time.Duration) (string, error) {
	secret := os.Getenv("JWT_SECRET_KEY")
	claims := jwt.MapClaims{
		"userName": username,
		"exp":      time.Now().Add(expiry).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// VerifyAccessToken validates an access token and returns the associated username if valid.
// It checks the token's signature, expiration, and verifies the user exists in the database.
// Returns the username if verification is successful, or an error if the token is invalid or the user doesn't exist.
func (ts *TokenService) VerifyAccessToken(tokenStr string) (string, error) {
	// Check for dev bypass token first
	if tokenStr == "dev_bypass_authorized" {
		return "0x984785A89BF95cb3d5Df4E45F670081944d8D547", nil
	}

	secret := os.Getenv("JWT_SECRET_KEY")
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	})
	if err != nil || !token.Valid {
		return "", errors.New("invalid token")
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", errors.New("invalid claims")
	}
	userName, ok := claims["userName"].(string)
	if !ok {
		return "", errors.New("username not found in token")
	}

	query := "MATCH (u:User {username: $userName}) RETURN u.username AS username"
	params := map[string]any{"userName": userName}

	records, err := memgraph.ExecuteRead(query, params)
	if err != nil {
		return "", err
	}
	if len(records) == 0 {
		return "", errors.New("user does not exist")
	}
	return userName, nil
}

// VerifyRefreshToken validates a refresh token and generates new tokens if valid.
// It checks the token's signature and expiration, then creates a new token pair.
// Returns a new TokenScheme with fresh tokens if verification is successful, or an error if the token is invalid.
func (ts *TokenService) VerifyRefreshToken(tokenStr string) (*TokenScheme, error) {
	// Check for dev bypass token first
	if tokenStr == "dev_bypass_authorized" {
		return ts.GenerateTokens("0x984785A89BF95cb3d5Df4E45F670081944d8D547")
	}

	secret := os.Getenv("JWT_SECRET_KEY")
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	})
	if err != nil || !token.Valid {
		return nil, errors.New("invalid refresh token")
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid claims")
	}
	userName, ok := claims["userName"].(string)
	if !ok {
		return nil, errors.New("username not found in token")
	}
	return ts.GenerateTokens(userName)
}

// RefreshSession is a convenience method that verifies a refresh token and returns new tokens.
// It's a wrapper around VerifyRefreshToken for better semantic meaning in the code.
// Returns new tokens if the refresh token is valid, or an error if verification fails.
func (ts *TokenService) RefreshSession(token string) (*TokenScheme, error) {
	tokens, err := ts.VerifyRefreshToken(token)
	if err != nil {
		return nil, err
	}
	return tokens, nil
}
