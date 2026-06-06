package email

import (
	"bytes"
	"fmt"
	"html/template"
	"net/smtp"
	"strings"
)

// SendEmail sends an email to the given address with the given subject and HTML body.
// Takes an EmailMessage object (so you have to have done the templating beforehand)
func SendEmail(emailMessage EmailMessage, smtpConfig SMTPConfig) error {
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

func SendPendingLoginEmail(emailTemplate PendingLoginEmailTemplate, to string, smtpConfig SMTPConfig) error {
	template, err := template.ParseFiles(smtpConfig.EmailTemplatePath + "/login.html")
	if err != nil {
		return err
	}

	buf := bytes.Buffer{}
	if err := template.Execute(&buf, emailTemplate); err != nil {
		return err
	}

	htmlBody := buf.String()
	emailMessage := EmailMessage{
		To:       to,
		Subject:  emailTemplate.Subject,
		HTMLBody: htmlBody,
	}
	return SendEmail(emailMessage, smtpConfig)
}
