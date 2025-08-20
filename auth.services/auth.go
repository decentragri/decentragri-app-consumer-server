package authservices

import (
	memgraph "decentragri-app-cx-server/db"
	tokenServices "decentragri-app-cx-server/token.services"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/gofiber/fiber/v2"
)

// CheckDevBypass checks if the request has a valid dev bypass token
// Returns true if bypass is valid, false otherwise
func CheckDevBypass(c *fiber.Ctx) bool {
	devBypassToken := os.Getenv("DEV_BYPASS_TOKEN")
	if devBypassToken == "" {
		return false // No dev token configured
	}

	// Check for bypass token in header
	bypassHeader := c.Get("X-Dev-Bypass-Token")
	if bypassHeader == devBypassToken {
		fmt.Println("Dev bypass token used for request:", c.Method(), c.Path())
		return true
	}

	// Check for bypass token in query parameter (alternative method)
	bypassQuery := c.Query("dev_bypass_token")
	if bypassQuery == devBypassToken {
		fmt.Println("Dev bypass token used for request:", c.Method(), c.Path())
		return true
	}

	return false
}

// GetNonce - Generate nonce for wallet authentication with validation
func GetNonce(walletAddress string) (GetNonceResponse, error) {
	// Validate wallet address
	if walletAddress == "" {
		return GetNonceResponse{}, errors.New("wallet address is required")
	}

	// Generate nonce
	nonce, err := GenerateNonce(walletAddress)
	if err != nil {
		return GetNonceResponse{}, errors.New("failed to generate nonce: " + err.Error())
	}

	response := GetNonceResponse{
		Nonce:   nonce,
		Message: "Please sign this nonce with your wallet to authenticate",
	}

	return response, nil
}

// AuthenticateWallet - Verify signature and handle login/register automatically with validation
func AuthenticateWallet(request AuthenticateWalletRequest) (AuthenticateWalletResponse, error) {
	// Validate required fields
	if request.WalletAddress == "" {
		return AuthenticateWalletResponse{}, errors.New("wallet address is required")
	}
	if request.Nonce == "" {
		return AuthenticateWalletResponse{}, errors.New("nonce is required")
	}
	if request.SignatureHex == "" {
		return AuthenticateWalletResponse{}, errors.New("signature is required")
	}
	if request.DeviceId == "" {
		return AuthenticateWalletResponse{}, errors.New("device ID is required")
	}
	// First verify the signature
	isVerified, err := VerifySignature(request.WalletAddress, request.Nonce, request.SignatureHex)
	if err != nil {
		return AuthenticateWalletResponse{}, errors.New("signature verification failed: " + err.Error())
	}
	if !isVerified {
		return AuthenticateWalletResponse{}, errors.New("signature verification failed")
	}

	// Check if user exists
	query := `MATCH (u:User {username: $username})`
	params := map[string]any{"username": request.WalletAddress}
	records, err := memgraph.ExecuteRead(query, params)
	if err != nil {
		return AuthenticateWalletResponse{}, errors.New("database error: " + err.Error())
	}

	isNewUser := len(records) == 0

	// If new user, create them
	if isNewUser {
		createQuery := `CREATE (u:User {
			username: $username,
			createdAt: timestamp(),
			walletAddress: $walletAddress,
			deviceId: $deviceId})
		RETURN u.username AS username`
		createParams := map[string]any{"username": request.WalletAddress, "walletAddress": request.WalletAddress, "deviceId": request.DeviceId}
		_, err = memgraph.ExecuteWrite(createQuery, createParams)
		if err != nil {
			return AuthenticateWalletResponse{}, errors.New("failed to create user: " + err.Error())
		}
	}

	// Generate tokens for both new and existing users
	tokenService := tokenServices.NewTokenService()
	token, err := tokenService.GenerateTokens(request.WalletAddress)
	if err != nil {
		return AuthenticateWalletResponse{}, errors.New("failed to generate tokens: " + err.Error())
	}

	var message string
	if isNewUser {
		message = "Welcome! Your account has been created successfully."
	} else {
		message = "Welcome back! You have been logged in successfully."
	}

	response := AuthenticateWalletResponse{
		WalletAddress: request.WalletAddress,
		Tokens:        *token,
		IsNewUser:     isNewUser,
		Message:       message,
		LoginType:     "wallet", // Indicate this is a wallet login
	}

	return response, nil
}

// RefreshSession - Verify refresh token and generate new tokens
func RefreshSession(refreshToken string) (tokenServices.TokenScheme, error) {
	// Validate refresh token
	if refreshToken == "" {
		return tokenServices.TokenScheme{}, errors.New("refresh token is required")
	}

	tokenService := tokenServices.NewTokenService()

	// Verify refresh token and generate new tokens
	tokens, err := tokenService.VerifyRefreshToken(refreshToken)
	if err != nil {
		return tokenServices.TokenScheme{}, errors.New("invalid or expired refresh token: " + err.Error())
	}

	return *tokens, nil
}

