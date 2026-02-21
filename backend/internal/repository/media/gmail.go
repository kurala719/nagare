package media

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/spf13/viper"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

// GmailProvider sends emails using Gmail API
type GmailProvider struct{}

// NewGmailProvider creates a Gmail provider
func NewGmailProvider() *GmailProvider {
	return &GmailProvider{}
}

// SendMessage sends an email via Gmail API
// The message argument is treated as the email body.
// The subject defaults to "Nagare Monitoring Alert".
func (p *GmailProvider) SendMessage(ctx context.Context, target, message string) error {
	return sendGmailMessage(ctx, target, "Nagare Monitoring Alert", message, false)
}

// SendGmailServ sends an email using Gmail API (global version)
func SendGmailServ(ctx context.Context, to, subject, body string) error {
	return sendGmailMessage(ctx, to, subject, body, false)
}

// SendGmailHTMLServ sends an HTML email using Gmail API
func SendGmailHTMLServ(ctx context.Context, to, subject, htmlBody string) error {
	return sendGmailMessage(ctx, to, subject, htmlBody, true)
}

// sendGmailMessage is the core logic for sending emails
func sendGmailMessage(ctx context.Context, to, subject, body string, isHTML bool) error {
	// Validate inputs
	if to == "" {
		return fmt.Errorf("recipient email address cannot be empty")
	}
	if subject == "" {
		return fmt.Errorf("email subject cannot be empty")
	}
	if body == "" {
		return fmt.Errorf("email body cannot be empty")
	}

	if !viper.GetBool("gmail.enabled") {
		return fmt.Errorf("gmail is disabled in configuration (set gmail.enabled=true)")
	}

	// Get Gmail client
	client, err := getGmailClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to create gmail client: %w", err)
	}

	// Create Gmail service
	srv, err := gmail.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return fmt.Errorf("failed to create gmail service: %w", err)
	}

	// Get sender email from config
	from := viper.GetString("gmail.from")
	if from == "" {
		from = "nagare-system@example.com"
	}

	// Prepare MIME message
	contentType := "text/plain"
	if isHTML {
		contentType = "text/html"
	}

	// Construct the email message
	var msgBuffer strings.Builder
	msgBuffer.WriteString(fmt.Sprintf("From: %s\r\n", from))
	msgBuffer.WriteString(fmt.Sprintf("To: %s\r\n", to))
	msgBuffer.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	msgBuffer.WriteString("MIME-Version: 1.0\r\n")
	msgBuffer.WriteString(fmt.Sprintf("Content-Type: %s; charset=\"UTF-8\"\r\n", contentType))
	msgBuffer.WriteString("Content-Transfer-Encoding: base64\r\n")
	msgBuffer.WriteString("\r\n")
	msgBuffer.WriteString(base64.StdEncoding.EncodeToString([]byte(body)))

	var msg gmail.Message
	msg.Raw = base64.URLEncoding.EncodeToString([]byte(msgBuffer.String()))

	// Send message
	_, err = srv.Users.Messages.Send("me", &msg).Do()
	if err != nil {
		return fmt.Errorf("failed to send email to %s: %w", to, err)
	}

	return nil
}

// getGmailClient retrieves a token, saves the token, then returns the generated client.
func getGmailClient(ctx context.Context) (*http.Client, error) {
	// Get credentials file path from config
	credentialsFile := viper.GetString("gmail.credentials_file")
	if credentialsFile == "" {
		credentialsFile = "configs/gmail_credentials.json"
	}

	// Read credentials file
	b, err := os.ReadFile(credentialsFile)
	if err != nil {
		return nil, fmt.Errorf("unable to read credentials file at %s: %w (ensure gmail.credentials_file is set in config and file exists)", credentialsFile, err)
	}

	// Parse credentials as Google OAuth2 config
	config, err := google.ConfigFromJSON(b, gmail.GmailSendScope)
	if err != nil {
		return nil, fmt.Errorf("unable to parse credentials file at %s: %w (ensure file contains valid Google OAuth2 credentials JSON)", credentialsFile, err)
	}

	// Get token file path from config
	tokenFile := viper.GetString("gmail.token_file")
	if tokenFile == "" {
		tokenFile = "configs/gmail_token.json"
	}

	// Load token from file
	tok, err := tokenFromFile(tokenFile)
	if err != nil {
		return nil, fmt.Errorf("gmail token not found or invalid at %s: %w. Please run Gmail authorization first (provide gmail.token_file path in config)", tokenFile, err)
	}

	// Create HTTP client with token
	return config.Client(ctx, tok), nil
}

// tokenFromFile retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, fmt.Errorf("unable to open token file at %s: %w", file, err)
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	if err != nil {
		return nil, fmt.Errorf("unable to decode token file at %s: %w", file, err)
	}
	// Check for access token but also consider refresh token might be present
	if tok.AccessToken == "" && tok.RefreshToken == "" {
		return nil, fmt.Errorf("token file at %s contains no access token or refresh token", file)
	}
	return tok, nil
}
