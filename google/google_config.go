package google

import (
	"context"
	"fmt"
	"os"

	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/docs/v1"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

type GoogleConfig struct {
	SpreadsheetID string
	CalendarID    string

	MembersSheetRanges       MembersSheetRangesType
	EventsSheetRanges        EventsSheetRangesType
	EventsMembersSheetRanges EventsMembersSheetRangesType

	DocsService     *docs.Service
	SheetsService   *sheets.Service
	CalendarService *calendar.Service
}

// returns a GoogleServices struct, containing the services used to interact with the google apis
func LoadGoogleServices(ctx context.Context, keyFilePath string) (GoogleConfig, error) {
	clientOption, err := getClientOption(keyFilePath)
	if err != nil {
		return GoogleConfig{}, fmt.Errorf("Failed to get client option: %w", err)
	}

	docsService, err := docs.NewService(ctx, clientOption)
	if err != nil {
		return GoogleConfig{}, fmt.Errorf("google.GetGoogleServices: %w", err)
	}
	sheetsService, err := sheets.NewService(ctx, clientOption)
	if err != nil {
		return GoogleConfig{}, fmt.Errorf("google.GetGoogleServices: %w", err)
	}
	calendarService, err := calendar.NewService(ctx, clientOption)
	if err != nil {
		return GoogleConfig{}, fmt.Errorf("google.GetGoogleServices: %w", err)
	}

	return GoogleConfig{
		SpreadsheetID: os.Getenv("SPREADSHEET_ID"),
		CalendarID:    os.Getenv("CALENDAR_ID"),

		MembersSheetRanges: MembersSheetRangesType{
			SheetName:     "2025-2026 Members",
			Names:         "2025-2026 Members!A2:A",
			AllHours:      "2025-2026 Members!B2:B",
			TermHours:     "2025-2026 Members!C2:C",
			GradYear:      "2025-2026 Members!D2:D",
			Class:         "2025-2026 Members!E2:E",
			Strikes:       "2025-2026 Members!F2:F",
			PersonalEmail: "2025-2026 Members!G2:G",
			SchoolEmail:   "2025-2026 Members!H2:H",
			PhoneNumber:   "2025-2026 Members!I2:I",
			ShirtSizes:    "2025-2026 Members!J2:J",
			PaidDues:      "2025-2026 Members!K2:K",
		},
		EventsMembersSheetRanges: EventsMembersSheetRangesType{
			SheetName: "2025-2026 EventsMembers",
			Events:    "2025-2026 EventsMembers!A2:A",
			Members:   "2025-2026 EventsMembers!B1:ZZ1",
		},
		EventsSheetRanges: EventsSheetRangesType{
			SheetName:     "2025-2026 Events",
			Events:        "2025-2026 Events!A2:A",
			Dates:         "2025-2026 Events!B2:B",
			StartTimes:    "2025-2026 Events!C2:C",
			EndTimes:      "2025-2026 Events!D2:D",
			Addresses:     "2025-2026 Events!E2:E",
			NofSlots:      "2025-2026 Events!F2:F",
			NofVolunteers: "2025-2026 Events!G2:G",
			TotalHours:    "2025-2026 Events!H2:H",
			Leaders:       "2025-2026 Events!I2:I",
			MadeBy:        "2025-2026 Events!J2:J",
		},

		DocsService:     docsService,
		SheetsService:   sheetsService,
		CalendarService: calendarService,
	}, nil
}

// uses the google_auth_key.json file to create client options
// this is used to get google services later
func getClientOption(keyFilepath string) (option.ClientOption, error) {
	if _, err := os.Stat(keyFilepath); os.IsNotExist(err) {
		return nil, fmt.Errorf("Service account key file not found: %w", err)
	}

	return option.WithAuthCredentialsFile(option.ServiceAccount, keyFilepath), nil
}
