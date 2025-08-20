package authservices

import (
	tokenServices "decentragri-app-cx-server/token.services"
)

type NonCustodialRegistration struct {
	WalletAddress string `json:"walletAddress"`
	Nonce         string `json:"nonce"`
	SignatureHex  string `json:"signatureHex"`
	DeviceId      string `json:"deviceId"`
}

// GetNonceRequest represents the request payload for getting a nonce
type GetNonceRequest struct {
	WalletAddress string `json:"walletAddress"`
}

// GetNonceResponse represents the response payload for getting a nonce
type GetNonceResponse struct {
	Nonce   string `json:"nonce"`
	Message string `json:"message"`
}

// AuthenticateWalletRequest represents the request payload for wallet authentication
type AuthenticateWalletRequest struct {
	WalletAddress string `json:"walletAddress"`
	Nonce         string `json:"nonce"`
	SignatureHex  string `json:"signatureHex"`
	DeviceId      string `json:"deviceId"`
}

// AuthenticateWalletResponse represents the response payload for wallet authentication
type AuthenticateWalletResponse struct {
	WalletAddress string                    `json:"walletAddress"`
	Tokens        tokenServices.TokenScheme `json:"tokens"`
	IsNewUser     bool                      `json:"isNewUser"`
	Message       string                    `json:"message"`
	LoginType     string                    `json:"loginType"` // "wallet" or "google"
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// RefreshTokenRequest represents the request payload for refreshing tokens
type RefreshTokenRequest struct {
	RefreshToken string `json:"refreshToken"`
}

// AuthenticateGoogleRequest represents the request payload for Google OAuth authentication
type AuthenticateGoogleRequest struct {
	IdToken  string `json:"idToken"`  // Google ID token from the client
	DeviceId string `json:"deviceId"` // Device ID for tracking
}

// AuthenticateGoogleResponse represents the response payload for Google OAuth authentication
type AuthenticateGoogleResponse struct {
	GoogleId  string                    `json:"googleId"`  // User's Google ID
	Email     string                    `json:"email"`     // User's email
	Name      string                    `json:"name"`      // User's full name
	Picture   string                    `json:"picture"`   // Profile picture URL
	Tokens    tokenServices.TokenScheme `json:"tokens"`    // JWT tokens
	IsNewUser   bool                    `json:"isNewUser"`   // Whether this is a new user
	LoginType   string                  `json:"loginType"`   // Type of login ("google")
	Message     string                  `json:"message"`     // Success message
	WalletAddress string                `json:"walletAddress"` // User's wallet address
}

// RefreshTokenResponse represents the response payload for refreshing tokens
type RefreshTokenResponse struct {
	RefreshToken string `json:"refreshToken"` // Include refresh token in response
	AccessToken  string `json:"accessToken"`  // Include access token in response
}

// GoogleTokenInfo represents the structure of Google's token verification response
type GoogleTokenInfo struct {
	Sub           string `json:"sub"`            // User's unique Google ID
	Email         string `json:"email"`          // User's email
	EmailVerified bool   `json:"email_verified"` // Whether email is verified
	Name          string `json:"name"`           // User's full name
	Picture       string `json:"picture"`        // User's profile picture URL
	GivenName     string `json:"given_name"`     // User's first name
	FamilyName    string `json:"family_name"`    // User's last name
	Aud           string `json:"aud"`            // Audience (your client ID)
	Iss           string `json:"iss"`            // Issuer (should be accounts.google.com)
	Exp           int64  `json:"exp"`            // Expiration time
	Iat           int64  `json:"iat"`            // Issued at time
}