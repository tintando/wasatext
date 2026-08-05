package database

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/gofrs/uuid"
)

// CreateGroup creates a new group conversation
func (db *appdbimpl) CreateGroup(name string, creatorID string, memberIDs []string) (*Group, error) {
	groupID, err := uuid.NewV4()
	if err != nil {
		return nil, fmt.Errorf("error generating group ID: %w", err)
	}

	tx, err := db.c.Begin()
	if err != nil {
		return nil, fmt.Errorf("error starting transaction: %w", err)
	}
	defer func() {
		if rbErr := tx.Rollback(); rbErr != nil && !errors.Is(rbErr, sql.ErrTxDone) {
			log.Printf("Transaction rollback failed: %v", rbErr)
		}
	}()

	now := time.Now()

	// Create conversation
	_, err = tx.Exec(
		"INSERT INTO conversations (id, type, name, created_at) VALUES (?, ?, ?, ?)",
		groupID.String(), "group", name, now,
	)
	if err != nil {
		return nil, fmt.Errorf("error creating group conversation: %w", err)
	}

	// Add creator as participant
	allMemberIDs := append([]string{creatorID}, memberIDs...)

	for _, memberID := range allMemberIDs {
		_, err = tx.Exec(
			"INSERT INTO conversation_participants (conversation_id, user_id, joined_at) VALUES (?, ?, ?)",
			groupID.String(), memberID, now,
		)
		if err != nil {
			return nil, fmt.Errorf("error adding group member: %w", err)
		}
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("error committing transaction: %w", err)
	}

	// Get group members for response
	members, err := db.GetGroupMembers(groupID.String())
	if err != nil {
		return nil, fmt.Errorf("error getting group members: %w", err)
	}

	group := &Group{
		ID:        groupID.String(),
		Name:      name,
		Members:   members,
		CreatedAt: now,
	}

	return group, nil
}

// GetGroup retrieves a group by ID
func (db *appdbimpl) GetGroup(groupID string) (*Group, error) {
	group := &Group{}
	err := db.c.QueryRow(
		"SELECT id, name, photo_url, created_at FROM conversations WHERE id = ? AND type = 'group'",
		groupID,
	).Scan(&group.ID, &group.Name, &group.PhotoURL, &group.CreatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrGroupNotFound
		}
		return nil, fmt.Errorf("error getting group: %w", err)
	}

	// Get group members
	members, err := db.GetGroupMembers(groupID)
	if err != nil {
		return nil, fmt.Errorf("error getting group members: %w", err)
	}

	group.Members = members
	return group, nil
}

// UpdateGroupName updates a group's name
func (db *appdbimpl) UpdateGroupName(groupID string, name string, userID string) error {
	// Check if user is a member of the group
	var count int
	err := db.c.QueryRow(
		"SELECT COUNT(*) FROM conversation_participants WHERE conversation_id = ? AND user_id = ?",
		groupID, userID,
	).Scan(&count)
	if err != nil {
		return fmt.Errorf("error checking group membership: %w", err)
	}

	if count == 0 {
		return ErrNotGroupMember
	}

	// Update group name
	result, err := db.c.Exec(
		"UPDATE conversations SET name = ? WHERE id = ? AND type = 'group'",
		name, groupID,
	)
	if err != nil {
		return fmt.Errorf("error updating group name: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrGroupNotFound
	}

	return nil
}

// SetGroupPhoto sets a group's photo URL
func (db *appdbimpl) SetGroupPhoto(groupID string, photoURL string, userID string) error {
	// Check if user is a member of the group
	var count int
	err := db.c.QueryRow(
		"SELECT COUNT(*) FROM conversation_participants WHERE conversation_id = ? AND user_id = ?",
		groupID, userID,
	).Scan(&count)
	if err != nil {
		return fmt.Errorf("error checking group membership: %w", err)
	}

	if count == 0 {
		return ErrNotGroupMember
	}

	// Update group photo
	result, err := db.c.Exec(
		"UPDATE conversations SET photo_url = ? WHERE id = ? AND type = 'group'",
		photoURL, groupID,
	)
	if err != nil {
		return fmt.Errorf("error setting group photo: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrGroupNotFound
	}

	return nil
}

// AddUserToGroup adds a user to a group
func (db *appdbimpl) AddUserToGroup(groupID string, userID string, adderID string) error {
	// Check if adder is a member of the group
	var count int
	err := db.c.QueryRow(
		"SELECT COUNT(*) FROM conversation_participants WHERE conversation_id = ? AND user_id = ?",
		groupID, adderID,
	).Scan(&count)
	if err != nil {
		return fmt.Errorf("error checking group membership: %w", err)
	}

	if count == 0 {
		return ErrNotGroupMember
	}

	// Check if user is already in the group
	err = db.c.QueryRow(
		"SELECT COUNT(*) FROM conversation_participants WHERE conversation_id = ? AND user_id = ?",
		groupID, userID,
	).Scan(&count)
	if err != nil {
		return fmt.Errorf("error checking if user is already in group: %w", err)
	}

	if count > 0 {
		return ErrAlreadyGroupMember
	}

	// Add user to group
	_, err = db.c.Exec(
		"INSERT INTO conversation_participants (conversation_id, user_id, joined_at) VALUES (?, ?, ?)",
		groupID, userID, time.Now(),
	)
	if err != nil {
		return fmt.Errorf("error adding user to group: %w", err)
	}

	return nil
}

// RemoveUserFromGroup removes a user from a group
func (db *appdbimpl) RemoveUserFromGroup(groupID string, userID string) error {
	result, err := db.c.Exec(
		"DELETE FROM conversation_participants WHERE conversation_id = ? AND user_id = ?",
		groupID, userID,
	)
	if err != nil {
		return fmt.Errorf("error removing user from group: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrNotGroupMember
	}

	return nil
}

// GetGroupMembers retrieves all members of a group
func (db *appdbimpl) GetGroupMembers(groupID string) ([]User, error) {
	rows, err := db.c.Query(`
		SELECT u.id, u.name, u.photo_url, u.created_at
		FROM users u
		JOIN conversation_participants cp ON u.id = cp.user_id
		WHERE cp.conversation_id = ?
		ORDER BY u.name`,
		groupID,
	)
	if err != nil {
		return nil, fmt.Errorf("error getting group members: %w", err)
	}
	defer rows.Close()

	var members []User
	for rows.Next() {
		var user User
		err := rows.Scan(&user.ID, &user.Name, &user.PhotoURL, &user.CreateAt)
		if err != nil {
			return nil, fmt.Errorf("error scanning group member: %w", err)
		}
		members = append(members, user)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating group members: %w", err)
	}

	return members, nil
}
