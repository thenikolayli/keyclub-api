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
)

// struct to represent a member
type Member struct {
	ID            string    `db:"id"`
	Name          string    `db:"name"`
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

type MemberToken struct {
	MemberID string `db:"member_id"`
	Token    string `db:"token"`
}

var MemberNotFoundError = errors.New("Member not found")

// gets a member via name
func GetMember(ctx context.Context, queryer sqlx.ExtContext, name string) (Member, error) {
	tokenizedName := TokenizeName(name)
	query := "SELECT member_id FROM member_tokens WHERE token = ?"
	args := []any{tokenizedName[0]}
	for _, token := range tokenizedName[1:] {
		query += "INTERSECT SELECT member_id FROM member_tokens WHERE token = ?"
		args = append(args, token)
	}

	var result string
	err := sqlx.GetContext(ctx, queryer, &result, query, args...)
	if errors.Is(err, sql.ErrNoRows) {
		return Member{}, MemberNotFoundError
	}
	if err != nil {
		return Member{}, fmt.Errorf("Error occurred while fetching member: %v", err)
	}

	var member Member
	err = sqlx.GetContext(
		ctx,
		queryer,
		&member,
		`SELECT * FROM members WHERE id = ?`,
		result,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return Member{}, MemberNotFoundError
	}
	if err != nil {
		return Member{}, fmt.Errorf("Error occurred while fetching member: %v", err)
	}

	return member, nil
}

// upserts member into the database
func UpsertMember(ctx context.Context, member Member, queryer sqlx.ExtContext) error {
	result, err := GetMember(ctx, queryer, member.Name)
	if errors.Is(err, MemberNotFoundError) {
		_, insertErr := sqlx.NamedExecContext(
			ctx,
			queryer,
			`INSERT INTO members
			(id, name, all_hours, term_hours, grad_year, class, strikes, personal_email, school_email, phone_number, shirt_size, paid_dues, created_at, updated_at)
			VALUES
			(:id, :name, :all_hours, :term_hours, :grad_year, :class, :strikes, :personal_email, :school_email, :phone_number, :shirt_size, :paid_dues, :created_at, :updated_at)`,
			member,
		)
		if insertErr != nil {
			slog.Error("members.members: insert member failed", "error", insertErr, "name", member.Name, "id", member.ID)
			return fmt.Errorf("Issue inserting member during upsert: %v", insertErr)
		}

		slog.Info("members.members: inserted member", "name", member.Name, "id", member.ID)
	} else if err != nil {
		slog.Error("members.members: lookup member failed", "error", err, "name", member.Name, "id", member.ID)
		return fmt.Errorf("Issue upserting member: %v", err)
	} else {
		sqlx.MustExecContext(
			ctx,
			queryer,
			`DELETE FROM member_tokens WHERE member_id = ?`,
			result.ID,
		)

		member.ID = result.ID
		member.UpdatedAt = time.Now()
		member.CreatedAt = result.CreatedAt

		_, updateErr := sqlx.NamedExecContext(
			ctx,
			queryer,
			`UPDATE members SET 
			name=:name, all_hours=:all_hours, term_hours=:term_hours, grad_year=:grad_year, class=:class, strikes=:strikes, personal_email=:personal_email, school_email=:school_email, phone_number=:phone_number, shirt_size=:shirt_size, paid_dues=:paid_dues, created_at=:created_at, updated_at=:updated_at
			WHERE id=:id`,
			member,
		)
		if updateErr != nil {
			slog.Error("members.members: update member failed", "error", updateErr, "name", member.Name, "id", member.ID)
			return fmt.Errorf("Issue updating member during upsert: %v", updateErr)
		}
	}

	for _, token := range TokenizeName(member.Name) {
		_, insertErr := sqlx.NamedExecContext(
			ctx,
			queryer,
			`INSERT INTO member_tokens (member_id, token) VALUES (:member_id, :token)`,
			MemberToken{MemberID: member.ID, Token: token},
		)
		if insertErr != nil {
			slog.Error("members.members: insert member failed", "error", insertErr, "name", member.Name, "id", member.ID)
			return fmt.Errorf("Issue inserting member during upsert: %v", insertErr)
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

func TokenizeName(name string) []string {
	tokens := make(map[string]struct{})
	for _, token := range strings.Fields(name) {
		token := strings.ToLower(token)
		token = strings.Trim(token, ".,;:!?\"'()[]{}<>")
		if token != "" {
			tokens[token] = struct{}{}
		}
	}
	tokensSlice := []string{}
	for token := range tokens {
		tokensSlice = append(tokensSlice, token)
	}
	return tokensSlice
}
