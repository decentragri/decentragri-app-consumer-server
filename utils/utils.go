package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"time"

	"math/big"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
)

// GetEnv loads environment variables from a .env file and retrieves the value of the specified environment variable.
// If the .env file cannot be loaded, the function logs a fatal error and terminates the program.
// The function returns the value of the environment variable corresponding to envName.
func GetEnv(envName string) string {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	env := os.Getenv(envName)

	// now do something with s3 or whatever
	return env
}

func EnginePost(uri string, body any) (string, error) {
	engineUri := GetEnv("ENGINE_URI")
	
	engineAccessToken := GetEnv("ENGINE_ACCESS_TOKEN")

	agent := fiber.Post(engineUri + uri)
	agent.Set("Authorization", "Bearer "+ engineAccessToken) // set Authorization header
	agent.JSON(body)                                        // set JSON body

	_, respBody, errs := agent.Bytes()
	if len(errs) > 0 {
		return "", errs[0]
	}

	return string(respBody), nil
}

func EngineGet(uri string) (string, error) {
	engineUri := GetEnv("ENGINE_URI")
	engineAccessToken := os.Getenv("ENGINE_ACCESS_TOKEN")
	fmt.Println("engine access token:", engineAccessToken)
	agent := fiber.Get(engineUri + uri)
	agent.Set("Authorization", "Bearer "+engineAccessToken) // set Authorization header

	_, respBody, errs := agent.Bytes()
	if len(errs) > 0 {
		return "", errs[0]
	}

	return string(respBody), nil
}

// ParseEther converts a string representation of Ether (e.g., "1.23") to its value in Wei as *big.Int.
// It assumes 18 decimals (1 Ether = 10^18 Wei).
func ParseEther(ether string) (*big.Int, error) {
	parts := strings.SplitN(ether, ".", 2)
	intPart := parts[0]
	decPart := ""
	if len(parts) == 2 {
		decPart = parts[1]
		if len(decPart) > 18 {
			decPart = decPart[:18] // trim to 18 decimals
		}
	}
	// Pad decimal part to 18 digits
	decPart = decPart + strings.Repeat("0", 18-len(decPart))

	weiStr := intPart + decPart
	wei := new(big.Int)
	_, ok := wei.SetString(weiStr, 10)
	if !ok {
		return nil, ErrInvalidEtherString
	}
	return wei, nil
}

var ErrInvalidEtherString = fmt.Errorf("invalid ether string")

// uploadPicBuffer uploads an image buffer to IPFS via thirdweb storage and returns the resulting URI.
func UploadPicBuffer(ctx context.Context, buffer []byte, fileName string) (string, error) {
	// Prepare multipart form
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	fw, err := w.CreateFormFile("file", fileName)
	if err != nil {
		return "", err
	}
	_, err = fw.Write(buffer)
	if err != nil {
		return "", err
	}
	w.Close()

	endpoint := "https://storage.thirdweb.com/ipfs/upload"
	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, &b)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())

	// Optionally: set thirdweb clientId and secretKey if required
	clientId := os.Getenv("THIRDWEB_CLIENT_ID")
	secretKey := os.Getenv("SECRET_KEY")
	if clientId != "" {
		req.Header.Set("x-client-id", clientId)
	}
	if secretKey != "" {
		req.Header.Set("x-secret-key", secretKey)
	}

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to upload to IPFS: %s", resp.Status)
	}

	var result struct {
		IpfsHash string `json:"IpfsHash"`
	}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return "", err
	}
	if result.IpfsHash == "" {
		return "", fmt.Errorf("no IpfsHash returned from upload")
	}
	return "ipfs://" + result.IpfsHash + "/" + fileName, nil
}
