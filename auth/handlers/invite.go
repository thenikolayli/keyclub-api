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

type inviteCreateRequest struct {
	Email     string `json:"email"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Role      string `json:"role"`
}

type inviteCreateResponse struct {
	Message string `json:"message"`
}

type inviteAcceptRequest struct {
	Token string `json:"token"`
}

type inviteAcceptResponse struct {
	Message string `json:"message"`
}

func InviteCreateHandler(db *sqlx.DB, inviteDuration time.Duration, frontendURL string, smtpConfig email.SMTPConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := auth.UserFromContext(r.Context())
		if !ok || user.Role != "officer" {
			web.WriteJSON(w, http.StatusOK, loginStartResponse{Message: "Unauthorized."})
			slog.Info("auth.invite: user is unauthorized")
			return
		}

		var req inviteCreateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			web.WriteJSON(w, http.StatusBadRequest, errorResponse{Message: "Invalid json."})
			slog.Error("auth.invite: decode json failed", "error", err)
			return
		}
		if err := req.Validate(db); err != nil {
			web.WriteJSON(w, http.StatusBadRequest, errorResponse{Message: err.Error()})
			slog.Info("auth.invite: validation failed", "error", err)
			return
		}

		token, err := auth.CreateInvite(r.Context(), db, inviteDuration, req.Email, req.FirstName, req.LastName, req.Role)
		if err != nil {
			web.WriteJSON(w, http.StatusInternalServerError, errorResponse{Message: "Failed to create invite."})
			slog.Error("auth.invite: create invite failed", "error", err)
			return
		}

		emailTemplate := auth.InviteEmailTemplate{
			Subject:   "Key Club Invite",
			FirstName: req.FirstName,
			Role:      req.Role,
			MagicLink: fmt.Sprintf("%s/admin/acceptinvite?token=%s", frontendURL, token),
		}
		if err := auth.SendInviteEmail(emailTemplate, req.Email, smtpConfig); err != nil {
			web.WriteJSON(w, http.StatusInternalServerError, errorResponse{Message: "Internal server error, contact the Webmaster."})
			slog.Error("auth.invite: send invite email failed", "error", err, "email", req.Email)
			return
		}
		web.WriteJSON(w, http.StatusAccepted, inviteCreateResponse{Message: "An invite has been sent to " + req.Email + "."})
		slog.Info("auth.invite: invite created", "email", req.Email)
	}
}

func InviteAcceptHandler(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req inviteAcceptRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			web.WriteJSON(w, http.StatusBadRequest, errorResponse{Message: "Invalid json."})
			slog.Error("auth.invite_accept: decode json failed", "error", err)
			return
		}

		transaction, err := db.BeginTxx(r.Context(), nil)
		if err != nil {
			web.WriteJSON(w, http.StatusInternalServerError, errorResponse{Message: "Internal server error, contact the Webmaster."})
			slog.Error("auth.invite_accept: begin transaction failed", "error", err)
			return
		}
		defer transaction.Rollback()

		err = auth.AcceptInvite(r.Context(), transaction, req.Token)
		if err != nil {
			if err == auth.InviteNotFoundError || err == auth.InviteExpiredError || err == auth.InviteAlreadyUsedError {
				web.WriteJSON(w, http.StatusBadRequest, errorResponse{Message: err.Error()})
				slog.Info("auth.invite_accept: invite accept failed", "error", err)
				return
			}
			web.WriteJSON(w, http.StatusInternalServerError, errorResponse{Message: "Internal server error, contact the Webmaster."})
			slog.Error("auth.invite_accept: invite accept failed", "error", err)
			return
		}
		if err := transaction.Commit(); err != nil {
			web.WriteJSON(w, http.StatusInternalServerError, errorResponse{Message: "Internal server error, contact the Webmaster."})
			slog.Error("auth.invite_accept: commit transaction failed", "error", err)
			return
		}

		web.WriteJSON(w, http.StatusOK, inviteAcceptResponse{Message: "Invite accepted successfully."})
		slog.Info("auth.invite_accept: invite accepted successfully")
	}
}

func (r inviteCreateRequest) Validate(db *sqlx.DB) error {
	r.Email = strings.TrimSpace(r.Email)
	r.FirstName = strings.TrimSpace(r.FirstName)
	r.LastName = strings.TrimSpace(r.LastName)
	r.Role = strings.TrimSpace(r.Role)

	if r.Email == "" || r.FirstName == "" || r.LastName == "" || r.Role == "" {
		return fmt.Errorf("all fields are required")
	}

	_, err := auth.GetUserByEmail(context.Background(), r.Email, db)
	if errors.Is(err, auth.UserNotFoundError) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("Failed to check if user exists: %v", err)
	}
	return fmt.Errorf("User with email %s already exists", r.Email)
}
