package handlers

import (
	"keyclub-api/auth"
	"keyclub-api/web"
	"log/slog"
	"net/http"

	"github.com/jmoiron/sqlx"
)

type logoutResponse struct {
	Message string `json:"message"`
}

// Logs the user out by revoking their session and deleting the session cookie
func LogoutHandler(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sessionCookie, err := r.Cookie("session")
		if err == nil {
			err := auth.RevokeSessionBySessionToken(r.Context(), sessionCookie.Value, db)
			if err != nil {
				web.WriteJSON(w, 500, errorResponse{Error: "Internal server error, contact the Webmaster."})
				slog.Error("auth.logout: revoke session by session token failed", "error", err)
				return
			}
		}
		sessionCookie.MaxAge = 0
		http.SetCookie(w, sessionCookie)
		web.WriteJSON(w, 200, logoutResponse{Message: "Logged out successfully."})
		slog.Info("auth.logout: logged out")
	}
}
