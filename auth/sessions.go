package auth

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type Session struct {
	ID        string     `db:"id"`
	UserID    string     `db:"user_id"`
	TokenHash string     `db:"token_hash"`
	CreatedAt time.Time  `db:"created_at"`
	ExpiresAt time.Time  `db:"expires_at"`
	RevokedAt *time.Time `db:"revoked_at"`
}

var SessionNotFoundError = errors.New("Session not found")
var SessionRevokedError = errors.New("Session revoked")
var SessionExpiredError = errors.New("Session expired")

type errorResponse struct {
	Error string `json:"error"`
}

// Creates a session and returns the token
func CreateSession(ctx context.Context, userID string, db *sqlx.DB, sessionDuration time.Duration) (string, error) {
	id := uuid.New().String()
	token := MustGenerateToken()
	createdAt := time.Now()
	expiresAt := createdAt.Add(sessionDuration)

	session := Session{
		ID:        id,
		UserID:    userID,
		TokenHash: MustHashToken(token),
		CreatedAt: createdAt,
		ExpiresAt: expiresAt,
	}

	_, err := db.NamedExecContext(ctx, "INSERT INTO sessions (id, user_id, token_hash, created_at, expires_at) VALUES (:id, :user_id, :token_hash, :created_at, :expires_at)", session)
	if err != nil {
		return "", err
	}

	return token, nil
}

// Gets a session by its token
func GetSessionByToken(ctx context.Context, token string, db *sqlx.DB) (Session, error) {
	var session Session
	err := db.GetContext(ctx, &session, "SELECT * FROM sessions WHERE token_hash = ?", MustHashToken(token))
	if errors.Is(err, sql.ErrNoRows) {
		return Session{}, SessionNotFoundError
	}
	if err != nil {
		return Session{}, err
	}
	return session, nil
}

// Checks if a session is valid
func IsValidSession(ctx context.Context, session Session, db *sqlx.DB) (bool, error) {
	if session.RevokedAt != nil {
		return false, SessionRevokedError
	}
	if session.ExpiresAt.Before(time.Now()) {
		return false, SessionExpiredError
	}
	return true, nil
}

// Revokes a session
func RevokeSessionBySessionToken(ctx context.Context, sessionToken string, db *sqlx.DB) error {
	_, err := db.ExecContext(ctx, "UPDATE sessions SET revoked_at = ? WHERE token_hash = ?", time.Now(), MustHashToken(sessionToken))
	if err != nil {
		return err
	}
	return nil
}
