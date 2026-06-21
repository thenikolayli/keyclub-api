package sync

import (
	"context"
	"fmt"
	"keyclub-api/events"
	"keyclub-api/google"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"google.golang.org/api/sheets/v4"
)

func SyncEvents(ctx context.Context, googleConfig google.GoogleConfig, eventSync *SyncState, db *sqlx.DB) error {
	slog.Info("sync.events: starting")

	eventSync.Mutex.Lock()
	defer eventSync.Mutex.Unlock()

	if err := syncEventsFromSheet(ctx, googleConfig, db); err != nil {
		return err
	}
	if err := syncEventsFromCalendar(ctx, googleConfig, db); err != nil {
		return err
	}

	eventSync.LastUpdated = time.Now()
	slog.Info("sync.events: completed")
	return nil
}

// syncs the events database entries from the events spreadsheet
// fetches values via an api call to the events spreadsheet
// formats the response to event structs
// updates the database based on structs
func syncEventsFromSheet(ctx context.Context, googleConfig google.GoogleConfig, db *sqlx.DB) error {
	eventValueRanges, err := getEventValueRanges(ctx, googleConfig)
	if err != nil {
		return fmt.Errorf("Failed to update events: %w", err)
	}
	formattedEventStructs := getEventStructs(eventValueRanges)

	transaction, err := db.BeginTxx(ctx, nil)
	if err != nil {
		slog.Error("sync.events: begin transaction failed", "error", err)
		return fmt.Errorf("Failed to create a transaction: %v", err)
	}
	defer transaction.Rollback()

	for _, each := range formattedEventStructs {
		err := events.UpsertEvent(ctx, each, transaction)
		if err != nil {
			return err
		}
	}
	if err := transaction.Commit(); err != nil {
		slog.Error("sync.events: commit transaction failed", "error", err)
		return fmt.Errorf("Failed to commit transaction: %w", err)
	}
	return nil
}

// scans calendar for events in the upcoming month, upserts them into the database
func syncEventsFromCalendar(ctx context.Context, googleConfig google.GoogleConfig, db *sqlx.DB) error {
	slog.Info("sync.events: fetching calendar events", "calendar_id", googleConfig.CalendarID)

	timeMin := time.Now().Format(time.RFC3339)
	timeMax := time.Now().AddDate(0, 1, 0).Format(time.RFC3339)
	calendarEvents, err := googleConfig.CalendarService.Events.List(googleConfig.CalendarID).TimeMin(timeMin).TimeMax(timeMax).Context(ctx).Do()
	if err != nil {
		slog.Error("sync.events: list calendar events failed", "error", err, "calendar_id", googleConfig.CalendarID)
		return fmt.Errorf("Failed to get events from calendar: %w", err)
	}
	slog.Info("sync.events: fetched calendar events", "count", len(calendarEvents.Items))

	formattedEvents := make([]events.Event, 0, len(calendarEvents.Items))
	for _, calEvent := range calendarEvents.Items {
		if len(calEvent.Attachments) == 0 {
			continue
		}
		attendanceDoc, err := events.ParseAttendanceDoc(ctx, events.DocsUrlToID(calEvent.Attachments[0].FileUrl), googleConfig.DocsService)
		if err != nil {
			slog.Warn("sync.events: get event info failed", "event", calEvent.Summary, "error", err)
			continue
		}
		formattedEvents = append(formattedEvents, attendanceDoc.Event)
	}
	slog.Info("sync.events: parsed calendar events", "count", len(formattedEvents))

	transaction, err := db.BeginTxx(ctx, nil)
	if err != nil {
		slog.Error("sync.events: begin transaction failed", "error", err)
		return fmt.Errorf("Failed to create a transaction: %v", err)
	}
	defer transaction.Rollback()

	for _, event := range formattedEvents {
		if err := events.UpsertEvent(ctx, event, transaction); err != nil {
			return err
		}
	}
	if err := transaction.Commit(); err != nil {
		slog.Error("sync.events: commit transaction failed", "error", err)
		return fmt.Errorf("Failed to commit transaction: %v", err)
	}
	return nil
}

