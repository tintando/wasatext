/*
Package database is the middleware between the app database and the code. All data (de)serialization (save/load) from a
persistent database are handled here. Database specific logic should never escape this package.

To use this package you need to apply migrations to the database if needed/wanted, connect to it (using the database
data source name from config), and then initialize an instance of AppDatabase from the DB connection.

For example, this code adds a parameter in `webapi` executable for the database data source name (add it to the
main.WebAPIConfiguration structure):

	DB struct {
		Filename string `conf:""`
	}

This is an example on how to migrate the DB and connect to it:

	// Start Database
	logger.Println("initializing database support")
	db, err := sql.Open("sqlite3", "./foo.db")
	if err != nil {
		logger.WithError(err).Error("error opening SQLite DB")
		return fmt.Errorf("opening SQLite: %w", err)
	}
	defer func() {
		logger.Debug("database stopping")
		_ = db.Close()
	}()

Then you can initialize the AppDatabase and pass it to the api package.
*/
package database

import (
	"database/sql"
	"errors"
	"fmt"
)

// AppDatabase is the high level interface for the DB
type AppDatabase interface {
	Ping() error

	// User methods
	CreateUser(name string) (*User, error)
	GetUserByID(userID string) (*User, error)
	GetUserByName(name string) (*User, error)
	UpdateUserName(userID string, name string) error
	SetUserPhoto(userID string, photoURL string) error
	SearchUsers(query string) ([]User, error)

	// Conversation methods
	GetUserConversations(userID string) ([]ConversationSummary, error)
	GetConversation(conversationID string, userID string) (*Conversation, error)
	CreateDirectConversation(user1ID, user2ID string) (string, error)
	GetOrCreateDirectConversation(user1ID, user2ID string) (string, error)

	// Message methods
	SendMessage(conversationID string, senderID string, message NewMessage) (*Message, error)
	SetMessagePhoto(messageID string, filename string) error
	DeleteMessage(messageID string, userID string) error
	ForwardMessage(messageID string, targetConversationID string, userID string) (*Message, error)
	MarkMessagesAsRead(conversationID string, userID string) error
	MarkMessagesAsDelivered(userID string) error
	MarkConversationMessagesAsRead(conversationID string, userID string) error
	GetMessageStatus(messageID string) (string, error)
	UpdateMessageStatus(messageID string, userID string, status string) error

	// Comment methods
	AddComment(messageID string, userID string, emoticon string) (*Comment, error)
	DeleteComment(commentID string, userID string) error
	DeleteCommentByMessageAndUser(messageID string, userID string) error

	// Group methods
	CreateGroup(name string, creatorID string, memberIDs []string) (*Group, error)
	GetGroup(groupID string) (*Group, error)
	UpdateGroupName(groupID string, name string, userID string) error
	SetGroupPhoto(groupID string, photoURL string, userID string) error
	AddUserToGroup(groupID string, userID string, adderID string) error
	RemoveUserFromGroup(groupID string, userID string) error
	GetGroupMembers(groupID string) ([]User, error)
}

type appdbimpl struct {
	c *sql.DB
}

// New returns a new instance of AppDatabase based on the SQLite connection `db`.
// `db` is required - an error will be returned if `db` is `nil`.
func New(db *sql.DB) (AppDatabase, error) {
	if db == nil {
		return nil, errors.New("database is required when building a AppDatabase")
	}

	// Check if tables exist. If not, the database is empty, and we need to create the structure
	err := createTables(db)
	if err != nil {
		return nil, fmt.Errorf("error creating database structure: %w", err)
	}

	return &appdbimpl{
		c: db,
	}, nil
}

func (db *appdbimpl) Ping() error {
	return db.c.Ping()
}

