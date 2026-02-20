package service

import (
	"context"
	"fmt"
	"nagare/internal/repository/media"

	"github.com/spf13/viper"
)

// SendEmailServ sends an email using the Gmail API
func SendEmailServ(to, subject, body string) error {
	enabled := viper.GetBool("gmail.enabled")
	if !enabled {
		return fmt.Errorf("gmail is disabled")
	}

	return media.SendGmailServ(context.Background(), to, subject, body)
}

// SendEmailHTMLServ sends an HTML email using the Gmail API
func SendEmailHTMLServ(to, subject, htmlBody string) error {
	enabled := viper.GetBool("gmail.enabled")
	if !enabled {
		return fmt.Errorf("gmail is disabled")
	}

	return media.SendGmailHTMLServ(context.Background(), to, subject, htmlBody)
}
