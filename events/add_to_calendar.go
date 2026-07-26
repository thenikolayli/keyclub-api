package events

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"time"

	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/docs/v1"
)

var EventExistsError = fmt.Errorf("Event already exists in calendar")

// Takes an attendance document ID and adds the event to the calendar, returning the link to the calendar event
func AddEventToCalendar(ctx context.Context, documentID string, calendarID string, docsService *docs.Service, calendarService *calendar.Service) (calendar.Event, error) {
	attendanceDoc, err := ParseAttendanceDoc(ctx, documentID, docsService)
	if err != nil {
		return calendar.Event{}, fmt.Errorf("Issue extracting event info while adding event to calendar: %w", err)
	}
	event := attendanceDoc.Event
	noDate := regexp.MustCompile(`\(.*?\)`)

	calendarEvent := &calendar.Event{
		Summary:     noDate.ReplaceAllString(event.Name, ""),
		Location:    event.Address,
		Description: event.Description,
		Start: &calendar.EventDateTime{
			DateTime: fmt.Sprintf("%sT%s", event.Date, event.StartTime),
			TimeZone: "America/Los_Angeles",
		},
		End: &calendar.EventDateTime{
			DateTime: fmt.Sprintf("%sT%s", event.Date, event.EndTime),
			TimeZone: "America/Los_Angeles",
		},
		Attachments: []*calendar.EventAttachment{
			{
				FileUrl: fmt.Sprintf("https://docs.google.com/document/d/%s/edit?tab=t.0", documentID),
				Title:   "Attendance Document",
			},
		},
	}

	err = alreadyExists(ctx, calendarService, calendarID, calendarEvent)
	if errors.Is(err, EventExistsError) {
		return calendar.Event{}, err
	}
	if err != nil {
		return calendar.Event{}, fmt.Errorf("Issue checking if event already exists: %w", err)
	}

	result, err := calendarService.Events.Insert(calendarID, calendarEvent).Context(ctx).SupportsAttachments(true).Do()
	if err != nil {
		return calendar.Event{}, fmt.Errorf("Issue inserting event into calendar: %w", err)
	}
	return *result, nil
}

func alreadyExists(ctx context.Context, calendarService *calendar.Service, calendarID string, event *calendar.Event) error {
	loc, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		return fmt.Errorf("failed to load timezone: %w", err)
	}
	startTime, err := time.ParseInLocation("2006-01-02T15:04:05", event.Start.DateTime, loc)
	if err != nil {
		return fmt.Errorf("failed to parse start time: %w", err)
	}
	endTime, err := time.ParseInLocation("2006-01-02T15:04:05", event.End.DateTime, loc)
	if err != nil {
		return fmt.Errorf("failed to parse end time: %w", err)
	}

	result, err := calendarService.Events.List(calendarID).TimeMin(startTime.Format(time.RFC3339)).TimeMax(endTime.Format(time.RFC3339)).Context(ctx).Do()
	if err != nil {
		fmt.Printf("Issue checking if event already exists: %v\n", err)
		return err
	}

	for _, item := range result.Items {
		if item.Summary == event.Summary {
			fmt.Printf("Event already exists: %s\n", item.HtmlLink)
			return EventExistsError
		}
	}
	return nil
}
