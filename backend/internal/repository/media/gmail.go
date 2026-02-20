package media

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

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
func (p *GmailProvider) SendMessage(ctx context.Context, target, message string) error {
	enabled := viper.GetBool("gmail.enabled")
	if !enabled {
		return nil
	}

	client, err := getGmailClient(ctx)
	if err != nil {
		return fmt.Errorf("unable to create gmail client: %v", err)
	}

	srv, err := gmail.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return fmt.Errorf("unable to retrieve Gmail client: %v", err)
	}

	user := "me"
	var msg gmail.Message

	boundary := "nagare-boundary"
	from := viper.GetString("gmail.from")
	if from == "" {
		from = "nagare-system@example.com"
	}
	subject := "Nagare Monitoring Alert"

	rawMsg := fmt.Sprintf("From: %s
"+
		"To: %s
"+
		"Subject: %s
"+
		"MIME-Version: 1.0
"+
		"Content-Type: multipart/alternative; boundary=%s
"+
		"
"+
		"--%s
"+
		"Content-Type: text/plain; charset="UTF-8"
"+
		"Content-Transfer-Encoding: base64
"+
		"
"+
		"%s
"+
		"--%s--",
		from, target, subject, boundary, boundary,
		base64.StdEncoding.EncodeToString([]byte(message)),
		boundary)

	msg.Raw = base64.URLEncoding.EncodeToString([]byte(rawMsg))

	_, err = srv.Users.Messages.Send(user, &msg).Do()
	if err != nil {
		return fmt.Errorf("unable to send message: %v", err)
	}

	return nil
}

// getGmailClient retrieves a token, saves the token, then returns the generated client.
func getGmailClient(ctx context.Context) (*http.Client, error) {
	credentialsFile := viper.GetString("gmail.credentials_file")
	if credentialsFile == "" {
		credentialsFile = "configs/gmail_credentials.json"
	}

	b, err := os.ReadFile(credentialsFile)
	if err != nil {
		return nil, fmt.Errorf("unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, gmail.GmailSendScope)
	if err != nil {
		return nil, fmt.Errorf("unable to parse client secret file to config: %v", err)
	}

	tokenFile := viper.GetString("gmail.token_file")
	if tokenFile == "" {
		tokenFile = "configs/gmail_token.json"
	}

	tok, err := tokenFromFile(tokenFile)
	if err != nil {
		// In a real CLI, we would prompt for authorization here.
		// For Nagare, we expect the token to be provided or managed externally.
		return nil, fmt.Errorf("gmail token not found or invalid at %s. Please authorize the application first", tokenFile)
	}
	return config.Client(ctx, tok), nil
}

// tokenFromFile retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// SendGmailServ sends an email using Gmail API (global version)
func SendGmailServ(ctx context.Context, to, subject, body string) error {
	enabled := viper.GetBool("gmail.enabled")
	if !enabled {
		return fmt.Errorf("gmail is disabled")
	}

	client, err := getGmailClient(ctx)
	if err != nil {
		return err
	}

	srv, err := gmail.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return err
	}

	from := viper.GetString("gmail.from")
	boundary := "nagare-boundary"
	rawMsg := fmt.Sprintf("From: %s
"+
		"To: %s
"+
		"Subject: %s
"+
		"MIME-Version: 1.0
"+
		"Content-Type: multipart/alternative; boundary=%s
"+
		"
"+
		"--%s
"+
		"Content-Type: text/plain; charset="UTF-8"
"+
		"Content-Transfer-Encoding: base64
"+
		"
"+
		"%s
"+
		"--%s--",
		from, to, subject, boundary, boundary,
		base64.StdEncoding.EncodeToString([]byte(body)),
		boundary)

	var msg gmail.Message
	msg.Raw = base64.URLEncoding.EncodeToString([]byte(rawMsg))

	_, err = srv.Users.Messages.Send("me", &msg).Do()
	return err
}

// SendGmailHTMLServ sends an HTML email using Gmail API
func SendGmailHTMLServ(ctx context.Context, to, subject, htmlBody string) error {
	enabled := viper.GetBool("gmail.enabled")
	if !enabled {
		return fmt.Errorf("gmail is disabled")
	}

	client, err := getGmailClient(ctx)
	if err != nil {
		return err
	}

	srv, err := gmail.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return err
	}

	from := viper.GetString("gmail.from")
	boundary := "nagare-boundary"
	rawMsg := fmt.Sprintf("From: %s
"+
		"To: %s
"+
		"Subject: %s
"+
		"MIME-Version: 1.0
"+
		"Content-Type: multipart/alternative; boundary=%s
"+
		"
"+
		"--%s
"+
		"Content-Type: text/html; charset="UTF-8"
"+
		"Content-Transfer-Encoding: base64
"+
		"
"+
		"%s
"+
		"--%s--",
		from, to, subject, boundary, boundary,
		base64.StdEncoding.EncodeToString([]byte(htmlBody)),
		boundary)

	var msg gmail.Message
	msg.Raw = base64.URLEncoding.EncodeToString([]byte(rawMsg))

	_, err = srv.Users.Messages.Send("me", &msg).Do()
	return err
}
