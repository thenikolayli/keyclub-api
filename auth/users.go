package auth

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

var UserNotFoundError = errors.New("user not found")

// Checks if a user with the given email exists
func UserExists(ctx context.Context, email string, db *sqlx.DB) (bool, error) {
	var user User
	err := db.GetContext(ctx, &user, "SELECT * FROM users WHERE email = ?", email)
	if errors.Is(err, sql.ErrNoRows) {
		return false, UserNotFoundError
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// Gets a user by their email and returns them
func GetUserByEmail(ctx context.Context, email string, db *sqlx.DB) (User, error) {
	var user User
	err := db.GetContext(ctx, &user, "SELECT * FROM users WHERE email = ?", email)
	if errors.Is(err, sql.ErrNoRows) {
		return User{}, UserNotFoundError
	}
	if err != nil {
		return User{}, err
	}
	return user, nil
}

func CreateUser(ctx context.Context, db *sqlx.DB, email, firstName, lastName, role string) (User, error) {
	if role == "" {
		role = "member"
	}

	user := User{
		ID:        uuid.New().String(),
		Email:     email,
		FirstName: firstName,
		LastName:  lastName,
		Role:      role,
		CreatedAt: time.Now(),
	}

	_, err := db.NamedExecContext(ctx,
		`INSERT INTO users (id, email, first_name, last_name, role, created_at) VALUES (:id, :email, :first_name, :last_name, :role, :created_at)`,
		user,
	)
	if err != nil {
		return User{}, err
	}

	return user, nil
}

// Creates a pending login for the given email
func CreatePendingLogin(ctx context.Context, user User, db *sqlx.DB, expiry time.Duration) (string, string, error) {
	loginID := MustGenerateToken()
	verifyToken := MustGenerateToken()
	expiresAt := time.Now().Add(expiry)
	createdAt := time.Now()

	pendingLogin := PendingLogin{
		ID:          uuid.New().String(),
		UserID:      user.ID,
		Email:       user.Email,
		LoginID:     MustHashToken(loginID),
		VerifyToken: MustHashToken(verifyToken),
		ExpiresAt:   expiresAt,
		CreatedAt:   createdAt,
	}

	_, err := db.NamedExecContext(ctx, `INSERT INTO pending_logins (id, user_id, email, login_id, verify_token, expires_at, created_at) VALUES (:id, :user_id, :email, :login_id, :verify_token, :expires_at, :created_at)`, pendingLogin)
	if err != nil {
		return "", "", err
	}

	return loginID, verifyToken, nil
}
