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
	ID        string    `db:"id" json:"id"`
	Email     string    `db:"email" json:"email"`
	FirstName string    `db:"first_name" json:"firstName"`
	LastName  string    `db:"last_name" json:"lastName"`
	Role      string    `db:"role" json:"role"`
	CreatedAt time.Time `db:"created_at" json:"createdAt"`
	UpdatedAt *time.Time `db:"updated_at" json:"updatedAt"`
}

var UserNotFoundError = errors.New("User not found")

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

func ListUsers(ctx context.Context, db *sqlx.DB, skip, limit int) ([]User, error) {
	var users []User
	err := db.SelectContext(ctx, &users, "SELECT * FROM users LIMIT ? OFFSET ?", limit, skip)
	if err != nil {
		return []User{}, err
	}
	return users, nil
}

func UpdateUser(ctx context.Context, db *sqlx.DB, updatedUser User) error {
	now := time.Now()
	updatedUser.UpdatedAt = &now
	_, err := db.NamedExecContext(
		ctx,
		`UPDATE users SET email = :email, first_name = :first_name, last_name = :last_name, role = :role, updated_at = :updated_at WHERE id = :id`,
		updatedUser,
	)
	return err
}

func DeleteUser(ctx context.Context, db *sqlx.DB, userID string) error {
	_, err := db.ExecContext(ctx, "DELETE FROM users WHERE id = ?", userID)
	return err
}

// Creates a new user in the database
func CreateUser(ctx context.Context, db *sqlx.DB, email, firstName, lastName, role string) (User, error) {
	if role == "" {
		role = "member"
	}
	id := uuid.New().String()
	createdAt := time.Now()

	user := User{
		ID:        id,
		Email:     email,
		FirstName: firstName,
		LastName:  lastName,
		Role:      role,
		CreatedAt: createdAt,
	}
	_, err := db.NamedExecContext(
		ctx,
		`INSERT INTO users (id, email, first_name, last_name, role, created_at) VALUES (:id, :email, :first_name, :last_name, :role, :created_at)`,
		user,
	)
	if err != nil {
		return User{}, err
	}

	return user, nil
}
