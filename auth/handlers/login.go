package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"keyclub-api/auth"
	"keyclub-api/email"
	"keyclub-api/web"
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

// needs to check if the user exists, if they do, send them an email
func Login(db *sqlx.DB, expiry time.Duration, smtpConfig email.SMTPConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req loginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			web.WriteJSON(w, 400, errorResponse{Error: "invalid json"})
			return
		}
		req.Email = strings.TrimSpace(strings.ToLower(req.Email))

		user, err := auth.GetUserByEmail(r.Context(), req.Email, db)
		if errors.Is(err, auth.UserNotFoundError) {
			web.WriteJSON(w, 202, loginResponse{Message: "If a user exists, they will be sent an email.", LoginID: auth.MustGenerateToken()})
			return
		}
		if err != nil {
			web.WriteJSON(w, 500, errorResponse{Error: "Internal server error, contact the Webmaster."})
			return
		}

		loginID, verifyToken, err := auth.CreatePendingLogin(r.Context(), user, db, expiry)
		if err != nil {
			web.WriteJSON(w, 500, errorResponse{Error: "Internal server error, contact the Webmaster."})
			return
		}

		emailTemplate := email.PendingLoginEmailTemplate{
			Subject:   "Attempted Login",
			FirstName: user.FirstName,
			MagicLink: fmt.Sprintf("https://jhskeyclub.com/api/auth/verifylogin?token=%s", verifyToken),
		}
		if err := email.SendPendingLoginEmail(emailTemplate, req.Email, smtpConfig); err != nil {
			web.WriteJSON(w, 500, errorResponse{Error: "Internal server error, contact the Webmaster."})
			return
		}

		web.WriteJSON(w, 202, loginResponse{Message: "If a user exists, they will be sent an email.", LoginID: loginID})
	}
}

// func VerifyLogin(db *sqlx.DB) http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		token := r.URL.Query().Get("token")
// 		if token == "" {
// 			web.WriteJSON(w, 400, errorResponse{Error: "token is required"})
// 			return
// 		}

// 	}
// }
