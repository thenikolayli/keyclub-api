package internal

import (
	"context"
	"errors"
	"fmt"
	"keyclub-api/google"
	"keyclub-api/sync"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jmoiron/sqlx"

	authHandlers "keyclub-api/auth/handlers"
	membersHandlers "keyclub-api/members/handlers"
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
	mux.HandleFunc("POST /auth/login/start", authHandlers.LoginStartHandler(a.DB, a.Config.Durations.PendingLoginExpiryDuration, a.Config.SMTPConfig, a.Config.FrontendURL))
	mux.HandleFunc("GET /auth/login/wait", authHandlers.LoginWaitHandler(a.DB, a.Config.Durations.LoginWaitTimeout, a.Config.Durations.SessionDuration))
	mux.HandleFunc("POST /auth/login/verify", authHandlers.LoginVerifyHandler(a.DB))
	mux.HandleFunc("GET /auth/logout", authHandlers.LogoutHandler(a.DB))
	mux.HandleFunc("POST /members/hours", membersHandlers.HoursHandler(a.DB))

	server := &http.Server{Addr: addr, Handler: cors(a.Config.FrontendURL, mux)}

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

// CORS config, so the browser can accept responses from the server
func cors(allowedOrigin string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
