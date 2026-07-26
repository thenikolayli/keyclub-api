package handlers

import (
	"encoding/json"
	"errors"
	"keyclub-api/auth"
	"keyclub-api/events"
	"keyclub-api/web"
	"log/slog"
	"net/http"

	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/docs/v1"
)

type addToCalendarRequest struct {
	URL string `json:"url"`
}

type addToCalendarResponse struct {
	URL string `json:"url"`
}

func AddToCalendarHandler(calendarID string, docsService *docs.Service, calendarService *calendar.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := auth.UserFromContext(r.Context())
		if !ok || user.Role != "officer" {
			web.WriteJSON(w, http.StatusOK, errorResponse{Message: "Unauthorized."})
			slog.Info("auth.invite: user is unauthorized")
			return
		}

		var req addToCalendarRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			web.WriteJSON(w, http.StatusInternalServerError, errorResponse{Message: "Internal server error, contact the Webmaster."})
			slog.Error("events.search: decode request failed", "error", err)
			return
		}

		documentID := events.DocsUrlToID(req.URL)
		slog.Info("id", "id", documentID)
		event, err := events.AddEventToCalendar(r.Context(), documentID, calendarID, docsService, calendarService)
		if errors.Is(err, events.EventExistsError) {
			web.WriteJSON(w, http.StatusConflict, errorResponse{Message: "Event already exists in calendar."})
			slog.Error("events.search: add event to calendar failed", "error", err)
			return
		}
		if err != nil {
			web.WriteJSON(w, http.StatusInternalServerError, errorResponse{Message: "Internal server error, contact the Webmaster."})
			slog.Error("events.search: add event to calendar failed", "error", err)
			return
		}

		web.WriteJSON(w, http.StatusOK, addToCalendarResponse{URL: event.HtmlLink})
	}
}
