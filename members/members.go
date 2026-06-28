package members

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// struct to represent a member
type Member struct {
	ID            string    `db:"id"`
	Firstname     string    `db:"first_name"`
	Nickname      string    `db:"nickname"`
	Middlename    string    `db:"middle_name"`
	Lastname      string    `db:"last_name"`
	AllHours      float64   `db:"all_hours"`
	TermHours     float64   `db:"term_hours"`
	GradYear      int       `db:"grad_year"`
	Class         string    `db:"class"`
	Strikes       int       `db:"strikes"`
	PersonalEmail string    `db:"personal_email"`
	SchoolEmail   string    `db:"school_email"`
	PhoneNumber   string    `db:"phone_number"`
	ShirtSize     string    `db:"shirt_size"`
	PaidDues      bool      `db:"paid_dues"`
	CreatedAt     time.Time `db:"created_at"`
	UpdatedAt     time.Time `db:"updated_at"`
}

// For functions like member lookup
type FormattedMember struct {
	Name          string
	AllHours      float64
	TermHours     float64
	GradYear      int
	Class         string
	Strikes       int
	PersonalEmail string
	SchoolEmail   string
	PhoneNumber   string
	ShirtSize     string
	PaidDues      bool
}

// struct to represnt a formatted name
type Name struct {
	First  string
	Nick   string
	Middle string
	Last   string
}

// formats a member for a more readable output, such as for the member lookup command
func (member Member) Format() FormattedMember {
	name := cases.Title(language.English).String(member.Firstname)
	if member.Nickname != "" {
		name += cases.Title(language.English).String(fmt.Sprintf(` "%v" `, member.Nickname))
	}
	if member.Middlename != "" {
		name += " " + cases.Title(language.English).String(member.Middlename)
	}
	name += " " + cases.Title(language.English).String(member.Lastname)

	if member.PersonalEmail == "" {
		member.PersonalEmail = "N/A"
	}
	if member.SchoolEmail == "" {
		member.SchoolEmail = "N/A"
	}
	if member.PhoneNumber == "" {
		member.PhoneNumber = "N/A"
	} else {
		member.PhoneNumber = formatPhoneNumber(member.PhoneNumber)
	}
	if member.ShirtSize == "" {
		member.ShirtSize = "N/A"
	}
	return FormattedMember{
		Name:          name,
		AllHours:      member.AllHours,
		TermHours:     member.TermHours,
		GradYear:      member.GradYear,
		Class:         member.Class,
		Strikes:       member.Strikes,
		PersonalEmail: member.PersonalEmail,
		SchoolEmail:   member.SchoolEmail,
		PhoneNumber:   member.PhoneNumber,
		ShirtSize:     member.ShirtSize,
		PaidDues:      member.PaidDues,
	}
}

// gets a member via name
func GetMember(ctx context.Context, db *sqlx.DB, name string) (Member, error) {
	formattedName := NewName(name)
	result := Member{}
	var err error

	// if first and last was given, try both, then reverse (if they put in last first)
	// then try by nickname if given, then by first
	if formattedName.First != "" && formattedName.Last != "" {
		err = db.GetContext(
			ctx, &result,
			`SELECT * FROM members WHERE first_name = ? AND last_name = ? LIMIT 1`,
			formattedName.First, formattedName.Last,
		)
		if err == sql.ErrNoRows {
			err = db.GetContext(
				ctx, &result,
				`SELECT * FROM members WHERE first_name = ? AND last_name = ? LIMIT 1`,
				formattedName.Last, formattedName.First,
			)
		}
	} else if formattedName.Nick != "" {
		err = db.GetContext(
			ctx, &result,
			`SELECT * FROM members WHERE nickname = ? OR first_name = ? LIMIT 1`,
			formattedName.Nick, formattedName.First,
		)
	}
	if err == sql.ErrNoRows {
		return Member{}, fmt.Errorf("No member found with the name %v", name)
	}
	if err != nil {
		return Member{}, fmt.Errorf("Error getting member hours: %v", err)
	}

	return result, nil
}

