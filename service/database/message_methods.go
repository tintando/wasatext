package database

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/gofrs/uuid"
)

// SendMessage sends a new message to a conversation
func (db *appdbimpl) SendMessage(conversationID string, senderID string, message NewMessage) (*Message, error) {
	// Check if user is a participant in the conversation
	var participantCount int
	err := db.c.QueryRow(
		"SELECT COUNT(*) FROM conversation_participants WHERE conversation_id = ? AND user_id = ?",
		conversationID, senderID,
	).Scan(&participantCount)
	if err != nil {
		return nil, fmt.Errorf("error checking conversation membership: %w", err)
	}

	if participantCount == 0 {
		return nil, ErrNotParticipant
	}

	messageID, err := uuid.NewV4()
	if err != nil {
		return nil, fmt.Errorf("error generating message ID: %w", err)
	}

	msg := &Message{
		ID:               messageID.String(),
		SenderID:         senderID,
		Content:          message.Content,
		MessageType:      message.MessageType,
		Timestamp:        time.Now(),
		Status:           MessageStatusSent,
		ReplyToMessageID: message.ReplyToMessageID,
		Comments:         []Comment{},
	}

	// Get sender name
	err = db.c.QueryRow("SELECT name FROM users WHERE id = ?", senderID).Scan(&msg.SenderName)
	if err != nil {
		return nil, fmt.Errorf("error getting sender name: %w", err)
	}

	// Start transaction for message and status creation
	tx, err := db.c.Begin()
	if err != nil {
		return nil, fmt.Errorf("error starting transaction: %w", err)
	}
	defer func() {
		if rbErr := tx.Rollback(); rbErr != nil && !errors.Is(rbErr, sql.ErrTxDone) {
			log.Printf("Transaction rollback failed: %v", rbErr)
		}
	}()

	// Insert message
	_, err = tx.Exec(`
		INSERT INTO messages (id, conversation_id, sender_id, content, message_type, photo_url, timestamp, status, reply_to_message_id)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		msg.ID, conversationID, msg.SenderID, msg.Content, msg.MessageType, msg.PhotoURL, msg.Timestamp, msg.Status, msg.ReplyToMessageID,
	)
	if err != nil {
		return nil, fmt.Errorf("error inserting message: %w", err)
	}

	// Get all conversation participants (intended recipients snapshot)
	participants, err := tx.Query(
		"SELECT user_id FROM conversation_participants WHERE conversation_id = ?",
		conversationID,
	)
	if err != nil {
		return nil, fmt.Errorf("error getting conversation participants: %w", err)
	}
	defer participants.Close()

	var recipientIDs []string
	for participants.Next() {
		var participantID string
		if err := participants.Scan(&participantID); err != nil {
			return nil, fmt.Errorf("error scanning participant: %w", err)
		}

		// Skip sender - they don't need to receive their own message
		if participantID != senderID {
			recipientIDs = append(recipientIDs, participantID)
		}
	}

	// Check for iteration errors
	if err = participants.Err(); err != nil {
		return nil, fmt.Errorf("error iterating participants: %w", err)
	}

	// Record intended recipients (snapshot at send time)
	for _, recipientID := range recipientIDs {
		_, err = tx.Exec(`
			INSERT INTO message_recipients (message_id, user_id)
			VALUES (?, ?)`,
			msg.ID, recipientID,
		)
		if err != nil {
			return nil, fmt.Errorf("error recording message recipient: %w", err)
		}

		// Initialize message status as 'sent' for each recipient
		_, err = tx.Exec(`
			INSERT INTO message_status (message_id, user_id, status, updated_at)
			VALUES (?, ?, ?, ?)`,
			msg.ID, recipientID, MessageStatusSent, msg.Timestamp,
		)
		if err != nil {
			return nil, fmt.Errorf("error inserting message status: %w", err)
		}
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("error committing transaction: %w", err)
	}

	return msg, nil
}

// SetMessagePhoto sets a message's photo URL
func (db *appdbimpl) SetMessagePhoto(messageID string, filename string) error {
	result, err := db.c.Exec(
		"UPDATE messages SET photo_url = ? WHERE id = ?",
		filename, messageID,
	)
	if err != nil {
		return fmt.Errorf("error setting message photo: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrMessageNotFound
	}

	return nil
}

// DeleteMessage soft deletes a message
func (db *appdbimpl) DeleteMessage(messageID string, userID string) error {
	result, err := db.c.Exec(
		"UPDATE messages SET deleted_at = ? WHERE id = ? AND sender_id = ?",
		time.Now(), messageID, userID,
	)
	if err != nil {
		return fmt.Errorf("error deleting message: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrMessageNotFound
	}

	return nil
}

// ForwardMessage forwards an existing message to another conversation
func (db *appdbimpl) ForwardMessage(messageID string, targetConversationID string, userID string) (*Message, error) {
	// Check if user is a participant in the target conversation
	var participantCount int
	err := db.c.QueryRow(
		"SELECT COUNT(*) FROM conversation_participants WHERE conversation_id = ? AND user_id = ?",
		targetConversationID, userID,
	).Scan(&participantCount)
	if err != nil {
		return nil, fmt.Errorf("error checking conversation membership: %w", err)
	}

	if participantCount == 0 {
		return nil, ErrNotParticipant
	}

	// Get original message with conversation ID
	var originalMessage Message
	var originalConversationID string
	err = db.c.QueryRow(`
		SELECT content, message_type, photo_url, conversation_id FROM messages 
		WHERE id = ? AND deleted_at IS NULL`,
		messageID,
	).Scan(&originalMessage.Content, &originalMessage.MessageType, &originalMessage.PhotoURL, &originalConversationID)

	if err != nil {
		return nil, fmt.Errorf("error getting original message: %w", err)
	}

	// Create a new message ID for the forwarded message
	newMessageID, err := uuid.NewV4()
	if err != nil {
		return nil, fmt.Errorf("error generating message ID: %w", err)
	}

	// Create the forwarded message
	forwardedMessage := &Message{
		ID:                        newMessageID.String(),
		SenderID:                  userID,
		Content:                   originalMessage.Content,
		MessageType:               originalMessage.MessageType,
		PhotoURL:                  originalMessage.PhotoURL, // Copy photo URL for photo messages
		Timestamp:                 time.Now(),
		Status:                    MessageStatusSent,
		ForwardFromMessageID:      &messageID,
		ForwardFromConversationID: &originalConversationID,
		Comments:                  []Comment{},
	}

	// Get sender name
	err = db.c.QueryRow("SELECT name FROM users WHERE id = ?", userID).Scan(&forwardedMessage.SenderName)
	if err != nil {
		return nil, fmt.Errorf("error getting sender name: %w", err)
	}

	// Start transaction for message and status creation
	tx, err := db.c.Begin()
	if err != nil {
		return nil, fmt.Errorf("error starting transaction: %w", err)
	}
	defer func() {
		if rbErr := tx.Rollback(); rbErr != nil && !errors.Is(rbErr, sql.ErrTxDone) {
			log.Printf("Transaction rollback failed: %v", rbErr)
		}
	}()

	// Insert forwarded message
	_, err = tx.Exec(`
		INSERT INTO messages (id, conversation_id, sender_id, content, message_type, photo_url, timestamp, status, forwarded_from_message_id, forwarded_from_conversation_id)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		forwardedMessage.ID, targetConversationID, forwardedMessage.SenderID, forwardedMessage.Content,
		forwardedMessage.MessageType, forwardedMessage.PhotoURL, forwardedMessage.Timestamp, forwardedMessage.Status,
		forwardedMessage.ForwardFromMessageID, forwardedMessage.ForwardFromConversationID,
	)
	if err != nil {
		return nil, fmt.Errorf("error inserting forwarded message: %w", err)
	}

	// Get all conversation participants (intended recipients snapshot)
	participants, err := tx.Query(
		"SELECT user_id FROM conversation_participants WHERE conversation_id = ?",
		targetConversationID,
	)
	if err != nil {
		return nil, fmt.Errorf("error getting conversation participants: %w", err)
	}
	defer participants.Close()

	var recipientIDs []string
	for participants.Next() {
		var participantID string
		if err := participants.Scan(&participantID); err != nil {
			return nil, fmt.Errorf("error scanning participant: %w", err)
		}

		// Skip sender - they don't need to receive their own message
		if participantID != userID {
			recipientIDs = append(recipientIDs, participantID)
		}
	}

	// Check for iteration errors
	if err = participants.Err(); err != nil {
		return nil, fmt.Errorf("error iterating participants: %w", err)
	}

	// Record intended recipients (snapshot at send time)
	for _, recipientID := range recipientIDs {
		_, err = tx.Exec(`
			INSERT INTO message_recipients (message_id, user_id)
			VALUES (?, ?)`,
			forwardedMessage.ID, recipientID,
		)
		if err != nil {
			return nil, fmt.Errorf("error recording message recipient: %w", err)
		}

		// Initialize message status as 'sent' for each recipient
		_, err = tx.Exec(`
			INSERT INTO message_status (message_id, user_id, status, updated_at)
			VALUES (?, ?, ?, ?)`,
			forwardedMessage.ID, recipientID, MessageStatusSent, forwardedMessage.Timestamp,
		)
		if err != nil {
			return nil, fmt.Errorf("error inserting message status: %w", err)
		}
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("error committing transaction: %w", err)
	}

	return forwardedMessage, nil
}

