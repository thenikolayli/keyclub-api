package auth

import (
	"context"
	"errors"
	"keyclub-api/web"
	"log/slog"
	"net/http"
	"time"

	"github.com/jmoiron/sqlx"
)

type contextKey struct{}

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
func SessionMiddleware(db *sqlx.DB, next http.Handler, sessionDuration time.Duration) http.Handler {
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
				sessionCookie.MaxAge = 0
				http.SetCookie(w, sessionCookie)
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
			r = r.WithContext(context.WithValue(r.Context(), contextKey{}, user))
			next.ServeHTTP(w, r)
			return
		} else if errors.Is(err, http.ErrNoCookie) {
			// Do nothing if there's no session cookie
		} else {
			web.WriteJSON(w, http.StatusInternalServerError, errorResponse{Error: "Internal server error, contact the Webmaster."})
			slog.Error("auth.middleware: read session cookie failed", "error", err)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// Safe getter function for the user from request context
// Using contextKey as the key to avoid potential collisions with other context keys
func UserFromContext(ctx context.Context) (User, bool) {
	user, ok := ctx.Value(contextKey{}).(User)
	return user, ok
}
