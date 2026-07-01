package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"keyclub-api/events"
	"keyclub-api/web"
	"log/slog"
	"net/http"
	"time"

	"github.com/jmoiron/sqlx"
)

type searchRequest struct {
	Times  []int     `json:"times"`
	Length []float64 `json:"length"`
	Dates  []string  `json:"dates"`
	Spots  []int     `json:"spots"`
}

type formattedEvent struct {
	Name         string  `json:"name"`
	Date         string  `json:"date"`
	StartTime    string  `json:"start_time"`
	EndTime      string  `json:"end_time"`
	Length       float64 `json:"length"`
	Address      string  `json:"address"`
	NofOpenSlots int     `json:"n_of_open_slots"`
	SignUpUrl    string  `json:"sign_up_url"`
	Description  string  `json:"description"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func SearchHandler(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req searchRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			web.WriteJSON(w, http.StatusInternalServerError, errorResponse{Error: "Internal server error, contact the Webmaster."})
			slog.Error("events.search: decode request failed", "error", err)
			return
		}

		minTime := fmt.Sprintf("%02d:00:00", req.Times[0])
		maxTime := fmt.Sprintf("%02d:00:00", req.Times[1])
		minLength := req.Length[0]
		maxLength := req.Length[1]
		minDate := req.Dates[0]
		maxDate := req.Dates[1]
		minSpots := req.Spots[0]
		maxSpots := req.Spots[1]
		query := `
			SELECT * FROM events
			WHERE start_time BETWEEN ? AND ?
			AND end_time BETWEEN ? AND ?
			AND (julianday(end_time) - julianday(start_time)) * 24 BETWEEN ? AND ?
			AND date BETWEEN ? AND ?
			AND n_of_slots - n_of_volunteers BETWEEN ? AND ?
		`
		var events []events.Event
		err := db.SelectContext(r.Context(), &events, query,
			minTime, maxTime,
			minTime, maxTime,
			minLength, maxLength,
			minDate, maxDate,
			minSpots, maxSpots,
		)
		if errors.Is(err, sql.ErrNoRows) {
			web.WriteJSON(w, http.StatusNotFound, errorResponse{Error: "No events found."})
			slog.Error("events.search: no events found")
			return
		}
		if err != nil {
			web.WriteJSON(w, http.StatusInternalServerError, errorResponse{Error: "Internal server error, contact the Webmaster."})
			slog.Error("events.search: select events failed", "error", err)
			return
		}

		formattedEvents := make([]formattedEvent, len(events))
		for i, event := range events {
			endTime, err := time.Parse("15:04:05", event.EndTime)
			if err != nil {
				web.WriteJSON(w, http.StatusInternalServerError, errorResponse{Error: "Internal server error, contact the Webmaster."})
				slog.Error("events.search: parse end time failed", "error", err)
				return
			}
			startTime, err := time.Parse("15:04:05", event.StartTime)
			if err != nil {
				web.WriteJSON(w, http.StatusInternalServerError, errorResponse{Error: "Internal server error, contact the Webmaster."})
				slog.Error("events.search: parse start time failed", "error", err)
				return
			}
			length := endTime.Sub(startTime)

			formattedEvents[i] = formattedEvent{
				Name:         event.Name,
				Date:         event.Date,
				StartTime:    event.StartTime,
				EndTime:      event.EndTime,
				Length:       length.Hours(),
				Address:      event.Address,
				NofOpenSlots: event.NofSlots - event.NofVolunteers,
				SignUpUrl:    event.SignUpUrl,
				Description:  event.Description,
			}
		}

		web.WriteJSON(w, http.StatusOK, formattedEvents)
		slog.Info("events.search: search successful")
	}
}
