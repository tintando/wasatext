package api

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/julienschmidt/httprouter"

	"wasatext/service/api/reqcontext"
	"wasatext/service/database"
	"wasatext/service/filestore"
)

type SendMessageRequest struct {
	Content          *string `json:"content,omitempty"`
	MessageType      string  `json:"messageType"`
	ReplyToMessageID *string `json:"replyToMessageId,omitempty"`
}

type ForwardMessageRequest struct {
	SourceConversationID string `json:"sourceConversationId"`
	SourceMessageID      string `json:"sourceMessageId"`
}

type UpdateMessageStatusRequest struct {
	Status string `json:"status"`
}

// sendMessage handles POST /conversations/{conversation-id}/messages.
//
// Photos arrive as multipart/form-data and need a file stored alongside the row,
// so they take a separate path; everything else is a plain text message whose
// fields come either from the form or from a JSON body.
func (rt *_router) sendMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	conversationID := ps.ByName("conversationId")
	if conversationID == "" {
		http.Error(w, "Conversation ID required", http.StatusBadRequest)
		return
	}

	isMultipart := strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data")

	var newMessage database.NewMessage
	switch {
	case isMultipart:
		if err := r.ParseMultipartForm(maxPhotoSize); err != nil {
			http.Error(w, "Error parsing multipart form", http.StatusBadRequest)
			return
		}

		messageType := r.FormValue("messageType")
		if messageType == "" {
			http.Error(w, "MessageType required in form", http.StatusBadRequest)
			return
		}

		if messageType == "photo" {
			rt.sendPhotoMessage(w, r, ctx, conversationID)
			return
		}

		content := r.FormValue("content")
		newMessage = database.NewMessage{
			Content:     &content,
			MessageType: messageType,
		}

	default:
		var req SendMessageRequest
		if !decodeJSONBody(w, r, &req) {
			return
		}

		if req.MessageType == "" {
			http.Error(w, "MessageType required", http.StatusBadRequest)
			return
		}

		newMessage = database.NewMessage{
			Content:          req.Content,
			MessageType:      req.MessageType,
			ReplyToMessageID: req.ReplyToMessageID,
		}
	}

	message, err := rt.db.SendMessage(conversationID, ctx.UserID, newMessage)
	if err != nil {
		writeSendMessageError(w, ctx, err, "Failed to send message")
		return
	}

	writeJSON(w, ctx, http.StatusCreated, message)
}

// sendPhotoMessage stores an uploaded photo and the message carrying it. The
// photo is written to disk first, so it is removed again if the message cannot
// be recorded.
func (rt *_router) sendPhotoMessage(w http.ResponseWriter, r *http.Request, ctx reqcontext.RequestContext, conversationID string) {
	file, _, err := r.FormFile("photo")
	if err != nil {
		http.Error(w, "Photo file required for photo message", http.StatusBadRequest)
		return
	}
	defer func() {
		if err := file.Close(); err != nil {
			ctx.Logger.WithError(err).Warn("failed to close uploaded photo")
		}
	}()

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(file); err != nil {
		http.Error(w, "Error reading photo file", http.StatusBadRequest)
		return
	}
	photoData := buf.Bytes()

	if len(photoData) == 0 {
		http.Error(w, "Empty photo file", http.StatusBadRequest)
		return
	}

	detectedType, err := filestore.ValidateImageFile(bytes.NewReader(photoData))
	if err != nil {
		http.Error(w, "Invalid image file", http.StatusBadRequest)
		return
	}

	fileStore := filestore.NewFileStore(rt.photos.MessagesPath)
	filename, err := fileStore.SavePhoto(bytes.NewReader(photoData), detectedType)
	if err != nil {
		http.Error(w, "Error saving photo", http.StatusInternalServerError)
		return
	}

	content := r.FormValue("content")
	var replyToPtr *string
	if replyTo := r.FormValue("replyToMessageId"); replyTo != "" {
		replyToPtr = &replyTo
	}

	message, err := rt.db.SendMessage(conversationID, ctx.UserID, database.NewMessage{
		Content:          &content,
		MessageType:      "photo",
		Photo:            photoData,
		ReplyToMessageID: replyToPtr,
	})
	if err != nil {
		discardPhoto(ctx, fileStore, filename)
		writeSendMessageError(w, ctx, err, "Failed to send message")
		return
	}

	// The stored filename can only be attached once the message has an ID.
	if err := rt.db.SetMessagePhoto(message.ID, filename); err != nil {
		discardPhoto(ctx, fileStore, filename)
		ctx.Logger.WithError(err).Error("failed to attach photo to message")
		http.Error(w, "Error updating message photo", http.StatusInternalServerError)
		return
	}

	// Clients read the photo back through the API, not from the stored filename.
	photoURL := fmt.Sprintf("/conversations/%s/messages/%s/photo", conversationID, message.ID)
	message.PhotoURL = &photoURL

	writeJSON(w, ctx, http.StatusCreated, message)
}