// fetches and returns google sheets api value ranges (unformatted)
// description and sign up url aren't stored in the sheet and aren't synced from or written to it
func getEventValueRanges(ctx context.Context, googleConfig google.GoogleConfig) ([]*sheets.ValueRange, error) {
	ranges := googleConfig.EventsSheetRanges
	slog.Info("sync.events: fetching sheet ranges", "spreadsheet_id", googleConfig.SpreadsheetID, "sheet", ranges.SheetName)

	data, err := googleConfig.SheetsService.Spreadsheets.Values.BatchGet(googleConfig.SpreadsheetID).Ranges(
		ranges.Events,
		ranges.Dates,
		ranges.StartTimes,
		ranges.EndTimes,
		ranges.Addresses,
		ranges.NofSlots,
		ranges.NofVolunteers,
		ranges.TotalHours,
		ranges.Leaders,
		ranges.MadeBy,
	).Context(ctx).Do()
	if err != nil {
		slog.Error("sync.events: batch get sheet ranges failed", "error", err, "spreadsheet_id", googleConfig.SpreadsheetID)
		return nil, fmt.Errorf("Failed to batch get spreadsheet ranges: %v", err)
	}

	slog.Info("sync.events: fetched sheet ranges", "columns", len(data.ValueRanges))
	return data.ValueRanges, nil
}

// takes the api call value ranges and turns them into an array of event structs
func getEventStructs(eventValueRanges []*sheets.ValueRange) []events.Event {
	if len(eventValueRanges) == 0 {
		slog.Warn("sync.events: no value ranges returned from sheet")
		return nil
	}

	// gets length based on the length of the events column
	eventValueRangesLength := len(eventValueRanges[0].Values)
	formattedEventArray := make([]events.Event, eventValueRangesLength)

	normalizedEvents := Normalize(eventValueRanges[0].Values, eventValueRangesLength, Parsers.String)
	normalizedDates := Normalize(eventValueRanges[1].Values, eventValueRangesLength, Parsers.String)
	normalizedStartTimes := Normalize(eventValueRanges[2].Values, eventValueRangesLength, Parsers.String)
	normalizedEndTimes := Normalize(eventValueRanges[3].Values, eventValueRangesLength, Parsers.String)
	normalizedAddresses := Normalize(eventValueRanges[4].Values, eventValueRangesLength, Parsers.String)
	normalizedNofSlots := Normalize(eventValueRanges[5].Values, eventValueRangesLength, Parsers.Int)
	normalizedNofVolunteers := Normalize(eventValueRanges[6].Values, eventValueRangesLength, Parsers.Int)
	normalizedTotalHours := Normalize(eventValueRanges[7].Values, eventValueRangesLength, Parsers.Float)
	normalizedLeaders := Normalize(eventValueRanges[8].Values, eventValueRangesLength, Parsers.String)
	normalizedMadeBy := Normalize(eventValueRanges[9].Values, eventValueRangesLength, Parsers.String)

	for i := range eventValueRangesLength {
		formattedEventArray[i] = events.Event{
			ID:            uuid.New().String(),
			Name:          normalizedEvents[i],
			Date:          normalizedDates[i],
			StartTime:     normalizedStartTimes[i],
			EndTime:       normalizedEndTimes[i],
			Address:       normalizedAddresses[i],
			NofSlots:      normalizedNofSlots[i],
			NofVolunteers: normalizedNofVolunteers[i],
			TotalHours:    normalizedTotalHours[i],
			Leaders:       normalizedLeaders[i],
			MadeBy:        normalizedMadeBy[i],
			SignUpUrl:     "", // sheet doesn't have this column
			Description:   "", // sheet doesn't have this column
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}
	}

	slog.Info("sync.events: built event structs from sheet", "count", len(formattedEventArray))
	return formattedEventArray
}
