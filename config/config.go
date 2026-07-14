package config

import (
	"keyclub-api/email"
	"time"
)

type Durations struct {
	PendingLoginExpiryDuration time.Duration
	LoginWaitTimeout           time.Duration
	InviteExpiryDuration       time.Duration
	SessionDuration            time.Duration
	MemberSyncTimeout          time.Duration
	EventSyncTimeout           time.Duration
}

type CookieConfig struct {
	Path     string
	Domain   string
	Secure   bool
	HttpOnly bool
}

type Config struct {
	SMTPConfig   email.SMTPConfig
	DBConfig     DBConfig
	CookieConfig CookieConfig
	Durations    Durations

	FrontendURL string
	APIURL      string
	Officers    []string
}

type DBConfig struct {
	SQLitePath     string
	MigrationsPath string
}
