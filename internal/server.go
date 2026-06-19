package internal

import (
	"keyclub-api/auth/handlers"
	"log"
	"net/http"
)

func (a *App) Start(addr string) error {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/auth/login/start", handlers.LoginStartHandler(a.DB, a.Config.PendingLoginExpiryDuration, a.Config.SMTPConfig, a.Config.FrontendURL))
	mux.HandleFunc("GET /api/auth/login/wait", handlers.LoginWaitHandler(a.DB, a.Config.LoginWaitTimeout, a.Config.SessionDuration))
	mux.HandleFunc("POST /api/auth/login/verify", handlers.LoginVerifyHandler(a.DB))
	mux.HandleFunc("GET /api/auth/logout", handlers.LogoutHandler(a.DB))

	log.Printf("listening on %s", addr)
	return http.ListenAndServe(addr, mux)
}
