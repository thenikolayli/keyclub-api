package internal

import (
	"context"
	"keyclub-api/auth/handlers"
	"keyclub-api/sync"
	"log"
	"net/http"
)

func (a *App) Start(addr string) error {
	go sync.SyncMembersFromSheet(context.Background(), a.GoogleConfig, &a.MemberSync, a.DB)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /auth/login/start", handlers.LoginStartHandler(a.DB, a.Config.Durations.PendingLoginExpiryDuration, a.Config.SMTPConfig, a.Config.FrontendURL))
	mux.HandleFunc("GET /auth/login/wait", handlers.LoginWaitHandler(a.DB, a.Config.Durations.LoginWaitTimeout, a.Config.Durations.SessionDuration))
	mux.HandleFunc("POST /auth/login/verify", handlers.LoginVerifyHandler(a.DB))
	mux.HandleFunc("GET /auth/logout", handlers.LogoutHandler(a.DB))

	log.Printf("listening on %s", addr)
	return http.ListenAndServe(addr, mux)
}
