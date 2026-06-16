package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"keyclub-api/auth"
	"keyclub-api/email"
	"keyclub-api/web"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
)

type verifyLoginRequest struct {
	VerifyToken string `json:"verify_token"`
}

type verifyLoginResponse struct {
	Message string `json:"message"`
}

type errorResponse struct {
	Error string `json:"error"`
}

type loginStartRequest struct {
	Email string `json:"email"`
}

type loginStartResponse struct {
	Message string `json:"message"`
}

type loginWaitResponse struct {
	Message string `json:"message"`
}

// New login attempt: creates a pending login, sends magic link email, and returns a cookie with the ID
// Existing login attempt: verifies the attempt ID and email correspond to an existing unexpired uncompleted pending login, if yes: does literally nothing, otherwise: does New Login Attempt
func LoginStartHandler(db *sqlx.DB, pendingLoginExpiry time.Duration, smtpConfig email.SMTPConfig, frontendURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req loginStartRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			web.WriteJSON(w, 400, errorResponse{Error: "invalid json"})
			slog.Error("auth.login_start: decode json failed", "error", err)
			return
		}
		req.Email = strings.TrimSpace(strings.ToLower(req.Email))
		attemptIDCookie, err := r.Cookie("attempt_id")
		if err == nil {
			isNew, err := auth.IsNewLoginAttempt(r.Context(), req.Email, attemptIDCookie.Value, db)
			if err != nil {
				web.WriteJSON(w, 500, errorResponse{Error: "Internal server error, contact the Webmaster."})
				slog.Error("auth.login_start: check existing attempt failed", "error", err, "email", req.Email, "attempt_id", attemptIDCookie.Value)
				return
			}
			if !isNew {
				web.WriteJSON(w, 202, loginStartResponse{Message: "If an account exists with this email, a magic link email will be sent."})
				slog.Info("auth.login_start: attempt already exists", "email", req.Email, "attempt_id", attemptIDCookie.Value)
				return
			}
		} else if errors.Is(err, http.ErrNoCookie) {
		} else {
			web.WriteJSON(w, 500, errorResponse{Error: "Internal server error, contact the Webmaster."})
			slog.Error("auth.login_start: read attempt_id cookie failed", "error", err)
			return
		}

		user, err := auth.GetUserByEmail(r.Context(), req.Email, db)
		if errors.Is(err, auth.UserNotFoundError) {
			newAttemptIDCookie := http.Cookie{
				Name:     "attempt_id",
				Value:    auth.MustGenerateToken(),
				Path:     "/",
				MaxAge:   int(pendingLoginExpiry.Seconds()),
				HttpOnly: true,
				Secure:   true,
				SameSite: http.SameSiteLaxMode,
			}
			http.SetCookie(w, &newAttemptIDCookie)
			web.WriteJSON(w, 202, loginStartResponse{Message: "If an account exists with this email, a magic link email will be sent."})
			slog.Info("auth.login_start: user not found (generic success)", "email", req.Email)
			return
		}
		if err != nil {
			web.WriteJSON(w, 500, errorResponse{Error: "Internal server error, contact the Webmaster."})
			slog.Error("auth.login_start: get user by email failed", "error", err, "email", req.Email)
			return
		}

		id, verifyToken, err := auth.CreatePendingLogin(r.Context(), user, db, pendingLoginExpiry)
		if err != nil {
			web.WriteJSON(w, 500, errorResponse{Error: "Internal server error, contact the Webmaster."})
			slog.Error("auth.login_start: create pending login failed", "error", err, "email", req.Email, "user_id", user.ID)
			return
		}

		emailTemplate := email.PendingLoginEmailTemplate{
			Subject:   "Attempted Login",
			FirstName: user.FirstName,
			MagicLink: fmt.Sprintf("%s/admin/verifylogin?verify_token=%s", frontendURL, verifyToken),
		}
		if err := email.SendPendingLoginEmail(emailTemplate, req.Email, smtpConfig); err != nil {
			web.WriteJSON(w, 500, errorResponse{Error: "Internal server error, contact the Webmaster."})
			slog.Error("auth.login_start: send magic link email failed", "error", err, "email", req.Email, "user_id", user.ID, "attempt_id", id)
			return
		}

		newAttemptIDCookie := http.Cookie{
			Name:     "attempt_id",
			Value:    id,
			Path:     "/",
			MaxAge:   int(pendingLoginExpiry.Seconds()),
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteLaxMode,
		}
		http.SetCookie(w, &newAttemptIDCookie)
		web.WriteJSON(w, 202, loginStartResponse{Message: "If an account exists with this email, a magic link email will be sent."})
		slog.Info("auth.login_start: started", "email", req.Email, "user_id", user.ID, "attempt_id", id)
	}
}

