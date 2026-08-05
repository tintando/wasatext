package database

import (
	"fmt"
	"time"

	"github.com/gofrs/uuid"
)

// AddComment adds a comment/reaction to a message
func (db *appdbimpl) AddComment(messageID string, userID string, emoticon string) (*Comment, error) {
	// Check if user is trying to react to their own message
	var messageSenderID string
	err := db.c.QueryRow("SELECT sender_id FROM messages WHERE id = ? AND deleted_at IS NULL", messageID).Scan(&messageSenderID)
	if err != nil {
		return nil, fmt.Errorf("error getting message sender: %w", err)
	}

	if messageSenderID == userID {
		return nil, ErrSelfReaction
	}

	commentID, err := uuid.NewV4()
	if err != nil {
		return nil, fmt.Errorf("error generating comment ID: %w", err)
	}

	comment := &Comment{
		ID:        commentID.String(),
		UserID:    userID,
		Emoticon:  emoticon,
		Timestamp: time.Now(),
	}

	// Get user name
	err = db.c.QueryRow("SELECT name FROM users WHERE id = ?", userID).Scan(&comment.UserName)
	if err != nil {
		return nil, fmt.Errorf("error getting user name: %w", err)
	}

	// Replace existing comment if user already reacted to this message
	_, err = db.c.Exec(
		"INSERT OR REPLACE INTO comments (id, message_id, user_id, emoticon, timestamp) VALUES (?, ?, ?, ?, ?)",
		comment.ID, messageID, comment.UserID, comment.Emoticon, comment.Timestamp,
	)
	if err != nil {
		return nil, fmt.Errorf("error adding comment: %w", err)
	}

	return comment, nil
}

// DeleteComment removes a comment from a message
func (db *appdbimpl) DeleteComment(commentID string, userID string) error {
	result, err := db.c.Exec(
		"DELETE FROM comments WHERE id = ? AND user_id = ?",
		commentID, userID,
	)
	if err != nil {
		return fmt.Errorf("error deleting comment: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrCommentNotFound
	}

	return nil
}

// DeleteCommentByMessageAndUser removes a user's reaction from a message
func (db *appdbimpl) DeleteCommentByMessageAndUser(messageID string, userID string) error {
	result, err := db.c.Exec(
		"DELETE FROM comments WHERE message_id = ? AND user_id = ?",
		messageID, userID,
	)
	if err != nil {
		return fmt.Errorf("error deleting comment: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrCommentNotFound
	}

	return nil
}

// GetUserReactionToMessage returns the user's reaction to a message, if any
func (db *appdbimpl) GetUserReactionToMessage(messageID string, userID string) (*Comment, error) {
	var comment Comment
	err := db.c.QueryRow(
		"SELECT id, user_id, emoticon, timestamp FROM comments WHERE message_id = ? AND user_id = ?",
		messageID, userID,
	).Scan(&comment.ID, &comment.UserID, &comment.Emoticon, &comment.Timestamp)

	if err != nil {
		return nil, err // Will be sql.ErrNoRows if no reaction found
	}

	// Get user name
	err = db.c.QueryRow("SELECT name FROM users WHERE id = ?", userID).Scan(&comment.UserName)
	if err != nil {
		return nil, fmt.Errorf("error getting user name: %w", err)
	}

	return &comment, nil
}
