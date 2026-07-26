package auth

import (
	"context"
	"errors"
	"keyclub-api/config"
	"keyclub-api/web"
	"log/slog"
	"net/http"
	"time"

	"github.com/jmoiron/sqlx"
)

// CORS middleware, so the browser can accept responses from the server
func CORSMiddleware(allowedOrigin string, next http.Handler) http.Handler {
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

// Intercepts every API request and checks if the user is authenticated
// Stores user info in the request context
func SessionMiddleware(db *sqlx.DB, next http.Handler, sessionDuration time.Duration, cookieCfg config.CookieConfig) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sessionCookie, err := r.Cookie("session")
		if err == nil {
			session, err := GetSessionByToken(r.Context(), sessionCookie.Value, db)
			if err != nil {
				web.WriteJSON(w, http.StatusInternalServerError, errorResponse{Error: "Internal server error, contact the Webmaster."})
				slog.Error("auth.middleware: get session by token failed", "error", err)
				return
			}

			valid, err := IsValidSession(r.Context(), session, db)
			if err != nil {
				web.WriteJSON(w, http.StatusInternalServerError, errorResponse{Error: "Internal server error, contact the Webmaster."})
				slog.Error("auth.middleware: is valid session failed", "error", err)
				return
			}
			if !valid {
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
				next.ServeHTTP(w, r)
				return
			}

			// If less then half the session duration left until expiry, extend the session
			if time.Until(session.ExpiresAt) < sessionDuration/2 {
				_, err := db.ExecContext(r.Context(), "UPDATE sessions SET expires_at = ? WHERE id = ?", time.Now().Add(sessionDuration), session.ID)
				if err != nil {
					web.WriteJSON(w, http.StatusInternalServerError, errorResponse{Error: "Internal server error, contact the Webmaster."})
					slog.Error("auth.middleware: extend session failed", "error", err)
					return
				}
				sessionCookie.Expires = time.Now().Add(sessionDuration)
				http.SetCookie(w, sessionCookie)
			}

			user, err := GetUserByID(r.Context(), session.UserID, db)
			if err != nil {
				web.WriteJSON(w, http.StatusInternalServerError, errorResponse{Error: "Internal server error, contact the Webmaster."})
				slog.Error("auth.middleware: get user by id failed", "error", err)
				return
			}
			r = r.WithContext(context.WithValue(r.Context(), "session_user", user))
			next.ServeHTTP(w, r)
			return
		} else if errors.Is(err, http.ErrNoCookie) {
			// Do nothing if there's no session cookie
			slog.Info("auth.middleware: no session cookie found, proceeding without user context")
		} else {
			// If some other error occurred
			web.WriteJSON(w, http.StatusInternalServerError, errorResponse{Error: "Internal server error, contact the Webmaster."})
			slog.Error("auth.middleware: read session cookie failed", "error", err)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// Safe getter function for the user from request context
func UserFromContext(ctx context.Context) (User, bool) {
	user, ok := ctx.Value("session_user").(User)
	return user, ok
}
