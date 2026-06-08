package internal

import (
	"keyclub-api/email"
	"log"
	"net/smtp"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	SMTPConfig email.SMTPConfig
	DBConfig   DBConfig

	PendingLoginExpiryDuration time.Duration
	InviteExpiryDuration       time.Duration

	FrontendURL string
	APIURL      string
}

func LoadConfig() Config {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	return Config{
		SMTPConfig: email.SMTPConfig{
			Address:           "smtp.gmail.com:587",
			User:              "jhskeyclub21@gmail.com",
			Auth:              smtp.PlainAuth("", "jhskeyclub21@gmail.com", os.Getenv("SMTP_PASSWORD"), "smtp.gmail.com"),
			EmailTemplatePath: "maizzle/build_production",
		},

		DBConfig: DBConfig{
			SQLitePath:     os.Getenv("DB_SQLITE_PATH"),
			MigrationsPath: os.Getenv("DB_MIGRATIONS_PATH"),
		},

		PendingLoginExpiryDuration: 1 * time.Hour,
		InviteExpiryDuration:       24 * time.Hour,

		FrontendURL: os.Getenv("FRONTEND_URL"),
		APIURL:      os.Getenv("API_URL"),
	}
}
