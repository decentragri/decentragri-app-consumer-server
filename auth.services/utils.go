package authservices

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"decentragri-app-cx-server/utils"

	"github.com/ethereum/go-ethereum/crypto"
)

// In production, store this in DB or Redis (keyed by wallet)
var nonceStore = map[string]string{}
var nonceMutex = sync.RWMutex{}

const NonceExpirationSeconds = 300 // 5 minutes

// GenerateNonce creates a random hex nonce
func GenerateNonce(wallet string) (string, error) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	nonce := hex.EncodeToString(b)

	// store nonce with timestamp for later verification
	nonceMutex.Lock()
	nonceStore[strings.ToLower(wallet)] = fmt.Sprintf("%s:%d", nonce, time.Now().Unix())
	nonceMutex.Unlock()

	return nonce, nil
}

// VerifySignature checks if signature belongs to the wallet
func VerifySignature(walletAddress string, nonce string, signatureHex string) (bool, error) {
	walletAddress = strings.ToLower(walletAddress)

	// Get stored nonce with thread safety
	nonceMutex.RLock()
	stored, exists := nonceStore[walletAddress]
	nonceMutex.RUnlock()

	if !exists {
		return false, errors.New("nonce not found or expired")
	}

	parts := strings.Split(stored, ":")
	if len(parts) != 2 {
		return false, errors.New("invalid nonce format")
	}

	if parts[0] != nonce {
		return false, errors.New("nonce mismatch")
	}

	// Check if nonce is expired
	timestamp, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return false, errors.New("invalid timestamp format")
	}

	if time.Now().Unix()-timestamp > NonceExpirationSeconds {
		// Clean up expired nonce
		nonceMutex.Lock()
		delete(nonceStore, walletAddress)
		nonceMutex.Unlock()
		return false, errors.New("nonce expired")
	}

	// Decode signature
	sig, err := hex.DecodeString(strings.TrimPrefix(signatureHex, "0x"))
	if err != nil {
		return false, errors.New("invalid signature hex")
	}

	// Validate signature length
	if len(sig) != 65 {
		return false, errors.New("invalid signature length")
	}

	// Adjust V value for recovery if needed
	if sig[64] != 27 && sig[64] != 28 {
		if sig[64] == 0 || sig[64] == 1 {
			sig[64] += 27
		} else {
			return false, errors.New("invalid recovery id")
		}
	}

	// Ethereum personal_sign uses prefix
	msg := fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(nonce), nonce)
	hash := crypto.Keccak256Hash([]byte(msg))

	// Recover public key from signature
	pubKey, err := crypto.SigToPub(hash.Bytes(), sig)
	if err != nil {
		return false, err
	}

	recoveredAddr := crypto.PubkeyToAddress(*pubKey).Hex()

	// Clean up nonce after successful verification to prevent replay attacks
	if strings.EqualFold(recoveredAddr, walletAddress) {
		nonceMutex.Lock()
		delete(nonceStore, walletAddress)
		nonceMutex.Unlock()
		return true, nil
	}

	return false, nil
}

// CleanupExpiredNonces removes expired nonces from memory
// Call this periodically to prevent memory leaks
func CleanupExpiredNonces() {
	nonceMutex.Lock()
	defer nonceMutex.Unlock()

	now := time.Now().Unix()
	for wallet, stored := range nonceStore {
		parts := strings.Split(stored, ":")
		if len(parts) == 2 {
			if timestamp, err := strconv.ParseInt(parts[1], 10, 64); err == nil {
				if now-timestamp > NonceExpirationSeconds {
					delete(nonceStore, wallet)
				}
			}
		}
	}
}

// CreateWalletRequest represents the request payload for creating a wallet
type CreateWalletRequest struct {
	Label string `json:"label"`
	Type  string `json:"type"`
}

// CreateWallet creates a new wallet using Thirdweb Engine
func CreateWallet(username string) (string, error) {
	requestBody := CreateWalletRequest{
		Label: username,
		Type:  "smart:local",
	}

	response, err := utils.EnginePost("/backend-wallet/create", requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to create wallet: %w", err)
	}

	return response, nil
}