// upserts member into the database
func UpsertMember(ctx context.Context, member Member, queryer sqlx.ExtContext) error {
	var result Member
	err := sqlx.GetContext(
		ctx, queryer,
		&result,
		"SELECT * from members WHERE first_name = ? AND last_name = ? LIMIT 1",
		member.Firstname, member.Lastname,
	)
	if errors.Is(err, sql.ErrNoRows) {
		_, insertErr := sqlx.NamedExecContext(
			ctx, queryer,
			`INSERT INTO members
			(id, first_name, nickname, middle_name, last_name, all_hours, term_hours, grad_year, class, strikes, personal_email, school_email, phone_number, shirt_size, paid_dues, created_at, updated_at)
			VALUES
			(:id, :first_name, :nickname, :middle_name, :last_name, :all_hours, :term_hours, :grad_year, :class, :strikes, :personal_email, :school_email, :phone_number, :shirt_size, :paid_dues, :created_at, :updated_at)`,
			member,
		)
		if insertErr != nil {
			slog.Error("members.members: insert member failed", "error", insertErr, "first_name", member.Firstname, "last_name", member.Lastname)
			return fmt.Errorf("Issue inserting member during upsert: %v", insertErr)
		}
		slog.Info("members.members: inserted member", "first_name", member.Firstname, "last_name", member.Lastname, "id", member.ID)
	} else if err != nil {
		slog.Error("members.members: lookup member failed", "error", err, "first_name", member.Firstname, "last_name", member.Lastname)
		return fmt.Errorf("Issue upserting member: %v", err)
	} else {
		member.ID = result.ID
		member.UpdatedAt = time.Now()
		_, updateErr := sqlx.NamedExecContext(
			ctx,
			queryer,
			`UPDATE members SET 
			first_name=:first_name, nickname=:nickname, middle_name=:middle_name, last_name=:last_name, all_hours=:all_hours, term_hours=:term_hours, grad_year=:grad_year, class=:class, strikes=:strikes, personal_email=:personal_email, school_email=:school_email, phone_number=:phone_number, shirt_size=:shirt_size, paid_dues=:paid_dues, created_at=:created_at, updated_at=:updated_at
			WHERE id=:id`,
			member,
		)
		if updateErr != nil {
			slog.Error("members.members: update member failed", "error", updateErr, "first_name", member.Firstname, "last_name", member.Lastname, "id", member.ID)
			return fmt.Errorf("Issue updating member during upsert: %v", updateErr)
		}
	}
	return nil
}

// formats phone numbers into this standard format: (XXX) XXX-XXXX
func formatPhoneNumber(phoneNumber string) string {
	cleanNumber := strings.ReplaceAll(phoneNumber, " ", "")
	cleanNumber = strings.ReplaceAll(cleanNumber, "-", "")
	cleanNumber = strings.ReplaceAll(cleanNumber, "(", "")
	cleanNumber = strings.ReplaceAll(cleanNumber, ")", "")

	if len(cleanNumber) == 10 {
		return fmt.Sprintf("(%s) %s-%s", cleanNumber[0:3], cleanNumber[3:6], cleanNumber[6:10])
	} else {
		return phoneNumber // if the number isn't 10 digits, return it as is
	}
}

// creates a new instance of type Name based on a string input
// this is a standalone function because it's often called on a string input, not a member struct
// names are in a First "Nick" Middle Last format
func NewName(name string) Name {
	nameParts := strings.Fields(name)

	if len(nameParts) == 2 {
		return Name{
			First: strings.ToLower(strings.Trim(nameParts[0], `"`)),
			Last:  strings.ToLower(strings.Trim(nameParts[1], `"`)),
		}
	} else if len(nameParts) == 3 {
		// First "Nick" Last vs First Middle Last
		if strings.Contains(nameParts[1], `"`) {
			return Name{
				First: strings.ToLower(strings.Trim(nameParts[0], `"`)),
				Nick:  strings.ToLower(strings.Trim(nameParts[1], `"`)),
				Last:  strings.ToLower(strings.Trim(nameParts[2], `"`)),
			}
		} else {
			return Name{
				First:  strings.ToLower(strings.Trim(nameParts[0], `"`)),
				Middle: strings.ToLower(strings.Trim(nameParts[1], `"`)),
				Last:   strings.ToLower(strings.Trim(nameParts[2], `"`)),
			}
		}
	} else if len(nameParts) == 4 {
		return Name{
			First:  strings.ToLower(strings.Trim(nameParts[0], `"`)),
			Nick:   strings.ToLower(strings.Trim(nameParts[1], `"`)),
			Middle: strings.ToLower(strings.Trim(nameParts[2], `"`)),
			Last:   strings.ToLower(strings.Trim(nameParts[3], `"`)),
		}
	}
	return Name{
		First: strings.ToLower(strings.Trim(nameParts[0], `"`)),
		Nick:  strings.ToLower(strings.Trim(nameParts[0], `"`)),
	}
}

// name speaks for itself
func SameName(name1 Name, name2 Name) bool {
	if name1.First == name2.First && name1.Last == name2.Last {
		return true
	}
	if name1.First == name2.Last && name1.Last == name2.First {
		return true
	}
	if name1.Nick != "" && (name1.Nick == name2.First || name1.Nick == name2.Nick) {
		return true
	}
	return false
}
