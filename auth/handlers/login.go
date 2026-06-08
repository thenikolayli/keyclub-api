package handlers

import (
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

type loginRequest struct {
	Email string `json:"email"`
}

type loginResponse struct {
	Message string `json:"message"`
	LoginID string `json:"login_id"`
}

type errorResponse struct {
	Error string `json:"error"`
}

type verifyLoginRequest struct {
	Token string `json:"token"`
}

type verifyLoginResponse struct {
	Message string `json:"message"`
}

// needs to check if the user exists, if they do, send them an email
func LoginHandler(db *sqlx.DB, expiry time.Duration, smtpConfig email.SMTPConfig, frontendURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req loginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			web.WriteJSON(w, 400, errorResponse{Error: "invalid json"})
			slog.Error("Login Failed at getting request body: Invalid json.", "error", err)
			return
		}
		req.Email = strings.TrimSpace(strings.ToLower(req.Email))

		user, err := auth.GetUserByEmail(r.Context(), req.Email, db)
		if errors.Is(err, auth.UserNotFoundError) {
			web.WriteJSON(w, 202, loginResponse{Message: "If a user exists, they will be sent an email.", LoginID: auth.MustGenerateToken()})
			slog.Info("Login Failed at getting user by email: User not found.", "email", req.Email)
			return
		}
		if err != nil {
			web.WriteJSON(w, 500, errorResponse{Error: "Internal server error, contact the Webmaster."})
			slog.Error("Login Failed at getting user by email: Internal server error.", "error", err)
			return
		}

		loginID, verifyToken, err := auth.CreatePendingLogin(r.Context(), user, db, expiry)
		if err != nil {
			web.WriteJSON(w, 500, errorResponse{Error: "Internal server error, contact the Webmaster."})
			slog.Error("Login Failed at creating pending login: Internal server error.", "error", err)
			return
		}

		emailTemplate := email.PendingLoginEmailTemplate{
			Subject:   "Attempted Login",
			FirstName: user.FirstName,
			MagicLink: fmt.Sprintf("%s/admin/verifylogin?token=%s", frontendURL, verifyToken),
		}
		if err := email.SendPendingLoginEmail(emailTemplate, req.Email, smtpConfig); err != nil {
			web.WriteJSON(w, 500, errorResponse{Error: "Internal server error, contact the Webmaster."})
			slog.Error("Login Failed at sending pending login email: Internal server error.", "error", err)
			return
		}

		web.WriteJSON(w, 202, loginResponse{Message: "If a user exists, they will be sent an email.", LoginID: loginID})
		slog.Info("Login Successful at creating pending login.", "email", req.Email, "login_id", loginID)
	}
}

func VerifyLoginHandler(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req verifyLoginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			web.WriteJSON(w, 400, errorResponse{Error: "Invalid json."})
			slog.Error("Verify Login Failed at getting request body: Invalid json.", "error", err)
			return
		}
		token := req.Token

		verified, err := auth.VerifyPendingLogin(r.Context(), token, db)
		if err != nil {
			web.WriteJSON(w, 400, errorResponse{Error: "Internal server error, contact the Webmaster."})
			slog.Error("Verify Login Failed at verifying token: Internal server error.", "error", err)
			return
		}
		if !verified {
			web.WriteJSON(w, 400, errorResponse{Error: "Invalid token or expired."})
			slog.Error("Verify Login Failed at verifying token: Invalid token or expired.")
			return
		}

		web.WriteJSON(w, 200, verifyLoginResponse{Message: "Login successful."})
	}
}
