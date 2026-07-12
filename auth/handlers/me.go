package handlers

import (
	"keyclub-api/auth"
	"keyclub-api/web"
	"log/slog"
	"net/http"

	"github.com/jmoiron/sqlx"
)

type meResponse struct {
	Email     string `json:"email"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Role      string `json:"role"`
}

func MeHandler(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := auth.UserFromContext(r.Context())
		if !ok {
			web.WriteJSON(w, http.StatusUnauthorized, errorResponse{Error: "Unauthorized."})
			slog.Error("auth.me: user not found in context")
			return
		}
		web.WriteJSON(w, http.StatusOK, meResponse{
			Email:     user.Email,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Role:      user.Role,
		})
	}
}
