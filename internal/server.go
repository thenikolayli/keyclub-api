package internal

import (
	"context"
	"fmt"
	authHandlers "keyclub-api/auth/handlers"
	membersHandlers "keyclub-api/members/handlers"
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
	mux.HandleFunc("POST /auth/login/start", authHandlers.LoginStartHandler(a.DB, a.Config.Durations.PendingLoginExpiryDuration, a.Config.SMTPConfig, a.Config.FrontendURL))
	mux.HandleFunc("GET /auth/login/wait", authHandlers.LoginWaitHandler(a.DB, a.Config.Durations.LoginWaitTimeout, a.Config.Durations.SessionDuration))
	mux.HandleFunc("POST /auth/login/verify", authHandlers.LoginVerifyHandler(a.DB))
	mux.HandleFunc("GET /auth/logout", authHandlers.LogoutHandler(a.DB))
	mux.HandleFunc("POST /members/hours", membersHandlers.HoursHandler(a.DB))

	log.Printf("listening on %s", addr)
	return http.ListenAndServe(addr, cors(a.Config.FrontendURL, mux))
}

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
