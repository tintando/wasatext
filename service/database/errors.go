package database

import "errors"

// Sentinel errors returned by the AppDatabase methods. Callers should match them
// with errors.Is rather than comparing error strings, so that the API layer can
// map a failure onto the right HTTP status without depending on wording.
var (
	// ErrUserNotFound is returned when a user does not exist.
	ErrUserNotFound = errors.New("user not found")

	// ErrUsernameTaken is returned when a username is already used by someone else.
	ErrUsernameTaken = errors.New("username already exists")

	// ErrNotParticipant is returned when a user acts on a conversation they do not belong to.
	ErrNotParticipant = errors.New("user is not a participant in this conversation")

	// ErrNotAuthorized is returned when a user may not access the requested conversation.
	ErrNotAuthorized = errors.New("user is not authorized to access this conversation")

	// ErrNotRecipient is returned when a user updates the status of a message not addressed to them.
	ErrNotRecipient = errors.New("user is not a recipient of this message")

	// ErrMessageNotFound is returned when a message does not exist or is not owned by the caller.
	ErrMessageNotFound = errors.New("message not found or unauthorized")

	// ErrCommentNotFound is returned when a comment does not exist or is not owned by the caller.
	ErrCommentNotFound = errors.New("comment not found")

	// ErrSelfReaction is returned when a user reacts to their own message.
	ErrSelfReaction = errors.New("cannot react to your own message")

	// ErrGroupNotFound is returned when a group does not exist.
	ErrGroupNotFound = errors.New("group not found")

	// ErrNotGroupMember is returned when a user acts on a group they do not belong to.
	ErrNotGroupMember = errors.New("user is not a member of this group")

	// ErrAlreadyGroupMember is returned when adding a user that is already in the group.
	ErrAlreadyGroupMember = errors.New("user is already in the group")
)
