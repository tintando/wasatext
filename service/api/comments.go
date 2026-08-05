package api

import (
	"net/http"

	"github.com/julienschmidt/httprouter"

	"wasatext/service/api/reqcontext"
)

type AddCommentRequest struct {
	Emoticon string `json:"emoticon"`
}

// commentMessage handles POST /conversations/{conversation-id}/messages/{message-id}/comments
func (rt *_router) commentMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	conversationID := ps.ByName("conversationId")
	messageID := ps.ByName("messageId")
	if conversationID == "" || messageID == "" {
		http.Error(w, "Conversation ID and Message ID required", http.StatusBadRequest)
		return
	}

	var req AddCommentRequest
	if !decodeJSONBody(w, r, &req) {
		return
	}

	if req.Emoticon == "" {
		http.Error(w, "Emoticon required", http.StatusBadRequest)
		return
	}

	comment, err := rt.db.AddComment(messageID, ctx.UserID, req.Emoticon)
	if err != nil {
		ctx.Logger.WithError(err).Error("failed to add comment")
		http.Error(w, "Failed to add comment", http.StatusInternalServerError)
		return
	}

	writeJSON(w, ctx, http.StatusCreated, comment)
}

// uncommentMessage handles DELETE /conversations/{conversation-id}/messages/{message-id}/comments/{comment-id}
func (rt *_router) uncommentMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	conversationID := ps.ByName("conversationId")
	messageID := ps.ByName("messageId")
	commentID := ps.ByName("commentId")
	if conversationID == "" || messageID == "" || commentID == "" {
		http.Error(w, "Conversation ID, Message ID, and Comment ID required", http.StatusBadRequest)
		return
	}

	if err := rt.db.DeleteComment(commentID, ctx.UserID); err != nil {
		http.Error(w, "Failed to delete comment", http.StatusForbidden)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// removeUserReaction handles DELETE /conversations/{conversation-id}/messages/{message-id}/user-reaction
func (rt *_router) removeUserReaction(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	conversationID := ps.ByName("conversationId")
	messageID := ps.ByName("messageId")
	if conversationID == "" || messageID == "" {
		http.Error(w, "Conversation ID and Message ID required", http.StatusBadRequest)
		return
	}

	if err := rt.db.DeleteCommentByMessageAndUser(messageID, ctx.UserID); err != nil {
		http.Error(w, "Failed to delete reaction", http.StatusForbidden)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
