package handlers

import (
	"encoding/json"
	"keyclub-api/auth"
	"keyclub-api/web"
	"log/slog"
	"net/http"
)

type inviteCreateRequest struct {
	Email     string `json:"email"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Role      string `json:"role"`
}

func InviteCreateHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := auth.UserFromContext(r.Context())
		if !ok || user.Role != "officer" {
			web.WriteJSON(w, http.StatusOK, loginStartResponse{Message: "Unauthorized."})
			slog.Info("auth.invite: user is unauthorized")
			return
		}

		var req inviteCreateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			web.WriteJSON(w, http.StatusBadRequest, errorResponse{Error: "Invalid json."})
			slog.Error("auth.invite: decode json failed", "error", err)
			return
		}

	}
}
