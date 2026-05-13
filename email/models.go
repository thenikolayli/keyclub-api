package email

type EmailMessage struct {
	To       string
	Subject  string
	HTMLBody string
}

type InviteEmail struct {
	Subject         string
	FirstName       string
	RoleLevelString string
	MagicLink       string
}