// MarkMessagesAsRead marks all messages in a conversation as read for a user
func (db *appdbimpl) MarkMessagesAsRead(conversationID string, userID string) error {
	// Update last_read_at for the participant
	_, err := db.c.Exec(
		"UPDATE conversation_participants SET last_read_at = ? WHERE conversation_id = ? AND user_id = ?",
		time.Now(), conversationID, userID,
	)
	if err != nil {
		return fmt.Errorf("error updating last read time: %w", err)
	}

	return nil
}

// MarkMessagesAsDelivered marks messages as delivered when user sees conversation list
func (db *appdbimpl) MarkMessagesAsDelivered(userID string) error {
	_, err := db.c.Exec(`
		UPDATE message_status 
		SET status = ?, updated_at = ? 
		WHERE user_id = ? AND status = ?`,
		MessageStatusDelivered, time.Now(), userID, MessageStatusSent,
	)
	if err != nil {
		return fmt.Errorf("error marking messages as delivered: %w", err)
	}
	return nil
}

// MarkConversationMessagesAsRead marks all unread messages in a conversation as read
func (db *appdbimpl) MarkConversationMessagesAsRead(conversationID string, userID string) error {
	_, err := db.c.Exec(`
		UPDATE message_status 
		SET status = ?, updated_at = ? 
		WHERE user_id = ? 
		AND message_id IN (
			SELECT id FROM messages 
			WHERE conversation_id = ? AND deleted_at IS NULL
		)
		AND status IN (?, ?)`,
		MessageStatusRead, time.Now(), userID, conversationID, MessageStatusSent, MessageStatusDelivered,
	)
	if err != nil {
		return fmt.Errorf("error marking conversation messages as read: %w", err)
	}
	return nil
}

