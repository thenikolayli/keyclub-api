package auth

import (
	"context"
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
