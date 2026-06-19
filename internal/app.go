package internal

import (
	"context"
	"keyclub-api/google"
	"keyclub-api/sync"
	"log"
	"os"
	"time"

	"github.com/jmoiron/sqlx"
)

type App struct {
	Config       Config
	DB           *sqlx.DB
	GoogleConfig google.GoogleConfig
	MemberSync   sync.SyncState
	EventSync    sync.SyncState
}

func NewApp() App {
	config := LoadConfig()
	db, err := LoadDatabase(config.DBConfig)
	if err != nil {
		log.Fatalf("Failed to load database: %v", err)
	}
	googleConfig, err := google.LoadGoogleServices(context.Background(), os.Getenv("GOOGLE_KEY_FILE_PATH"))
	if err != nil {
		log.Fatalf("Failed to load google services: %v", err)
	}

	return App{
		Config:       config,
		DB:           db,
		GoogleConfig: googleConfig,
		MemberSync:   sync.SyncState{LastUpdated: time.Now(), UpdateTimeout: config.Durations.MemberSyncTimeout},
		EventSync:    sync.SyncState{LastUpdated: time.Now(), UpdateTimeout: config.Durations.EventSyncTimeout},
	}
}
