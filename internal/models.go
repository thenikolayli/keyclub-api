package internal

import "net/smtp"

type SMTPConfig struct {
	Host     string
	Port     string
	Address  string
	User     string
	Password string
	Auth     smtp.Auth
}
