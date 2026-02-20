package media

import (
	"context"
	"fmt"
	"net/smtp"

	"github.com/jordan-wright/email"
	"github.com/spf13/viper"
)

// EmailProvider sends emails using SMTP
type EmailProvider struct{}

// NewEmailProvider creates an email provider
func NewEmailProvider() *EmailProvider {
	return &EmailProvider{}
}

// SendMessage sends an email
func (p *EmailProvider) SendMessage(ctx context.Context, target, message string) error {
	enabled := viper.GetBool("smtp.enabled")
	if !enabled {
		return nil // Quietly skip if disabled
	}

	host := viper.GetString("smtp.host")
	port := viper.GetInt("smtp.port")
	username := viper.GetString("smtp.username")
	password := viper.GetString("smtp.password")
	from := viper.GetString("smtp.from")

	if host == "" || port == 0 || from == "" {
		return fmt.Errorf("SMTP configuration incomplete")
	}

	e := email.NewEmail()
	e.From = from
	e.To = []string{target}
	e.Subject = "Nagare Monitoring Alert"
	e.Text = []byte(message)

	addr := fmt.Sprintf("%s:%d", host, port)
	auth := smtp.PlainAuth("", username, password, host)

	if port == 465 {
		return e.SendWithTLS(addr, auth, nil)
	}
	return e.Send(addr, auth)
}