// deleteMessage handles DELETE /conversations/{conversation-id}/messages/{message-id}
func (rt *_router) deleteMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	conversationID := ps.ByName("conversationId")
	messageID := ps.ByName("messageId")
	if conversationID == "" || messageID == "" {
		http.Error(w, "Conversation ID and Message ID required", http.StatusBadRequest)
		return
	}

	if err := rt.db.DeleteMessage(messageID, ctx.UserID); err != nil {
		http.Error(w, "Failed to delete message", http.StatusForbidden)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// getMessagePhoto handles GET /conversations/{conversation-id}/messages/{message-id}/photo
func (rt *_router) getMessagePhoto(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	conversationID := ps.ByName("conversationId")
	messageID := ps.ByName("messageId")
	if conversationID == "" || messageID == "" {
		http.Error(w, "Conversation ID and Message ID required", http.StatusBadRequest)
		return
	}

	// Loading the conversation is also what enforces access: a non-participant
	// never gets far enough to name a message.
	conversation, err := rt.db.GetConversation(conversationID, ctx.UserID)
	if err != nil {
		writeConversationLookupError(w, err)
		return
	}

	var message *database.Message
	for i := range conversation.Messages {
		if conversation.Messages[i].ID == messageID {
			message = &conversation.Messages[i]
			break
		}
	}

	if message == nil {
		http.Error(w, "Message not found", http.StatusNotFound)
		return
	}

	if message.MessageType != "photo" {
		http.Error(w, "No photo available", http.StatusNotFound)
		return
	}

	servePhotoFile(w, r, rt.photos.MessagesPath, message.PhotoURL)
}

// forwardMessage handles POST /conversations/{conversation-id}/forwarded_messages
func (rt *_router) forwardMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	targetConversationID := ps.ByName("conversationId")
	if targetConversationID == "" {
		http.Error(w, "Conversation ID required", http.StatusBadRequest)
		return
	}

	var req ForwardMessageRequest
	if !decodeJSONBody(w, r, &req) {
		return
	}

	if req.SourceConversationID == "" || req.SourceMessageID == "" {
		http.Error(w, "sourceConversationId and sourceMessageId required", http.StatusBadRequest)
		return
	}

	message, err := rt.db.ForwardMessage(req.SourceMessageID, targetConversationID, ctx.UserID)
	if err != nil {
		writeSendMessageError(w, ctx, err, "Failed to forward message")
		return
	}

	writeJSON(w, ctx, http.StatusCreated, message)
}

// updateMessageStatus handles PUT /conversations/{conversation-id}/messages/{message-id}
func (rt *_router) updateMessageStatus(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	messageID := ps.ByName("messageId")
	if messageID == "" {
		http.Error(w, "Message ID required", http.StatusBadRequest)
		return
	}

	var req UpdateMessageStatusRequest
	if !decodeJSONBody(w, r, &req) {
		return
	}

	if req.Status != "delivered" && req.Status != "read" {
		http.Error(w, "Status must be 'delivered' or 'read'", http.StatusBadRequest)
		return
	}

	if err := rt.db.UpdateMessageStatus(messageID, ctx.UserID, req.Status); err != nil {
		if errors.Is(err, database.ErrNotRecipient) {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		ctx.Logger.WithError(err).Error("failed to update message status")
		http.Error(w, "Failed to update message status", http.StatusInternalServerError)
		return
	}

	writeMessage(w, ctx, http.StatusOK, "Message status updated successfully")
}

// writeSendMessageError maps a message write failure onto its HTTP status.
// failure is the message reported to the client and logged for the operator.
func writeSendMessageError(w http.ResponseWriter, ctx reqcontext.RequestContext, err error, failure string) {
	if errors.Is(err, database.ErrNotParticipant) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	ctx.Logger.WithError(err).Error(failure)
	http.Error(w, failure, http.StatusInternalServerError)
}
