package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from .env file in parent directory
	if err := godotenv.Load("../.env"); err != nil {
		log.Println("Warning: Could not load .env file:", err)
	}

	clientID := os.Getenv("CLIENT_ID")
	fmt.Printf("CLIENT_ID from environment: '%s'\n", clientID)
	fmt.Printf("Length: %d\n", len(clientID))
	
	if clientID == "" {
		fmt.Println("CLIENT_ID is empty!")
	} else if clientID == "0d9edd02498fa29e97955ef964713296" {
		fmt.Println("This is the OLD CLIENT_ID!")
	} else if clientID == "758a938bc85320ceb23c40418e01618a" {
		fmt.Println("This is the NEW CLIENT_ID!")
	} else {
		fmt.Printf("Unknown CLIENT_ID: %s\n", clientID)
	}
}
