package internal

import (
	"keyclub-api/auth/handlers"
	"log"
	"net/http"
)

func (a *App) Start(addr string) error {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/auth/login", handlers.LoginHandler(a.DB, a.Config.PendingLoginExpiryDuration, a.Config.SMTPConfig, a.Config.FrontendURL))
	mux.HandleFunc("POST /api/auth/verifylogin", handlers.VerifyLoginHandler(a.DB))

	log.Printf("listening on %s", addr)
	return http.ListenAndServe(addr, mux)
}
