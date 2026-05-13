package auth

import "time"

// User represents a user in the database.
// Role level represents their level: 0 = member, 1 = leader, 2 = officer
type User struct {
	ID        int       `db:"id"`
	CreatedAt time.Time `db:"created_at"`
	FirstName string    `db:"first_name"`
	LastName  string    `db:"last_name"`
	Email     string    `db:"email"`
	RoleLevel int       `db:"role_level"`
}

type Session struct {
	ID        int        `db:"id"`
	UserID    int        `db:"user_id"`
	TokenHash string     `db:"token_hash"`
	CreatedAt time.Time  `db:"created_at"`
	ExpiresAt time.Time  `db:"expires_at"`
	RevokedAt *time.Time `db:"revoked_at"`
	LoginIP   string     `db:"login_ip"`
	UserAgent string     `db:"user_agent"`
}

type Invite struct {
	ID        int        `db:"id"`
	CreatedAt time.Time  `db:"created_at"`
	ExpiresAt time.Time  `db:"expires_at"`
	UsedAt    *time.Time `db:"used_at"`
	Email     string     `db:"email"`
	RoleLevel int        `db:"role_level"`
}
