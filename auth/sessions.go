package auth

import (
	"time"
)

type Session struct {
	ID        string     `db:"id"`
	UserID    string     `db:"user_id"`
	TokenHash string     `db:"token_hash"`
	CreatedAt time.Time  `db:"created_at"`
	ExpiresAt time.Time  `db:"expires_at"`
	RevokedAt *time.Time `db:"revoked_at"`
}

// func CreateSession(ctx context.Context, userID string, db *sqlx.DB) (string, error) {
// 	token := auth.MustGenerateToken()
// 	createdAt := time.Now()
// 	expiresAt := createdAt.Add(24 * time.Hour)
// 	tokenHash := auth.MustHashToken(token)
// 	id := uuid.New().String()

// 	session := Session{
// 		ID:        id,
// 		UserID:    userID,
// 		TokenHash: tokenHash,
// 		CreatedAt: createdAt,
// 		ExpiresAt: expiresAt,
// 	}

// 	_, err := db.NamedExecContext(ctx, "INSERT INTO sessions (id, user_id, token_hash, created_at, expires_at) VALUES (:id, :user_id, :token_hash, :created_at, :expires_at)", session)
// 	if err != nil {
// 		return "", err
// 	}

// 	return token, nil
// }
