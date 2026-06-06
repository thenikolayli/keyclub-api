package internal

import (
	"log"

	"github.com/jmoiron/sqlx"
)

type App struct {
	Config Config
	DB     *sqlx.DB
}

func NewApp() App {
	config := LoadConfig()
	db, err := LoadDatabase(config.DBConfig)
	if err != nil {
		log.Fatalf("Failed to load database: %v", err)
	}

	return App{Config: config, DB: db}
}