// Confirms that an attempt id cookies exists, awaits the user to verify the login via the magic link email
// If the cookie doesn't exist, creates a dummy one that will be deleted at the end
// Deletes cookie at the end
func LoginWaitHandler(db *sqlx.DB, loginWaitTimeout time.Duration) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		attemptIDCookie, err := r.Cookie("attempt_id")
		if err == nil {
		} else if errors.Is(err, http.ErrNoCookie) {
			web.WriteJSON(w, 404, errorResponse{Error: "Attempt ID cookie not found."})
			slog.Info("auth.login_wait: attempt_id cookie missing")
			return
		} else {
			web.WriteJSON(w, 500, errorResponse{Error: "Internal server error, contact the Webmaster."})
			slog.Error("auth.login_wait: read attempt_id cookie failed", "error", err)
			return
		}

		wait, unregister := auth.RegisterLoginWaiter(attemptIDCookie.Value)
		defer unregister()

		if !waitForLoginVerified(r.Context(), wait, loginWaitTimeout) {
			web.WriteJSON(w, 408, errorResponse{Error: "login timed out waiting for email confirmation"})
			slog.Info("auth.login_wait: timed out", "attempt_id", attemptIDCookie.Value)
			return
		}

		// Create session

		deleteAttemptIDCookie := http.Cookie{
			Name:     "attempt_id",
			Value:    "",
			Path:     "/",
			MaxAge:   -1,
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteLaxMode,
		}
		http.SetCookie(w, &deleteAttemptIDCookie)
		web.WriteJSON(w, 200, loginWaitResponse{Message: "Login confirmed. You can return to the login page."})
		slog.Info("auth.login_wait: verified", "attempt_id", attemptIDCookie.Value)
	}
}

// Verifies a pending login once a user clicks the magic link in the email
func LoginVerifyHandler(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req verifyLoginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			web.WriteJSON(w, 400, errorResponse{Error: "Invalid json."})
			slog.Error("auth.login_verify: decode json failed", "error", err)
			return
		}

		attemptID, err := auth.VerifyPendingLogin(r.Context(), req.VerifyToken, db)
		if errors.Is(err, auth.PendingLoginNotFoundError) ||
			errors.Is(err, auth.PendingLoginExpiredError) ||
			errors.Is(err, auth.PendingLoginAlreadyUsedError) {
			web.WriteJSON(w, 400, errorResponse{Error: "Invalid token or expired."})
			slog.Info("auth.login_verify: invalid/expired token", "error", err)
			return
		}
		if err != nil {
			web.WriteJSON(w, 500, errorResponse{Error: "Internal server error, contact the Webmaster."})
			slog.Error("auth.login_verify: verify pending login failed", "error", err)
			return
		}

		auth.NotifyLoginWaiter(attemptID)

		web.WriteJSON(w, 200, verifyLoginResponse{Message: "Login confirmed. You can return to the login page."})
		slog.Info("auth.login_verify: verified", "attempt_id", attemptID)
	}
}

// Waits for a login waiter to be notified or the timeout expires
func waitForLoginVerified(ctx context.Context, wait <-chan struct{}, timeout time.Duration) bool {
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	select {
	case <-wait:
		return true
	case <-timer.C:
		return false
	case <-ctx.Done():
		return false
	}
}
