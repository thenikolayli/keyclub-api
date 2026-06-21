package internal

import (
	"context"
	"fmt"
	"keyclub-api/auth/handlers"
	"keyclub-api/sync"
	"log"
	"log/slog"
	"net/http"
)

func (a *App) Start(addr string) error {
	if err := sync.SyncMembers(context.Background(), a.GoogleConfig, &a.MemberSync, a.DB); err != nil {
		slog.Error("server: failed to sync members", "error", err)
		return fmt.Errorf("failed to sync members: %v", err)
	}

	if err := sync.SyncEvents(context.Background(), a.GoogleConfig, &a.EventSync, a.DB); err != nil {
		slog.Error("server: failed to sync events", "error", err)
		return fmt.Errorf("failed to sync events: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /auth/login/start", handlers.LoginStartHandler(a.DB, a.Config.Durations.PendingLoginExpiryDuration, a.Config.SMTPConfig, a.Config.FrontendURL))
	mux.HandleFunc("GET /auth/login/wait", handlers.LoginWaitHandler(a.DB, a.Config.Durations.LoginWaitTimeout, a.Config.Durations.SessionDuration))
	mux.HandleFunc("POST /auth/login/verify", handlers.LoginVerifyHandler(a.DB))
	mux.HandleFunc("GET /auth/logout", handlers.LogoutHandler(a.DB))

	log.Printf("listening on %s", addr)
	return http.ListenAndServe(addr, mux)
}
