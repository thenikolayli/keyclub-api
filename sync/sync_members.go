package sync

import (
	"context"
	"database/sql"
	"fmt"
	"keyclub-api/google"
	"keyclub-api/members"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"google.golang.org/api/sheets/v4"
)

// updates the member database entries
// fetches values via an api call to the hours spreadsheet
// formats the response to member structs
// updates the database based on structs
func SyncMembersFromSheet(ctx context.Context, googleConfig google.GoogleConfig, memberSync *SyncState, db *sqlx.DB) error {
	memberSync.Mutex.Lock()
	defer memberSync.Mutex.Unlock()

	memberValueRanges, err := getMemberValueRanges(ctx, googleConfig)
	if err != nil {
		return fmt.Errorf("Failed to update members: %v", err)
	}

	formattedMemberStructs := getMemberStructs(memberValueRanges)

	transaction, err := db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("Failed to create a transaction: %v", err)
	}
	for _, each := range formattedMemberStructs {
		err := upsertMember(ctx, each, transaction)
		if err != nil {
			return err
		}
	}
	transaction.Commit()

	memberSync.LastUpdated = time.Now()
	return nil
}

// takes in a formatted member struct and a transaction and upserts their row
// checks if a member with the same name exists
// if they don't, insert them
// otherwise, update
func upsertMember(ctx context.Context, member members.Member, transaction *sqlx.Tx) error {
	result := members.Member{}
	err := transaction.GetContext(
		ctx, &result,
		"SELECT * from members WHERE first_name = ? AND last_name = ? LIMIT 1",
		member.Firstname, member.Lastname,
	)
	if err == sql.ErrNoRows {
		_, insertErr := transaction.NamedExecContext(
			ctx,
			`INSERT INTO members
			(first_name, nickname, middle_name, last_name, all_hours, term_hours, grad_year, class, strikes, personal_email, school_email, phone_number, shirt_size, paid_dues)
			VALUES
			(:first_name, :nickname, :middle_name, :last_name, :all_hours, :term_hours, :grad_year, :class, :strikes, :personal_email, :school_email, :phone_number, :shirt_size, :paid_dues)`,
			member,
		)
		if insertErr != nil {
			return fmt.Errorf("Issue inserting member during upsert: %v", insertErr)
		}
	} else if err != nil {
		return fmt.Errorf("Issue upserting member: %v", err)
	} else {
		member.ID = result.ID // to update the correct row based on primary key (id)
		_, updateErr := transaction.NamedExecContext(
			ctx,
			`UPDATE members SET 
			first_name=:first_name, nickname=:nickname, middle_name=:middle_name, last_name=:last_name, all_hours=:all_hours, term_hours=:term_hours, grad_year=:grad_year, class=:class, strikes=:strikes, personal_email=:personal_email, school_email=:school_email, phone_number=:phone_number, shirt_size=:shirt_size, paid_dues=:paid_dues
			WHERE id=:id`,
			member,
		)
		if updateErr != nil {
			return fmt.Errorf("Issue updating member during upsert: %v", updateErr)
		}
	}
	return nil
}

// fetches and returns google sheets api value ranges (unformatted)
func getMemberValueRanges(ctx context.Context, googleConfig google.GoogleConfig) ([]*sheets.ValueRange, error) {
	r := googleConfig.MembersSheetRanges
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
		return nil, fmt.Errorf("Failed to batch get spreadsheet ranges: %v", err)
	}
	return data.ValueRanges, nil
}

// takes the api call value ranges and turns them into an array of member structs
func getMemberStructs(memberValueRanges []*sheets.ValueRange) []members.Member {
	// gets length based on the length of the names column
	memberValueRangesLength := len(memberValueRanges[0].Values)
	formattedMemberArray := make([]members.Member, memberValueRangesLength)

	normalizedNames := NormalizeStringValues(memberValueRanges[0].Values, memberValueRangesLength)
	normalizedAllHours := NormalizeFloatValues(memberValueRanges[1].Values, memberValueRangesLength)
	normalizedTermHours := NormalizeFloatValues(memberValueRanges[2].Values, memberValueRangesLength)
	normalizedGradYears := NormalizeIntValues(memberValueRanges[3].Values, memberValueRangesLength)
	normalizedClasses := NormalizeStringValues(memberValueRanges[4].Values, memberValueRangesLength)
	normalizedStrikes := NormalizeIntValues(memberValueRanges[5].Values, memberValueRangesLength)
	normalizedPersonalEmails := NormalizeStringValues(memberValueRanges[6].Values, memberValueRangesLength)
	normalizedSchoolEmails := NormalizeStringValues(memberValueRanges[7].Values, memberValueRangesLength)
	normalizedPhoneNumbers := NormalizeStringValues(memberValueRanges[8].Values, memberValueRangesLength)
	normalizedShirtSizes := NormalizeStringValues(memberValueRanges[9].Values, memberValueRangesLength)
	normalizedPaidDues := NormalizeBoolValues(memberValueRanges[10].Values, memberValueRangesLength)

	for i := range memberValueRangesLength - 1 {
		name := members.NewName(normalizedNames[i])

		formattedMemberArray[i] = members.Member{
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
			ID:            uuid.New().String(),
		}
	}

	return formattedMemberArray
}
