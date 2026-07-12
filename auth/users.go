package auth

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// User represents a user in the database.
// Role level represents their level: member, leader, officer
// Though, members shouldn't have roles, technically...
type User struct {
	ID        string    `db:"id"`
	Email     string    `db:"email"`
	FirstName string    `db:"first_name"`
	LastName  string    `db:"last_name"`
	Role      string    `db:"role"`
	CreatedAt time.Time `db:"created_at"`
}

var UserNotFoundError = errors.New("User not found")
var PendingLoginNotFoundError = errors.New("Pending login not found")
var PendingLoginExpiredError = errors.New("Pending login expired")
var PendingLoginAlreadyUsedError = errors.New("Pending login already used")

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

// Gets a user ID by their pending login attempt ID
func GetUserIDByAttemptID(ctx context.Context, attemptID string, db *sqlx.DB) (string, error) {
	var pendingLogin PendingLogin
	err := db.GetContext(ctx, &pendingLogin, "SELECT * FROM pending_logins WHERE id = ?", attemptID)
	if errors.Is(err, sql.ErrNoRows) {
		return "", PendingLoginNotFoundError
	}
	if err != nil {
		return "", err
	}
	return pendingLogin.UserID, nil
}

// Gets a user by their ID
func GetUserByID(ctx context.Context, id string, db *sqlx.DB) (User, error) {
	var user User
	err := db.GetContext(ctx, &user, "SELECT * FROM users WHERE id = ?", id)
	if errors.Is(err, sql.ErrNoRows) {
		return User{}, UserNotFoundError
	}
	if err != nil {
		return User{}, err
	}
	return user, nil
}

// Creates a new user in the database
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
