package auth

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"html/template"
	"keyclub-api/email"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type Invite struct {
	ID         string     `db:"id"`
	Email      string     `db:"email"`
	FirstName  string     `db:"first_name"`
	LastName   string     `db:"last_name"`
	Role       string     `db:"role"`
	TokenHash  string     `db:"token_hash"`
	CreatedAt  time.Time  `db:"created_at"`
	ExpiresAt  time.Time  `db:"expires_at"`
	AcceptedAt *time.Time `db:"accepted_at"`
}

type InviteEmailTemplate struct {
	Subject   string
	FirstName string
	Role      string
	MagicLink string
}

var InviteNotFoundError = errors.New("Invite not found")
var InviteExpiredError = errors.New("Invite expired")
var InviteAlreadyUsedError = errors.New("Invite already used")

func SendInviteEmail(emailTemplate InviteEmailTemplate, to string, smtpConfig email.SMTPConfig) error {
	template, err := template.ParseFiles(smtpConfig.EmailTemplatePath + "/invite.html")
	if err != nil {
		return err
	}

	buf := bytes.Buffer{}
	if err := template.Execute(&buf, emailTemplate); err != nil {
		return err
	}

	htmlBody := buf.String()
	emailMessage := email.EmailMessage{
		To:       to,
		Subject:  emailTemplate.Subject,
		HTMLBody: htmlBody,
	}
	return email.SendEmail(emailMessage, smtpConfig)
}

// Creates invite, returns plaintext token for magic link
func CreateInvite(ctx context.Context, db *sqlx.DB, expiry time.Duration, email, firstName, lastName, role string) (string, error) {
	if role == "" {
		role = "member"
	}
	token := MustGenerateToken()
	tokenHash := MustHashToken(token)
	createdAt := time.Now()
	expiresAt := createdAt.Add(expiry)

	invite := Invite{
		ID:        uuid.New().String(),
		Email:     email,
		FirstName: firstName,
		LastName:  lastName,
		Role:      role,
		TokenHash: tokenHash,
		CreatedAt: createdAt,
		ExpiresAt: expiresAt,
	}

	_, err := db.NamedExecContext(ctx,
		`INSERT INTO invites (id, email, first_name, last_name, role, token_hash, created_at, expires_at) VALUES (:id, :email, :first_name, :last_name, :role, :token_hash, :created_at, :expires_at)`,
		invite,
	)
	if err != nil {
		return "", err
	}
	return token, nil
}

func AcceptInvite(ctx context.Context, queryer sqlx.ExtContext, token string) error {
	// gets invite
	var invite Invite
	err := sqlx.GetContext(ctx, queryer, &invite, "SELECT * FROM invites WHERE token_hash = ?", MustHashToken(token))
	if errors.Is(err, sql.ErrNoRows) {
		return InviteNotFoundError
	}
	if err != nil {
		return err
	}
	if invite.ExpiresAt.Before(time.Now()) {
		return InviteExpiredError
	}
	if invite.AcceptedAt != nil {
		return InviteAlreadyUsedError
	}

	// updates invite to mark it as accepted
	now := time.Now()
	_, err = sqlx.NamedExecContext(ctx, queryer, "UPDATE invites SET accepted_at = :accepted_at WHERE id = :id", map[string]any{
		"accepted_at": now,
		"id":          invite.ID,
	})
	if err != nil {
		return err
	}

	// creates new user from invite
	id := uuid.New().String()
	newUser := User{
		ID:        id,
		Email:     invite.Email,
		FirstName: invite.FirstName,
		LastName:  invite.LastName,
		Role:      invite.Role,
		CreatedAt: now,
	}
	_, err = sqlx.NamedExecContext(
		ctx,
		queryer,
		"INSERT INTO users (id, email, first_name, last_name, role, created_at) VALUES (:id, :email, :first_name, :last_name, :role, :created_at)",
		newUser,
	)
	if err != nil {
		return err
	}

	return nil
}
