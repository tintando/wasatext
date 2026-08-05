package database

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/gofrs/uuid"
)

// GetUserConversations retrieves all conversations for a user
func (db *appdbimpl) GetUserConversations(userID string) ([]ConversationSummary, error) {
	query := `
		SELECT 
			c.id,
			c.type,
			CASE 
				WHEN c.type = 'direct' THEN (
					SELECT u.name FROM users u 
					JOIN conversation_participants cp ON u.id = cp.user_id 
					WHERE cp.conversation_id = c.id AND cp.user_id != ?
					LIMIT 1
				)
				ELSE c.name
			END as name,
			CASE 
				WHEN c.type = 'direct' THEN (
					SELECT u.photo_url FROM users u 
					JOIN conversation_participants cp ON u.id = cp.user_id 
					WHERE cp.conversation_id = c.id AND cp.user_id != ?
					LIMIT 1
				)
				ELSE c.photo_url
			END as photo_url,
			CASE 
				WHEN c.type = 'direct' THEN (
					SELECT u.id FROM users u 
					JOIN conversation_participants cp ON u.id = cp.user_id 
					WHERE cp.conversation_id = c.id AND cp.user_id != ?
					LIMIT 1
				)
				ELSE NULL
			END as other_user_id,
			m.content,
			m.timestamp,
			CASE WHEN m.message_type = 'photo' THEN 1 ELSE 0 END as is_photo,
			CASE WHEN m.forwarded_from_message_id IS NOT NULL THEN 1 ELSE 0 END as is_forwarded,
			COALESCE(unread.unread_count, 0) as unread_count
		FROM conversations c
		JOIN conversation_participants cp ON c.id = cp.conversation_id
		LEFT JOIN (
			SELECT 
				conversation_id,
				content,
				timestamp,
				message_type,
				forwarded_from_message_id,
				ROW_NUMBER() OVER (PARTITION BY conversation_id ORDER BY timestamp DESC) as rn
			FROM messages 
			WHERE deleted_at IS NULL
		) m ON c.id = m.conversation_id AND m.rn = 1
		LEFT JOIN (
			SELECT 
				m.conversation_id,
				COUNT(*) as unread_count
			FROM messages m
			JOIN conversation_participants cp ON m.conversation_id = cp.conversation_id
			WHERE cp.user_id = ? 
			AND m.sender_id != ?
			AND m.deleted_at IS NULL
			AND (cp.last_read_at IS NULL OR m.timestamp > cp.last_read_at)
			GROUP BY m.conversation_id
		) unread ON c.id = unread.conversation_id
		WHERE cp.user_id = ?
		ORDER BY COALESCE(m.timestamp, c.created_at) DESC`

	rows, err := db.c.Query(query, userID, userID, userID, userID, userID, userID)
	if err != nil {
		return nil, fmt.Errorf("error getting user conversations: %w", err)
	}
	defer rows.Close()

	var conversations []ConversationSummary
	for rows.Next() {
		var conv ConversationSummary
		var lastMessageContent sql.NullString
		var lastMessageTimestamp sql.NullTime
		var isPhoto sql.NullBool
		var isForwarded sql.NullBool

		err := rows.Scan(
			&conv.ID,
			&conv.Type,
			&conv.Name,
			&conv.PhotoURL,
			&conv.OtherUserID,
			&lastMessageContent,
			&lastMessageTimestamp,
			&isPhoto,
			&isForwarded,
			&conv.UnreadCount,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning conversation: %w", err)
		}

		// Add last message if exists
		if lastMessageTimestamp.Valid {
			var content string
			var messageType string

			// Determine message type and content
			switch {
			case isForwarded.Bool && isPhoto.Bool:
				messageType = "forward"
				if lastMessageContent.Valid && lastMessageContent.String != "" {
					content = "🔗 Forwarded 📷 Photo: " + lastMessageContent.String
				} else {
					content = "🔗 Forwarded 📷 Photo"
				}
			case isForwarded.Bool:
				messageType = "forward"
				if lastMessageContent.Valid && lastMessageContent.String != "" {
					content = "🔗 Forwarded: " + lastMessageContent.String
				} else {
					content = "🔗 Forwarded message"
				}
			case isPhoto.Bool:
				messageType = "photo"
				if lastMessageContent.Valid && lastMessageContent.String != "" {
					content = "📷 Photo: " + lastMessageContent.String
				} else {
					content = "📷 Photo"
				}
			default:
				messageType = "text"
				if lastMessageContent.Valid && lastMessageContent.String != "" {
					content = lastMessageContent.String
				} else {
					content = "Message"
				}
			}

			conv.LastMessage = &MessageSummary{
				Content:     content,
				MessageType: messageType,
				Timestamp:   lastMessageTimestamp.Time,
			}
		}

		conversations = append(conversations, conv)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating conversations: %w", err)
	}

	return conversations, nil
}

