package auth

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type PendingLogin struct {
	ID              string     `db:"id"`
	UserID          string     `db:"user_id"`
	Email           string     `db:"email"`
	VerifyTokenHash string     `db:"verify_token_hash"`
	CreatedAt       time.Time  `db:"created_at"`
	ExpiresAt       time.Time  `db:"expires_at"`
	CompletedAt     *time.Time `db:"completed_at"`
}

// Creates a pending login for the given user. Returns the plaintext verify token for the magic link.
func CreatePendingLogin(ctx context.Context, user User, db *sqlx.DB, expiry time.Duration) (string, string, error) {
	verifyToken := MustGenerateToken()
	createdAt := time.Now()
	expiresAt := createdAt.Add(expiry)
	id := uuid.New().String()

	pendingLogin := PendingLogin{
		ID:              id,
		UserID:          user.ID,
		Email:           user.Email,
		VerifyTokenHash: MustHashToken(verifyToken),
		CreatedAt:       createdAt,
		ExpiresAt:       expiresAt,
	}

	_, err := db.NamedExecContext(ctx, `INSERT INTO pending_logins (id, user_id, email, verify_token_hash, expires_at, created_at) VALUES (:id, :user_id, :email, :verify_token_hash, :expires_at, :created_at)`, pendingLogin)
	if err != nil {
		return "", "", err
	}

	return id, verifyToken, nil
}

// Verifies a pending login and notifies the waiter
func VerifyPendingLogin(ctx context.Context, verifyToken string, db *sqlx.DB) (string, error) {
	var pendingLogin PendingLogin
	err := db.GetContext(ctx, &pendingLogin, "SELECT * FROM pending_logins WHERE verify_token_hash = ?", MustHashToken(verifyToken))
	if errors.Is(err, sql.ErrNoRows) {
		return "", PendingLoginNotFoundError
	}
	if err != nil {
		return "", err
	}
	if pendingLogin.ExpiresAt.Before(time.Now()) {
		return "", PendingLoginExpiredError
	}
	if pendingLogin.CompletedAt != nil {
		return "", PendingLoginAlreadyUsedError
	}

	_, err = db.ExecContext(ctx, "UPDATE pending_logins SET completed_at = ? WHERE id = ?", time.Now(), pendingLogin.ID)
	if err != nil {
		return "", err
	}
	return pendingLogin.ID, nil
}

// New login attempt if there's no pending login with the same id and email or if it expired or has been completed
func IsNewLoginAttempt(ctx context.Context, email string, attemptID string, db *sqlx.DB) (bool, error) {
	var pendingLogin PendingLogin
	err := db.GetContext(ctx, &pendingLogin, "SELECT * FROM pending_logins WHERE email = ? AND id = ?", email, attemptID)
	if errors.Is(err, sql.ErrNoRows) {
		return true, nil
	}
	if err != nil {
		return false, err
	}
	if pendingLogin.ExpiresAt.Before(time.Now()) {
		return true, nil
	}
	if pendingLogin.CompletedAt != nil {
		return true, nil
	}
	return false, nil
}
