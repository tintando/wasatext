package database

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gofrs/uuid"
)

// CreateUser creates a new user with the given name
func (db *appdbimpl) CreateUser(name string) (*User, error) {
	// Generate a new UUID for the user
	userID, err := uuid.NewV4()
	if err != nil {
		return nil, fmt.Errorf("error generating user ID: %w", err)
	}

	user := &User{
		ID:       userID.String(),
		Name:     name,
		CreateAt: time.Now(),
	}

	_, err = db.c.Exec(
		"INSERT INTO users (id, name, created_at) VALUES (?, ?, ?)",
		user.ID, user.Name, user.CreateAt,
	)
	if err != nil {
		return nil, fmt.Errorf("error creating user: %w", err)
	}

	return user, nil
}

// GetUserByID retrieves a user by their ID
func (db *appdbimpl) GetUserByID(userID string) (*User, error) {
	user := &User{}
	err := db.c.QueryRow(
		"SELECT id, name, photo_url, created_at FROM users WHERE id = ?",
		userID,
	).Scan(&user.ID, &user.Name, &user.PhotoURL, &user.CreateAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("error getting user by ID: %w", err)
	}

	return user, nil
}

// GetUserByName retrieves a user by their name
func (db *appdbimpl) GetUserByName(name string) (*User, error) {
	user := &User{}
	err := db.c.QueryRow(
		"SELECT id, name, photo_url, created_at FROM users WHERE name = ?",
		name,
	).Scan(&user.ID, &user.Name, &user.PhotoURL, &user.CreateAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("error getting user by name: %w", err)
	}

	return user, nil
}

// UpdateUserName updates a user's name
func (db *appdbimpl) UpdateUserName(userID string, name string) error {
	result, err := db.c.Exec(
		"UPDATE users SET name = ? WHERE id = ?",
		name, userID,
	)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return ErrUsernameTaken
		}
		return fmt.Errorf("error updating user name: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrUserNotFound
	}

	return nil
}

// SetUserPhoto sets a user's profile photo URL
func (db *appdbimpl) SetUserPhoto(userID string, photoURL string) error {
	result, err := db.c.Exec(
		"UPDATE users SET photo_url = ? WHERE id = ?",
		photoURL, userID,
	)
	if err != nil {
		return fmt.Errorf("error setting user photo: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrUserNotFound
	}

	return nil
}

// SearchUsers searches for users by name (case-insensitive partial match)
func (db *appdbimpl) SearchUsers(query string) ([]User, error) {
	searchPattern := "%" + strings.ToLower(query) + "%"

	rows, err := db.c.Query(
		"SELECT id, name, photo_url, created_at FROM users WHERE LOWER(name) LIKE ? ORDER BY name",
		searchPattern,
	)
	if err != nil {
		return nil, fmt.Errorf("error searching users: %w", err)
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		err := rows.Scan(&user.ID, &user.Name, &user.PhotoURL, &user.CreateAt)
		if err != nil {
			return nil, fmt.Errorf("error scanning user: %w", err)
		}
		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating users: %w", err)
	}

	return users, nil
}
