package app

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	SMTPHost     string
	SMTPPort     string
	SMTPAddress  string
	SMTPUser     string
	SMTPPassword string
}

func LoadConfig() *Config {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	return &Config{
		SMTPHost:     "smtp.gmail.com",
		SMTPPort:     "587",
		SMTPAddress:  "smtp.gmail.com:587",
		SMTPUser:     "jhskeyclub21@gmail.com",
		SMTPPassword: os.Getenv("SMTP_PASSWORD"),
	}
}
