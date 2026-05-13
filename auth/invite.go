package auth

import (
	"bytes"
	"html/template"
	"keyclub-api/email"
	"keyclub-api/internal"
)

// first, get the template done
// make a new email message and send it to the person
func SendInvite(emailAddress string, firstName string, roleLevel int, config internal.Config) error {
	template, err := template.ParseFiles(config.EmailTemplatePath + "/invite.html")
	if err != nil {
		return err
	}

	buf := bytes.Buffer{}
	if err := template.Execute(&buf, email.InviteEmail{
		Subject:         "JHS Key Club Invitation",
		FirstName:       firstName,
		RoleLevelString: GetRoleLevelString(roleLevel),
		MagicLink:       "https://jhskeyclub.com",
	}); err != nil {
		return err
	}

	htmlBody := buf.String()
	emailMessage := email.EmailMessage{
		To:       emailAddress,
		Subject:  "JHS Key Club Invitation",
		HTMLBody: htmlBody,
	}

	return email.SendEmail(emailMessage, config.SMTPConfig)
}