// GetMessageStatus calculates the aggregated status for a message (for display to sender)
func (db *appdbimpl) GetMessageStatus(messageID string) (string, error) {
	// Get total number of intended recipients for this message
	var totalRecipients int
	err := db.c.QueryRow(
		"SELECT COUNT(*) FROM message_recipients WHERE message_id = ?",
		messageID,
	).Scan(&totalRecipients)
	if err != nil {
		return MessageStatusSent, fmt.Errorf("error getting recipient count: %w", err)
	}

	// If no recipients (shouldn't happen), return sent
	if totalRecipients == 0 {
		return MessageStatusSent, nil
	}

	// Check if ALL recipients have read the message
	var readCount int
	err = db.c.QueryRow(`
		SELECT COUNT(*) FROM message_status ms
		JOIN message_recipients mr ON ms.message_id = mr.message_id AND ms.user_id = mr.user_id
		WHERE ms.message_id = ? AND ms.status = ?`,
		messageID, MessageStatusRead,
	).Scan(&readCount)
	if err != nil {
		return MessageStatusSent, fmt.Errorf("error getting read count: %w", err)
	}

	if readCount == totalRecipients {
		return MessageStatusRead, nil
	}

	// Check if ALL recipients have at least received (delivered) the message
	var deliveredCount int
	err = db.c.QueryRow(`
		SELECT COUNT(*) FROM message_status ms
		JOIN message_recipients mr ON ms.message_id = mr.message_id AND ms.user_id = mr.user_id
		WHERE ms.message_id = ? AND ms.status IN (?, ?)`,
		messageID, MessageStatusDelivered, MessageStatusRead,
	).Scan(&deliveredCount)
	if err != nil {
		return MessageStatusSent, fmt.Errorf("error getting delivered count: %w", err)
	}

	if deliveredCount == totalRecipients {
		return MessageStatusDelivered, nil
	}

	// Not all recipients have received it yet
	return MessageStatusSent, nil
}

// UpdateMessageStatus updates the status of a specific message for a user
func (db *appdbimpl) UpdateMessageStatus(messageID string, userID string, status string) error {
	// Check if user is a recipient of this message
	var recipientCount int
	err := db.c.QueryRow(
		"SELECT COUNT(*) FROM message_recipients WHERE message_id = ? AND user_id = ?",
		messageID, userID,
	).Scan(&recipientCount)
	if err != nil {
		return fmt.Errorf("error checking message recipient: %w", err)
	}

	if recipientCount == 0 {
		return ErrNotRecipient
	}

	// Validate status
	if status != MessageStatusDelivered && status != MessageStatusRead {
		return fmt.Errorf("invalid status: %s", status)
	}

	// Update message status
	_, err = db.c.Exec(`
		UPDATE message_status 
		SET status = ?, updated_at = ? 
		WHERE message_id = ? AND user_id = ?`,
		status, time.Now(), messageID, userID,
	)
	if err != nil {
		return fmt.Errorf("error updating message status: %w", err)
	}

	return nil
}