// GetConversation retrieves a specific conversation with all messages
func (db *appdbimpl) GetConversation(conversationID string, userID string) (*Conversation, error) {
	// Check if user is a participant in the conversation
	var participantCount int
	err := db.c.QueryRow(
		"SELECT COUNT(*) FROM conversation_participants WHERE conversation_id = ? AND user_id = ?",
		conversationID, userID,
	).Scan(&participantCount)
	if err != nil {
		return nil, fmt.Errorf("error checking conversation membership: %w", err)
	}

	if participantCount == 0 {
		return nil, ErrNotAuthorized
	}

	// Get conversation details
	var conversation Conversation
	var name sql.NullString
	err = db.c.QueryRow(
		"SELECT id, type, name FROM conversations WHERE id = ?",
		conversationID,
	).Scan(&conversation.ID, &conversation.Type, &name)
	if err == nil && name.Valid {
		conversation.Name = name.String
	}
	if err != nil {
		return nil, fmt.Errorf("error getting conversation: %w", err)
	}

	// Get participants
	participantRows, err := db.c.Query(`
		SELECT u.id, u.name, u.photo_url, u.created_at
		FROM users u
		JOIN conversation_participants cp ON u.id = cp.user_id
		WHERE cp.conversation_id = ?
		ORDER BY u.name`,
		conversationID,
	)
	if err != nil {
		return nil, fmt.Errorf("error getting participants: %w", err)
	}
	defer participantRows.Close()

	for participantRows.Next() {
		var participant User
		err := participantRows.Scan(&participant.ID, &participant.Name, &participant.PhotoURL, &participant.CreateAt)
		if err != nil {
			return nil, fmt.Errorf("error scanning participant: %w", err)
		}
		conversation.Participants = append(conversation.Participants, participant)
	}

	// Check for iteration errors
	if err = participantRows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating participants: %w", err)
	}

	// Get messages with comments
	messageRows, err := db.c.Query(`
		SELECT 
			m.id, m.sender_id, u.name as sender_name, m.content, m.message_type, 
			m.photo_url, m.timestamp, m.status, m.reply_to_message_id, m.forwarded_from_message_id, m.forwarded_from_conversation_id
		FROM messages m
		JOIN users u ON m.sender_id = u.id
		WHERE m.conversation_id = ? AND m.deleted_at IS NULL
		ORDER BY m.timestamp DESC`,
		conversationID,
	)
	if err != nil {
		return nil, fmt.Errorf("error getting messages: %w", err)
	}
	defer messageRows.Close()

	messageIndexMap := make(map[string]int)
	var messages []Message
	for messageRows.Next() {
		var message Message
		err := messageRows.Scan(
			&message.ID, &message.SenderID, &message.SenderName, &message.Content,
			&message.MessageType, &message.PhotoURL, &message.Timestamp,
			&message.Status, &message.ReplyToMessageID, &message.ForwardFromMessageID,
			&message.ForwardFromConversationID,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning message: %w", err)
		}

		// For messages sent by the requesting user, calculate aggregated status
		if message.SenderID == userID {
			aggregatedStatus, err := db.GetMessageStatus(message.ID)
			if err != nil {
				// Log error but don't fail - use original status
				message.Status = MessageStatusSent
			} else {
				message.Status = aggregatedStatus
			}
		}

		message.Comments = []Comment{} // Initialize empty comments slice
		messages = append(messages, message)
		// Store the index of this message
		messageIndexMap[message.ID] = len(messages) - 1
	}

	// Check for iteration errors
	if err = messageRows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating messages: %w", err)
	}

	// Get comments for all messages
	if len(messageIndexMap) > 0 {
		messageIDs := getMessageIDsFromIndexMap(messageIndexMap)

		commentRows, err := db.c.Query(`
			SELECT c.id, c.message_id, c.user_id, u.name as user_name, c.emoticon, c.timestamp
			FROM comments c
			JOIN users u ON c.user_id = u.id
			WHERE c.message_id IN (`+getPlaceholders(len(messageIndexMap))+`)
			ORDER BY c.timestamp ASC`,
			messageIDs...,
		)
		if err != nil {
			return nil, fmt.Errorf("error getting comments: %w", err)
		}
		defer commentRows.Close()

		commentCount := 0
		for commentRows.Next() {
			var comment Comment
			var messageID string
			err := commentRows.Scan(
				&comment.ID, &messageID, &comment.UserID, &comment.UserName,
				&comment.Emoticon, &comment.Timestamp,
			)
			if err != nil {
				return nil, fmt.Errorf("error scanning comment: %w", err)
			}

			if index, exists := messageIndexMap[messageID]; exists {
				messages[index].Comments = append(messages[index].Comments, comment)
				commentCount++
			}
		}

		// Check for iteration errors
		if err = commentRows.Err(); err != nil {
			return nil, fmt.Errorf("error iterating comments: %w", err)
		}
	}

	// Assign the final messages slice to conversation
	conversation.Messages = messages

	return &conversation, nil
}

