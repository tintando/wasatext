package api

import (
	"errors"
	"net/http"

	"github.com/julienschmidt/httprouter"

	"wasatext/service/api/reqcontext"
	"wasatext/service/database"
)

type CreateConversationRequest struct {
	UserID string `json:"userId"`
}

type UpdateConversationRequest struct {
	Name string `json:"name"`
}

// createDirectConversation handles POST /conversations
func (rt *_router) createDirectConversation(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	var req CreateConversationRequest
	if !decodeJSONBody(w, r, &req) {
		return
	}

	if req.UserID == "" {
		http.Error(w, "UserID required", http.StatusBadRequest)
		return
	}

	if req.UserID == ctx.UserID {
		http.Error(w, "Cannot create conversation with yourself", http.StatusBadRequest)
		return
	}

	conversationID, err := rt.db.GetOrCreateDirectConversation(ctx.UserID, req.UserID)
	if err != nil {
		ctx.Logger.WithError(err).Error("failed to create conversation")
		http.Error(w, "Failed to create conversation", http.StatusInternalServerError)
		return
	}

	// Get the full conversation object to return
	conversation, err := rt.db.GetConversation(conversationID, ctx.UserID)
	if err != nil {
		http.Error(w, "Failed to retrieve conversation details", http.StatusInternalServerError)
		return
	}

	writeJSON(w, ctx, http.StatusCreated, conversation)
}

// getMyConversations handles GET /conversations
func (rt *_router) getMyConversations(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	// Opening the conversation list is what marks incoming messages as delivered.
	if err := rt.db.MarkMessagesAsDelivered(ctx.UserID); err != nil {
		ctx.Logger.WithError(err).Warn("failed to mark messages as delivered")
	}

	conversations, err := rt.db.GetUserConversations(ctx.UserID)
	if err != nil {
		ctx.Logger.WithError(err).Error("failed to get conversations")
		http.Error(w, "Failed to get conversations", http.StatusInternalServerError)
		return
	}

	writeJSON(w, ctx, http.StatusOK, conversations)
}

// getConversation handles GET /conversations/{conversation-id}
func (rt *_router) getConversation(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	conversationID := ps.ByName("conversationId")
	if conversationID == "" {
		http.Error(w, "Conversation ID required", http.StatusBadRequest)
		return
	}

	// Opening a conversation marks its messages read, and clears the unread
	// counter shown in the conversation list. Neither is worth failing the read.
	if err := rt.db.MarkConversationMessagesAsRead(conversationID, ctx.UserID); err != nil {
		ctx.Logger.WithError(err).Warn("failed to mark conversation messages as read")
	}
	if err := rt.db.MarkMessagesAsRead(conversationID, ctx.UserID); err != nil {
		ctx.Logger.WithError(err).Warn("failed to update last read time")
	}

	conversation, err := rt.db.GetConversation(conversationID, ctx.UserID)
	if err != nil {
		writeConversationLookupError(w, err)
		return
	}

	writeJSON(w, ctx, http.StatusOK, conversation)
}

// updateConversation handles PUT /conversations/{conversation-id}
func (rt *_router) updateConversation(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	conversationID := ps.ByName("conversationId")
	if conversationID == "" {
		http.Error(w, "Conversation ID required", http.StatusBadRequest)
		return
	}

	var req UpdateConversationRequest
	if !decodeJSONBody(w, r, &req) {
		return
	}

	if req.Name == "" {
		http.Error(w, "Name required", http.StatusBadRequest)
		return
	}

	// Only groups carry a name, so this is a group rename.
	if err := rt.db.UpdateGroupName(conversationID, req.Name, ctx.UserID); err != nil {
		http.Error(w, "Failed to update conversation name", http.StatusForbidden)
		return
	}

	conversation, err := rt.db.GetConversation(conversationID, ctx.UserID)
	if err != nil {
		http.Error(w, "Failed to retrieve updated conversation", http.StatusInternalServerError)
		return
	}

	writeJSON(w, ctx, http.StatusOK, conversation)
}

// getConversationMessages handles GET /conversations/{conversation-id}/messages
func (rt *_router) getConversationMessages(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	conversationID := ps.ByName("conversationId")
	if conversationID == "" {
		http.Error(w, "Conversation ID required", http.StatusBadRequest)
		return
	}

	conversation, err := rt.db.GetConversation(conversationID, ctx.UserID)
	if err != nil {
		writeConversationLookupError(w, err)
		return
	}

	writeJSON(w, ctx, http.StatusOK, map[string]interface{}{"messages": conversation.Messages})
}

// writeConversationLookupError maps a GetConversation failure onto its HTTP status.
func writeConversationLookupError(w http.ResponseWriter, err error) {
	if errors.Is(err, database.ErrNotAuthorized) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	http.Error(w, "Conversation not found", http.StatusNotFound)
}
