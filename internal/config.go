package internal

import (
	"log"
	"net/smtp"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	SMTPConfig SMTPConfig

	InviteExpiryDuration time.Duration
	EmailTemplatePath    string
}

func LoadConfig() Config {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	return Config{
		SMTPConfig: SMTPConfig{
			Address: "smtp.gmail.com:587",
			User:    "jhskeyclub21@gmail.com",
			Auth:    smtp.PlainAuth("", "jhskeyclub21@gmail.com", os.Getenv("SMTP_PASSWORD"), "smtp.gmail.com"),
		},

		InviteExpiryDuration: 24 * time.Hour,
		EmailTemplatePath:    os.Getenv("EMAIL_TEMPLATE_PATH"),
	}
}