// Helper function to generate SQL placeholders
func getPlaceholders(count int) string {
	if count == 0 {
		return ""
	}
	placeholders := "?"
	for i := 1; i < count; i++ {
		placeholders += ",?"
	}
	return placeholders
}

// Helper function to get message IDs from index map
func getMessageIDsFromIndexMap(indexMap map[string]int) []interface{} {
	ids := make([]interface{}, 0, len(indexMap))
	for id := range indexMap {
		ids = append(ids, id)
	}
	return ids
}

// CreateDirectConversation creates a new direct conversation between two users
func (db *appdbimpl) CreateDirectConversation(user1ID, user2ID string) (string, error) {
	conversationID, err := uuid.NewV4()
	if err != nil {
		return "", fmt.Errorf("error generating conversation ID: %w", err)
	}

	tx, err := db.c.Begin()
	if err != nil {
		return "", fmt.Errorf("error starting transaction: %w", err)
	}
	defer func() {
		if rbErr := tx.Rollback(); rbErr != nil && !errors.Is(rbErr, sql.ErrTxDone) {
			log.Printf("Transaction rollback failed: %v", rbErr)
		}
	}()

	// Create conversation
	_, err = tx.Exec(
		"INSERT INTO conversations (id, type, created_at) VALUES (?, ?, ?)",
		conversationID.String(), "direct", time.Now(),
	)
	if err != nil {
		return "", fmt.Errorf("error creating conversation: %w", err)
	}

	// Add participants
	_, err = tx.Exec(
		"INSERT INTO conversation_participants (conversation_id, user_id, joined_at) VALUES (?, ?, ?), (?, ?, ?)",
		conversationID.String(), user1ID, time.Now(),
		conversationID.String(), user2ID, time.Now(),
	)
	if err != nil {
		return "", fmt.Errorf("error adding participants: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return "", fmt.Errorf("error committing transaction: %w", err)
	}

	return conversationID.String(), nil
}

// GetOrCreateDirectConversation gets existing or creates new direct conversation
func (db *appdbimpl) GetOrCreateDirectConversation(user1ID, user2ID string) (string, error) {
	// Check if conversation already exists
	var conversationID string
	err := db.c.QueryRow(`
		SELECT c.id FROM conversations c
		JOIN conversation_participants cp1 ON c.id = cp1.conversation_id
		JOIN conversation_participants cp2 ON c.id = cp2.conversation_id
		WHERE c.type = 'direct' 
		AND cp1.user_id = ? AND cp2.user_id = ?
		AND NOT EXISTS (
			SELECT 1 FROM conversation_participants cp3 
			WHERE cp3.conversation_id = c.id 
			AND cp3.user_id NOT IN (?, ?)
		)
	`, user1ID, user2ID, user1ID, user2ID).Scan(&conversationID)

	if err == nil {
		return conversationID, nil
	}

	if !errors.Is(err, sql.ErrNoRows) {
		return "", fmt.Errorf("error checking existing conversation: %w", err)
	}

	// Create new conversation
	return db.CreateDirectConversation(user1ID, user2ID)
}
