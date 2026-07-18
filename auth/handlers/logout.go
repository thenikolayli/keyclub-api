package handlers

import (
	"keyclub-api/auth"
	"keyclub-api/config"
	"keyclub-api/web"
	"log/slog"
	"net/http"
	"time"

	"github.com/jmoiron/sqlx"
)

type logoutResponse struct {
	Message string `json:"message"`
}

// Logs the user out by revoking their session and deleting the session cookie
func LogoutHandler(db *sqlx.DB, cookieCfg config.CookieConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sessionCookie, err := r.Cookie("session")
		if err == nil {
			err := auth.RevokeSessionBySessionToken(r.Context(), sessionCookie.Value, db)
			if err != nil {
				web.WriteJSON(w, http.StatusInternalServerError, errorResponse{Message: "Internal server error, contact the Webmaster."})
				slog.Error("auth.logout: revoke session by session token failed", "error", err)
				return
			}
		}
		http.SetCookie(w, &http.Cookie{
			Name:     "session",
			Value:    "",
			Path:     cookieCfg.Path,
			Domain:   cookieCfg.Domain,
			MaxAge:   -1,
			Expires:  time.Unix(0, 0),
			HttpOnly: cookieCfg.HttpOnly,
			Secure:   cookieCfg.Secure,
			SameSite: http.SameSiteLaxMode,
		})
		web.WriteJSON(w, http.StatusOK, logoutResponse{Message: "Logged out successfully."})
		slog.Info("auth.logout: logged out")
	}
}
