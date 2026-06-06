package email

import "net/smtp"

type SMTPConfig struct {
	Address           string
	User              string
	Auth              smtp.Auth
	EmailTemplatePath string
}

type EmailMessage struct {
	To       string
	Subject  string
	HTMLBody string
}

// type InviteEmail struct {
// 	Subject         string
// 	FirstName       string
// 	RoleLevelString string
// 	MagicLink       string
// }

type PendingLoginEmailTemplate struct {
	Subject   string
	FirstName string
	MagicLink string
}
