package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
)

func main() {
	// Read credentials
	b, err := os.ReadFile("configs/gmail_credentials.json")
	if err != nil {
		log.Fatal("Unable to read credentials file:", err)
	}

	config, err := google.ConfigFromJSON(b, gmail.GmailSendScope)
	if err != nil {
		log.Fatal("Unable to parse credentials:", err)
	}

	// Generate auth URL
	authURL := config.AuthCodeURL("state", oauth2.AccessTypeOffline)
	fmt.Printf("Go to this URL to authorize:\n%v\n\n", authURL)

	// Get authorization code
	var code string
	fmt.Print("Enter the authorization code: ")
	fmt.Scanln(&code)

	// Exchange for token
	tok, err := config.Exchange(context.Background(), code)
	if err != nil {
		log.Fatal("Unable to retrieve token:", err)
	}

	// Save token
	f, err := os.Create("configs/gmail_token.json")
	if err != nil {
		log.Fatal("Unable to create token file:", err)
	}
	defer f.Close()

	json.NewEncoder(f).Encode(tok)
	fmt.Println("Token saved successfully to configs/gmail_token.json!")
}
