package internal

import (
	"context"
	"errors"
	"fmt"
	"keyclub-api/auth"
	"keyclub-api/config"
	"keyclub-api/email"
	"keyclub-api/google"
	"keyclub-api/sync"
	"log"
	"log/slog"
	"net/http"
	"net/smtp"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"

	authHandlers "keyclub-api/auth/handlers"
	eventsHandlers "keyclub-api/events/handlers"
	membersHandlers "keyclub-api/members/handlers"
)

type App struct {
	Config       config.Config
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

func (a *App) Start(addr string) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	scheduler, err := StartScheduler(context.Background(), a.GoogleConfig, &a.MemberSync, &a.EventSync, a.DB)
	if err != nil {
		slog.Error("app: failed to start scheduler", "error", err)
		return fmt.Errorf("failed to start scheduler: %v", err)
	}
	scheduler.Start()

	mux := http.NewServeMux()
	mux.HandleFunc("POST /auth/login/start", authHandlers.LoginStartHandler(a.DB, a.Config.Durations.PendingLoginExpiryDuration, a.Config.SMTPConfig, a.Config.FrontendURL, a.Config.CookieConfig))
	mux.HandleFunc("GET /auth/login/wait", authHandlers.LoginWaitHandler(a.DB, a.Config.Durations.LoginWaitTimeout, a.Config.Durations.SessionDuration, a.Config.CookieConfig))
	mux.HandleFunc("POST /auth/login/verify", authHandlers.LoginVerifyHandler(a.DB))
	mux.HandleFunc("GET /auth/logout", authHandlers.LogoutHandler(a.DB, a.Config.CookieConfig))
	mux.HandleFunc("GET /auth/me", authHandlers.MeHandler(a.DB))

	mux.HandleFunc("POST /members/hours", membersHandlers.HoursHandler(a.DB))

	mux.HandleFunc("POST /events/search", eventsHandlers.SearchHandler(a.DB))

	server := &http.Server{
		Addr:    addr,
		Handler: auth.CORSMiddleware(a.Config.FrontendURL, auth.SessionMiddleware(a.DB, mux, a.Config.Durations.SessionDuration)),
	}

	// Create a channel to receive server errors
	serverErr := make(chan error, 1)
	go func() {
		slog.Info("app: started", "address", addr)
		serverErr <- server.ListenAndServe()
	}()

	// Wait for server errors or context cancellation
	// If the server receives an error it shuts down regardless, so shut down the scheduler as well
	select {
	case err := <-serverErr:
		scheduler.Shutdown()
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return fmt.Errorf("server error: %v", err)
	case <-ctx.Done():
		slog.Info("app: shutting down")
	}

	// Shutdown server and scheduler
	// If the shutdown takes longer than 10 seconds, force shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		slog.Error("app: server shutdown error", "error", err)
	}
	if err := scheduler.Shutdown(); err != nil {
		slog.Error("app: scheduler shutdown error", "error", err)
	}
	return nil
}

func LoadConfig() config.Config {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	return config.Config{
		SMTPConfig: email.SMTPConfig{
			Address:           "smtp.gmail.com:587",
			User:              "jhskeyclub21@gmail.com",
			Auth:              smtp.PlainAuth("", "jhskeyclub21@gmail.com", os.Getenv("SMTP_PASSWORD"), "smtp.gmail.com"),
			EmailTemplatePath: "maizzle/build_production",
		},

		DBConfig: config.DBConfig{
			SQLitePath:     os.Getenv("DB_SQLITE_PATH"),
			MigrationsPath: os.Getenv("DB_MIGRATIONS_PATH"),
		},

		Durations: config.Durations{
			PendingLoginExpiryDuration: 1 * time.Hour,
			LoginWaitTimeout:           5 * time.Minute,
			InviteExpiryDuration:       24 * time.Hour,
			SessionDuration:            14 * 24 * time.Hour,
			MemberSyncTimeout:          1 * time.Hour,
			EventSyncTimeout:           1 * time.Hour,
		},

		CookieConfig: config.CookieConfig{
			Path:     "/",
			Domain:   os.Getenv("COOKIE_DOMAIN"),
			Secure:   os.Getenv("COOKIE_SECURE") == "true",
			HttpOnly: true,
		},

		FrontendURL: os.Getenv("FRONTEND_URL"),
		APIURL:      os.Getenv("API_URL"),
		Officers:    strings.Split(os.Getenv("OFFICERS"), ", "),
	}
}
