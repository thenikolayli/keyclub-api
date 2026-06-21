package sync

import (
	"context"
	"fmt"
	"keyclub-api/google"
	"keyclub-api/members"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"google.golang.org/api/sheets/v4"
)

// updates the member database entries
// fetches values via an api call to the hours spreadsheet
// formats the response to member structs
// updates the database based on structs
func SyncMembers(ctx context.Context, googleConfig google.GoogleConfig, memberSync *SyncState, db *sqlx.DB) error {
	slog.Info("sync.members: starting")

	memberSync.Mutex.Lock()
	defer memberSync.Mutex.Unlock()

	memberValueRanges, err := getMemberValueRanges(ctx, googleConfig)
	if err != nil {
		return fmt.Errorf("Failed to update members: %v", err)
	}

	formattedMemberStructs := getMemberStructs(memberValueRanges)

	transaction, err := db.BeginTxx(ctx, nil)
	if err != nil {
		slog.Error("sync.members: begin transaction failed", "error", err)
		return fmt.Errorf("Failed to create a transaction: %v", err)
	}
	defer transaction.Rollback()

	for _, each := range formattedMemberStructs {
		if err := members.UpsertMember(ctx, each, transaction); err != nil {
			return err
		}
	}
	if err := transaction.Commit(); err != nil {
		slog.Error("sync.members: commit transaction failed", "error", err)
		return fmt.Errorf("Failed to commit transaction: %v", err)
	}

	memberSync.LastUpdated = time.Now()
	slog.Info("sync.members: completed", "count", len(formattedMemberStructs))
	return nil
}

// fetches and returns google sheets api value ranges (unformatted)
func getMemberValueRanges(ctx context.Context, googleConfig google.GoogleConfig) ([]*sheets.ValueRange, error) {
	r := googleConfig.MembersSheetRanges
	slog.Info("sync.members: fetching sheet ranges", "spreadsheet_id", googleConfig.SpreadsheetID, "sheet", r.SheetName)

	data, err := googleConfig.SheetsService.Spreadsheets.Values.BatchGet(googleConfig.SpreadsheetID).Ranges(
		r.Names,
		r.AllHours,
		r.TermHours,
		r.GradYear,
		r.Class,
		r.Strikes,
		r.PersonalEmail,
		r.SchoolEmail,
		r.PhoneNumber,
		r.ShirtSizes,
		r.PaidDues,
	).Context(ctx).Do()
	if err != nil {
		slog.Error("sync.members: batch get sheet ranges failed", "error", err, "spreadsheet_id", googleConfig.SpreadsheetID)
		return nil, fmt.Errorf("Failed to batch get spreadsheet ranges: %v", err)
	}

	slog.Info("sync.members: fetched sheet ranges", "columns", len(data.ValueRanges))
	return data.ValueRanges, nil
}

// takes the api call value ranges and turns them into an array of member structs
func getMemberStructs(memberValueRanges []*sheets.ValueRange) []members.Member {
	if len(memberValueRanges) == 0 {
		slog.Warn("sync.members: no value ranges returned from sheet")
		return nil
	}

	memberValueRangesLength := len(memberValueRanges[0].Values)
	formattedMemberArray := make([]members.Member, memberValueRangesLength)

	normalizedNames := Normalize(memberValueRanges[0].Values, memberValueRangesLength, Parsers.String)
	normalizedAllHours := Normalize(memberValueRanges[1].Values, memberValueRangesLength, Parsers.Float)
	normalizedTermHours := Normalize(memberValueRanges[2].Values, memberValueRangesLength, Parsers.Float)
	normalizedGradYears := Normalize(memberValueRanges[3].Values, memberValueRangesLength, Parsers.Int)
	normalizedClasses := Normalize(memberValueRanges[4].Values, memberValueRangesLength, Parsers.String)
	normalizedStrikes := Normalize(memberValueRanges[5].Values, memberValueRangesLength, Parsers.Int)
	normalizedPersonalEmails := Normalize(memberValueRanges[6].Values, memberValueRangesLength, Parsers.String)
	normalizedSchoolEmails := Normalize(memberValueRanges[7].Values, memberValueRangesLength, Parsers.String)
	normalizedPhoneNumbers := Normalize(memberValueRanges[8].Values, memberValueRangesLength, Parsers.String)
	normalizedShirtSizes := Normalize(memberValueRanges[9].Values, memberValueRangesLength, Parsers.String)
	normalizedPaidDues := Normalize(memberValueRanges[10].Values, memberValueRangesLength, Parsers.Bool)

	for i := range memberValueRangesLength {
		name := members.NewName(normalizedNames[i])

		formattedMemberArray[i] = members.Member{
			ID:            uuid.New().String(),
			Firstname:     name.First,
			Nickname:      name.Nick,
			Middlename:    name.Middle,
			Lastname:      name.Last,
			AllHours:      normalizedAllHours[i],
			TermHours:     normalizedTermHours[i],
			GradYear:      normalizedGradYears[i],
			Class:         normalizedClasses[i],
			PersonalEmail: normalizedPersonalEmails[i],
			SchoolEmail:   normalizedSchoolEmails[i],
			PhoneNumber:   normalizedPhoneNumbers[i],
			Strikes:       normalizedStrikes[i],
			ShirtSize:     normalizedShirtSizes[i],
			PaidDues:      normalizedPaidDues[i],
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}
	}

	slog.Info("sync.members: built member structs from sheet", "count", len(formattedMemberArray))
	return formattedMemberArray
}
