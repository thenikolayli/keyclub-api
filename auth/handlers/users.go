package handlers

import (
	"encoding/json"
	"keyclub-api/auth"
	"keyclub-api/web"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/jmoiron/sqlx"
)

type updateUserRequest struct {
	Email     string `json:"email"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Role      string `json:"role"`
}

type updateUserResponse struct {
	Message string `json:"message"`
}

type deleteUserResponse struct {
	Message string `json:"message"`
}

func ListUsersHandler(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := auth.UserFromContext(r.Context())
		if !ok || user.Role != "officer" {
			web.WriteJSON(w, http.StatusUnauthorized, errorResponse{Message: "Unauthorized."})
			slog.Info("auth.users: user is unauthorized")
			return
		}

		skip, err := strconv.Atoi(r.URL.Query().Get("skip"))
		if err != nil || skip < 0 {
			web.WriteJSON(w, http.StatusBadRequest, errorResponse{Message: "Invalid skip parameter."})
			slog.Error("auth.invite: invalid skip parameter", "error", err)
			return
		}
		limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
		if err != nil || limit <= 0 {
			web.WriteJSON(w, http.StatusBadRequest, errorResponse{Message: "Invalid limit parameter."})
			slog.Error("auth.invite: invalid limit parameter", "error", err)
			return
		}
		
		users, err := auth.ListUsers(r.Context(), db, skip, limit)
		if err != nil {
			web.WriteJSON(w, http.StatusInternalServerError, errorResponse{Message: "Failed to list users."})
			slog.Error("auth.users: list users failed", "error", err)
			return
		}

		web.WriteJSON(w, http.StatusOK, users)	
		slog.Info("auth.users: users listed", "count", len(users))
	}
}

func UpdateUserHandler(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := auth.UserFromContext(r.Context())
		if !ok || user.Role != "officer" {
			web.WriteJSON(w, http.StatusUnauthorized, errorResponse{Message: "Unauthorized."})
			slog.Info("auth.users: user is unauthorized")
			return
		}

		var req updateUserRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			web.WriteJSON(w, http.StatusBadRequest, errorResponse{Message: "Invalid json."})
			slog.Error("auth.users: decode json failed", "error", err)
			return
		}

		id := r.PathValue("id")
		if err := auth.UpdateUser(r.Context(), db, auth.User{
			ID:        id,
			Email:     req.Email,
			FirstName: req.FirstName,
			LastName:  req.LastName,
			Role:      req.Role,
		}); err != nil {
			web.WriteJSON(w, http.StatusInternalServerError, errorResponse{Message: "Failed to update user."})
			slog.Error("auth.users: update user failed", "error", err)
			return
		}

		web.WriteJSON(w, http.StatusOK, updateUserResponse{Message: "User updated successfully."})
		slog.Info("auth.users: user updated", "id", id)
	}
}

func DeleteUserHandler(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := auth.UserFromContext(r.Context())
		if !ok || user.Role != "officer" {
			web.WriteJSON(w, http.StatusUnauthorized, errorResponse{Message: "Unauthorized."})
			slog.Info("auth.users: user is unauthorized")
			return
		}

		id := r.PathValue("id")
		if err := auth.DeleteUser(r.Context(), db, id); err != nil {
			web.WriteJSON(w, http.StatusInternalServerError, errorResponse{Message: "Failed to delete user."})
			slog.Error("auth.users: delete user failed", "error", err)
			return
		}

		web.WriteJSON(w, http.StatusOK, deleteUserResponse{Message: "User deleted successfully."})
		slog.Info("auth.users: user deleted", "id", id)
	}
}
