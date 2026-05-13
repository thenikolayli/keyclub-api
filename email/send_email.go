package email

import (
	"fmt"
	"keyclub-api/internal"
	"net/smtp"
	"strings"
)

// SendEmail sends an email to the given address with the given subject and HTML body.
// Takes an EmailMessage object (so you have to have done the templating beforehand)
func SendEmail(emailMessage EmailMessage, smtpConfig internal.SMTPConfig) error {
	message := strings.Join([]string{
		fmt.Sprintf(`From: "JHS Key Club" <%s>`, smtpConfig.User),
		fmt.Sprintf("To: %s", emailMessage.To),
		fmt.Sprintf("Subject: %s", emailMessage.Subject),
		"MIME-Version: 1.0",
		"Content-Type: text/html; charset=\"UTF-8\"",
		"Content-Transfer-Encoding: 7bit",
		"",
		emailMessage.HTMLBody,
	}, "\r\n")

	return smtp.SendMail(smtpConfig.Address, smtpConfig.Auth, smtpConfig.User, []string{emailMessage.To}, []byte(message))
}