// createTables creates all the necessary tables for WASAText
func createTables(db *sql.DB) error {
	// Enable foreign key support in SQLite
	_, err := db.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		return fmt.Errorf("error enabling foreign keys: %w", err)
	}

	// Users table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL UNIQUE,
			photo_url TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("error creating users table: %w", err)
	}

	// Conversations table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS conversations (
			id TEXT PRIMARY KEY,
			type TEXT NOT NULL CHECK (type IN ('direct', 'group')),
			name TEXT,
			photo_url TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("error creating conversations table: %w", err)
	}

	// Conversation participants table (many-to-many relationship)
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS conversation_participants (
			conversation_id TEXT NOT NULL,
			user_id TEXT NOT NULL,
			joined_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			last_read_at DATETIME,
			PRIMARY KEY (conversation_id, user_id),
			FOREIGN KEY (conversation_id) REFERENCES conversations(id) ON DELETE CASCADE,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)
	`)
	if err != nil {
		return fmt.Errorf("error creating conversation_participants table: %w", err)
	}

	// Messages table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS messages (
			id TEXT PRIMARY KEY,
			conversation_id TEXT NOT NULL,
			sender_id TEXT NOT NULL,
			content TEXT,
			message_type TEXT NOT NULL CHECK (message_type IN ('text', 'photo')),
			photo_url TEXT,
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
			status TEXT NOT NULL DEFAULT 'sent' CHECK (status IN ('sent', 'delivered', 'read')),
			reply_to_message_id TEXT,
			forwarded_from_message_id TEXT,
			forwarded_from_conversation_id TEXT,
			deleted_at DATETIME,
			FOREIGN KEY (conversation_id) REFERENCES conversations(id) ON DELETE CASCADE,
			FOREIGN KEY (sender_id) REFERENCES users(id) ON DELETE CASCADE,
			FOREIGN KEY (reply_to_message_id) REFERENCES messages(id) ON DELETE SET NULL,
			FOREIGN KEY (forwarded_from_message_id) REFERENCES messages(id) ON DELETE SET NULL
		)
	`)
	if err != nil {
		return fmt.Errorf("error creating messages table: %w", err)
	}

	// Comments table (reactions on messages)
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS comments (
			id TEXT PRIMARY KEY,
			message_id TEXT NOT NULL,
			user_id TEXT NOT NULL,
			emoticon TEXT NOT NULL,
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (message_id) REFERENCES messages(id) ON DELETE CASCADE,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
			UNIQUE(message_id, user_id)
		)
	`)
	if err != nil {
		return fmt.Errorf("error creating comments table: %w", err)
	}

	// Message status table (for tracking read/delivered status per user)
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS message_status (
			message_id TEXT NOT NULL,
			user_id TEXT NOT NULL,
			status TEXT NOT NULL CHECK (status IN ('sent', 'delivered', 'read')),
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (message_id, user_id),
			FOREIGN KEY (message_id) REFERENCES messages(id) ON DELETE CASCADE,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)
	`)
	if err != nil {
		return fmt.Errorf("error creating message_status table: %w", err)
	}

	// Message recipients table (for tracking intended recipients at send time)
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS message_recipients (
			message_id TEXT NOT NULL,
			user_id TEXT NOT NULL,
			PRIMARY KEY (message_id, user_id),
			FOREIGN KEY (message_id) REFERENCES messages(id) ON DELETE CASCADE,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)
	`)
	if err != nil {
		return fmt.Errorf("error creating message_recipients table: %w", err)
	}

	// Create indices for better performance
	indices := []string{
		"CREATE INDEX IF NOT EXISTS idx_users_name ON users(name)",
		"CREATE INDEX IF NOT EXISTS idx_conversations_type ON conversations(type)",
		"CREATE INDEX IF NOT EXISTS idx_messages_conversation_id ON messages(conversation_id)",
		"CREATE INDEX IF NOT EXISTS idx_messages_sender_id ON messages(sender_id)",
		"CREATE INDEX IF NOT EXISTS idx_messages_timestamp ON messages(timestamp DESC)",
		"CREATE INDEX IF NOT EXISTS idx_messages_deleted_at ON messages(deleted_at)",
		"CREATE INDEX IF NOT EXISTS idx_comments_message_id ON comments(message_id)",
		"CREATE INDEX IF NOT EXISTS idx_comments_user_id ON comments(user_id)",
		"CREATE INDEX IF NOT EXISTS idx_conversation_participants_user_id ON conversation_participants(user_id)",
		"CREATE INDEX IF NOT EXISTS idx_message_status_user_id ON message_status(user_id)",
		"CREATE INDEX IF NOT EXISTS idx_message_status_message_id ON message_status(message_id)",
		"CREATE INDEX IF NOT EXISTS idx_message_recipients_message_id ON message_recipients(message_id)",
		"CREATE INDEX IF NOT EXISTS idx_message_recipients_user_id ON message_recipients(user_id)",
	}

	for _, indexSQL := range indices {
		_, err = db.Exec(indexSQL)
		if err != nil {
			return fmt.Errorf("error creating index: %w", err)
		}
	}

	return nil
}
