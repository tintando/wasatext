package database

import "time"

// Message status constants
const (
	MessageStatusSent      = "sent"
	MessageStatusDelivered = "delivered"
	MessageStatusRead      = "read"
)

// User represents a user in the system
type User struct {
	ID       string    `json:"id" db:"id"`
	Name     string    `json:"name" db:"name"`
	PhotoURL *string   `json:"photoUrl,omitempty" db:"photo_url"`
	CreateAt time.Time `json:"-" db:"created_at"`
}

// Group represents a group conversation
type Group struct {
	ID        string    `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	PhotoURL  *string   `json:"photoUrl,omitempty" db:"photo_url"`
	Members   []User    `json:"members"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
}

// ConversationSummary represents a conversation summary for the conversations list
type ConversationSummary struct {
	ID          string          `json:"id" db:"id"`
	Type        string          `json:"type" db:"type"` // "direct" or "group"
	Name        string          `json:"name" db:"name"`
	PhotoURL    *string         `json:"photoUrl,omitempty" db:"photo_url"`
	OtherUserID *string         `json:"otherUserId,omitempty" db:"other_user_id"`
	LastMessage *MessageSummary `json:"lastMessage,omitempty"`
	UnreadCount int             `json:"unreadCount" db:"unread_count"`
}

// MessageSummary represents a message summary for conversation previews
type MessageSummary struct {
	Content     string    `json:"content,omitempty" db:"content"`
	MessageType string    `json:"messageType" db:"message_type"`
	Timestamp   time.Time `json:"timestamp" db:"timestamp"`
}

// Conversation represents a full conversation with messages
type Conversation struct {
	ID           string    `json:"id" db:"id"`
	Type         string    `json:"type" db:"type"`           // "direct" or "group"
	Name         string    `json:"name,omitempty" db:"name"` // Group name, empty for direct conversations
	Participants []User    `json:"participants"`
	Messages     []Message `json:"messages"`
}

// Message represents a message in a conversation
type Message struct {
	ID                        string     `json:"id" db:"id"`
	SenderID                  string     `json:"senderId" db:"sender_id"`
	SenderName                string     `json:"senderName" db:"sender_name"`
	Content                   *string    `json:"content,omitempty" db:"content"`
	MessageType               string     `json:"messageType" db:"message_type"` // "text" or "photo"
	PhotoURL                  *string    `json:"photoUrl,omitempty" db:"photo_url"`
	Timestamp                 time.Time  `json:"timestamp" db:"timestamp"`
	Status                    string     `json:"status" db:"status"` // "sent", "delivered", "read"
	Comments                  []Comment  `json:"comments"`
	ReplyToMessageID          *string    `json:"replyToMessageId,omitempty" db:"reply_to_message_id"`
	ForwardFromMessageID      *string    `json:"forwardFromMessageId,omitempty" db:"forwarded_from_message_id"`
	ForwardFromConversationID *string    `json:"forwardFromConversationId,omitempty" db:"forwarded_from_conversation_id"`
	DeletedAt                 *time.Time `json:"-" db:"deleted_at"`
}

// NewMessage represents a new message being sent
type NewMessage struct {
	Content          *string `json:"content,omitempty"`
	MessageType      string  `json:"messageType"`
	Photo            []byte  `json:"-"` // Binary photo data
	ReplyToMessageID *string `json:"replyToMessageId,omitempty"`
}

// Comment represents a comment/reaction on a message
type Comment struct {
	ID        string    `json:"id" db:"id"`
	UserID    string    `json:"userId" db:"user_id"`
	UserName  string    `json:"userName" db:"user_name"`
	Emoticon  string    `json:"emoticon" db:"emoticon"`
	Timestamp time.Time `json:"timestamp" db:"timestamp"`
}

// ConversationParticipant represents a participant in a conversation
type ConversationParticipant struct {
	ConversationID string     `db:"conversation_id"`
	UserID         string     `db:"user_id"`
	JoinedAt       time.Time  `db:"joined_at"`
	LastReadAt     *time.Time `db:"last_read_at"`
}

// MessageStatus represents message delivery status
type MessageStatus struct {
	MessageID string    `db:"message_id"`
	UserID    string    `db:"user_id"`
	Status    string    `db:"status"` // "sent", "delivered", "read"
	UpdatedAt time.Time `db:"updated_at"`
}
