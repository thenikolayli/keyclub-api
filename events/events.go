package events

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/jmoiron/sqlx"
)

// struct to represent a logged event
type Event struct {
	ID            string    `db:"id"`
	Name          string    `db:"name"`
	Date          string    `db:"date"`
	StartTime     string    `db:"start_time"`
	EndTime       string    `db:"end_time"`
	Address       string    `db:"address"`
	NofSlots      int       `db:"n_of_slots"`
	NofVolunteers int       `db:"n_of_volunteers"`
	TotalHours    float64   `db:"total_hours"`
	Leaders       string    `db:"leaders"`
	MadeBy        string    `db:"made_by"`
	SignUpUrl     string    `db:"sign_up_url"`
	Description   string    `db:"description"`
	CreatedAt     time.Time `db:"created_at"`
	UpdatedAt     time.Time `db:"updated_at"`
}

// to represent a row in a sign up doc
// hours start index is the index of the hours cell, it's where you put the calculated hours
type MemberAttendance struct {
	Name            string
	Hours           float64
	HoursStartIndex int
	HoursEndIndex   int
}

// upserts event struct
func UpsertEvent(ctx context.Context, event Event, queryer sqlx.ExtContext) error {
	var result Event
	err := sqlx.GetContext(
		ctx,
		queryer,
		&result,
		"SELECT * from events WHERE name = ? LIMIT 1",
		event.Name,
	)
	if errors.Is(err, sql.ErrNoRows) {
		_, insertErr := sqlx.NamedExecContext(
			ctx,
			queryer,
			`
			INSERT INTO events
			(id, name, date, start_time, end_time, address, n_of_slots, n_of_volunteers, total_hours, leaders, made_by, sign_up_url, description, created_at, updated_at)
			VALUES
			(:id, :name, :date, :start_time, :end_time, :address, :n_of_slots, :n_of_volunteers, :total_hours, :leaders, :made_by, :sign_up_url, :description, :created_at, :updated_at)`,
			event,
		)
		if insertErr != nil {
			slog.Error("events.events: insert event failed", "error", insertErr, "name", event.Name)
			return fmt.Errorf("Issue inserting event during upsert: %v", insertErr)
		}
		slog.Info("events.events: inserted event", "name", event.Name, "id", event.ID)
	} else if err != nil {
		slog.Error("events.events: lookup event failed", "error", err, "name", event.Name)
		return fmt.Errorf("Issue upserting event: %v", err)
	} else {
		event.ID = result.ID // to update the correct row based on primary key (id)
		event.UpdatedAt = time.Now()
		_, updateErr := sqlx.NamedExecContext(
			ctx, queryer,
			`
			UPDATE events SET 
			name=:name, date=:date, start_time=:start_time, end_time=:end_time, address=:address, n_of_slots=:n_of_slots, n_of_volunteers=:n_of_volunteers, total_hours=:total_hours, leaders=:leaders, made_by=:made_by, sign_up_url=:sign_up_url, description=:description, updated_at=:updated_at
			WHERE id=:id
		`, event,
		)
		if updateErr != nil {
			slog.Error("events.events: update event failed", "error", updateErr, "name", event.Name, "id", event.ID)
			return fmt.Errorf("Issue updating event during upsert: %v", updateErr)
		}
		// slog.Info("events.events: updated event", "name", event.Name, "id", event.ID)
	}
	return nil
}

// formats the member attendances for the response
// returns a slice of strings, each string is the name and hours of the member
func FormatMemberAttendances(members []MemberAttendance) []string {
	formatted := make([]string, len(members))
	for i, member := range members {
		formatted[i] = fmt.Sprintf("%s - %.2f hours", member.Name, member.Hours)
	}
	return formatted
}
