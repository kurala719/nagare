package service

import (
	"fmt"
	"net/smtp"

	"github.com/jordan-wright/email"
	"github.com/spf13/viper"
)

// SendEmailServ sends an email using the global SMTP configuration
func SendEmailServ(to, subject, body string) error {
	enabled := viper.GetBool("smtp.enabled")
	if !enabled {
		return fmt.Errorf("SMTP is disabled")
	}

	host := viper.GetString("smtp.host")
	port := viper.GetInt("smtp.port")
	username := viper.GetString("smtp.username")
	password := viper.GetString("smtp.password")
	from := viper.GetString("smtp.from")

	if host == "" || port == 0 || from == "" {
		return fmt.Errorf("SMTP configuration is incomplete")
	}

	e := email.NewEmail()
	e.From = from
	e.To = []string{to}
	e.Subject = subject
	e.Text = []byte(body)

	addr := fmt.Sprintf("%s:%d", host, port)
	auth := smtp.PlainAuth("", username, password, host)

	// Use TLS if port is 465, otherwise try standard send
	if port == 465 {
		return e.SendWithTLS(addr, auth, nil)
	}
	
	return e.Send(addr, auth)
}

// SendEmailHTMLServ sends an HTML email
func SendEmailHTMLServ(to, subject, htmlBody string) error {
	enabled := viper.GetBool("smtp.enabled")
	if !enabled {
		return fmt.Errorf("SMTP is disabled")
	}

	host := viper.GetString("smtp.host")
	port := viper.GetInt("smtp.port")
	username := viper.GetString("smtp.username")
	password := viper.GetString("smtp.password")
	from := viper.GetString("smtp.from")

	e := email.NewEmail()
	e.From = from
	e.To = []string{to}
	e.Subject = subject
	e.HTML = []byte(htmlBody)

	addr := fmt.Sprintf("%s:%d", host, port)
	auth := smtp.PlainAuth("", username, password, host)

	if port == 465 {
		return e.SendWithTLS(addr, auth, nil)
	}
	
	return e.Send(addr, auth)
}