// VerifyGoogleToken verifies the Google ID token with Google's servers
func VerifyGoogleToken(idToken string) (*GoogleTokenInfo, error) {
	if idToken == "" {
		return nil, errors.New("ID token is required")
	}

	// Google's token verification endpoint
	verifyURL := fmt.Sprintf("https://oauth2.googleapis.com/tokeninfo?id_token=%s", idToken)

	// Make request to Google
	req := fiber.Get(verifyURL)
	status, body, errs := req.Bytes()
	if len(errs) > 0 {
		return nil, fmt.Errorf("failed to verify token with Google: %v", errs[0])
	}

	if status != 200 {
		return nil, fmt.Errorf("Google token verification failed with status %d: %s", status, string(body))
	}

	// Parse response
	var tokenInfo GoogleTokenInfo
	if err := json.Unmarshal(body, &tokenInfo); err != nil {
		return nil, fmt.Errorf("failed to parse Google response: %w", err)
	}

	// Verify the audience (client ID)
	expectedClientId := os.Getenv("GOOGLE_CLIENT_ID")
	if expectedClientId == "" {
		return nil, errors.New("GOOGLE_CLIENT_ID environment variable not set")
	}

	if tokenInfo.Aud != expectedClientId {
		return nil, errors.New("invalid audience in token")
	}

	// Verify issuer
	if tokenInfo.Iss != "accounts.google.com" && tokenInfo.Iss != "https://accounts.google.com" {
		return nil, errors.New("invalid issuer in token")
	}

	// Verify email is verified
	if !tokenInfo.EmailVerified {
		return nil, errors.New("email not verified by Google")
	}

	return &tokenInfo, nil
}

// AuthenticateGoogle handles Google OAuth authentication
func AuthenticateGoogle(request AuthenticateGoogleRequest) (AuthenticateGoogleResponse, error) {
	// Validate required fields
	if request.IdToken == "" {
		return AuthenticateGoogleResponse{}, errors.New("ID token is required")
	}
	if request.DeviceId == "" {
		return AuthenticateGoogleResponse{}, errors.New("device ID is required")
	}

	// Verify the Google ID token
	tokenInfo, err := VerifyGoogleToken(request.IdToken)
	if err != nil {
		return AuthenticateGoogleResponse{}, fmt.Errorf("Google token verification failed: %w", err)
	}

	// Use Google ID as username for consistency
	username := tokenInfo.Sub

	// Check if user exists in database
	query := `MATCH (u:User {googleId: $googleId})`
	params := map[string]any{"googleId": tokenInfo.Sub}
	records, err := memgraph.ExecuteRead(query, params)
	if err != nil {
		return AuthenticateGoogleResponse{}, fmt.Errorf("database error: %w", err)
	}

	isNewUser := len(records) == 0
	var walletAddress string

	// If new user, create them
	if isNewUser {
		walletAddress, err = CreateWallet(username) // Create a wallet for the new user
		if err != nil {
			return AuthenticateGoogleResponse{}, fmt.Errorf("failed to create wallet: %w", err)
		}

		createQuery := `CREATE (u:User {
			username: $username,
			googleId: $googleId,
			email: $email,
			name: $name,
			picture: $picture,
			createdAt: timestamp(),
			deviceId: $deviceId,
			walletAddress: $walletAddress,
			authProvider: 'google'
		}) RETURN u.username AS username`

		createParams := map[string]any{
			"username":      walletAddress,
			"googleId":      tokenInfo.Sub,
			"email":         tokenInfo.Email,
			"name":          tokenInfo.Name,
			"picture":       tokenInfo.Picture,
			"deviceId":      request.DeviceId,
			"walletAddress": walletAddress,
		}

		_, err = memgraph.ExecuteWrite(createQuery, createParams)
		if err != nil {
			return AuthenticateGoogleResponse{}, fmt.Errorf("failed to create user: %w", err)
		}
	} else {
		// Update existing user's info and get wallet address
		updateQuery := `MATCH (u:User {googleId: $googleId})
			SET u.email = $email, u.name = $name, u.picture = $picture, u.deviceId = $deviceId
			RETURN u.walletAddress AS walletAddress`

		updateParams := map[string]any{
			"googleId": tokenInfo.Sub,
			"email":    tokenInfo.Email,
			"name":     tokenInfo.Name,
			"picture":  tokenInfo.Picture,
			"deviceId": request.DeviceId,
		}

		records, err := memgraph.ExecuteRead(updateQuery, updateParams)
		if err != nil {
			return AuthenticateGoogleResponse{}, fmt.Errorf("failed to update user: %w", err)
		}

		if len(records) > 0 {
			if addr, ok := records[0].Get("walletAddress"); ok {
				if walletAddr, ok := addr.(string); ok {
					walletAddress = walletAddr
				}
			}
		}
	}

	// Generate JWT tokens
	tokenService := tokenServices.NewTokenService()
	tokens, err := tokenService.GenerateTokens(username)
	if err != nil {
		return AuthenticateGoogleResponse{}, fmt.Errorf("failed to generate tokens: %w", err)
	}

	var message string
	if isNewUser {
		message = "Welcome! Your Google account has been linked successfully."
	} else {
		message = "Welcome back! You have been logged in with Google."
	}

	response := AuthenticateGoogleResponse{
		WalletAddress: walletAddress,
		Tokens:        *tokens,
		IsNewUser:     isNewUser,
		Message:       message,
		LoginType:     "google", // Indicate this is a Google login
	}

	return response, nil
}
