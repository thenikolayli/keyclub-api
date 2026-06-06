package internal

import (
	"keyclub-api/auth/handlers"
	"log"
	"net/http"
)

func (a *App) Start(addr string) error {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/auth/login", handlers.Login(a.DB, a.Config.PendingLoginExpiryDuration, a.Config.SMTPConfig))

	log.Printf("listening on %s", addr)
	return http.ListenAndServe(addr, mux)
}
