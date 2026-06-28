package handlers

import (
	"encoding/json"
	"keyclub-api/members"
	"keyclub-api/web"
	"log/slog"
	"net/http"

	"github.com/jmoiron/sqlx"
)

type hoursRequest struct {
	Name string `json:"name"`
}

type hoursResponse struct {
	Name      string  `json:"name"`
	AllHours  float64 `json:"all_hours"`
	TermHours float64 `json:"term_hours"`
	GradYear  int     `json:"grad_year"`
	Class     string  `json:"class"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func HoursHandler(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req hoursRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			web.WriteJSON(w, http.StatusInternalServerError, errorResponse{Error: "Internal server error, contact the Webmaster."})
			slog.Error("members.hours: decode request failed", "error", err)
			return
		}

		member, err := members.GetMember(r.Context(), db, req.Name)
		if err != nil {
			web.WriteJSON(w, http.StatusNotFound, errorResponse{Error: "Member not found."})
			slog.Error("members.hours: get member failed", "error", err, "name", req.Name)
			return
		}

		formattedMember := member.Format()
		web.WriteJSON(w, http.StatusOK, hoursResponse{
			Name:      formattedMember.Name,
			AllHours:  formattedMember.AllHours,
			TermHours: formattedMember.TermHours,
			GradYear:  formattedMember.GradYear,
			Class:     formattedMember.Class,
		})
	}
}
