package api

import (
	"net/http"

	"github.com/julienschmidt/httprouter"

	"wasatext/service/api/reqcontext"
)

type CreateGroupRequest struct {
	Name      string   `json:"name"`
	MemberIds []string `json:"memberIds"`
}

type UpdateGroupRequest struct {
	Name string `json:"name"`
}

type AddMemberRequest struct {
	UserID string `json:"userId"`
}

// requireGroupMember replies with an HTTP error and returns false unless userID
// belongs to groupID. Group content is only visible to its members.
func (rt *_router) requireGroupMember(w http.ResponseWriter, groupID string, userID string) bool {
	members, err := rt.db.GetGroupMembers(groupID)
	if err != nil {
		http.Error(w, "Group not found", http.StatusNotFound)
		return false
	}

	for _, member := range members {
		if member.ID == userID {
			return true
		}
	}

	http.Error(w, "Forbidden", http.StatusForbidden)
	return false
}

// createGroup handles POST /groups
func (rt *_router) createGroup(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	var req CreateGroupRequest
	if !decodeJSONBody(w, r, &req) {
		return
	}

	if req.Name == "" {
		http.Error(w, "Group name required", http.StatusBadRequest)
		return
	}

	group, err := rt.db.CreateGroup(req.Name, ctx.UserID, req.MemberIds)
	if err != nil {
		ctx.Logger.WithError(err).Error("failed to create group")
		http.Error(w, "Failed to create group", http.StatusInternalServerError)
		return
	}

	writeJSON(w, ctx, http.StatusCreated, group)
}

// getGroup handles GET /groups/{group-id}
func (rt *_router) getGroup(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	groupID := ps.ByName("groupId")
	if groupID == "" {
		http.Error(w, "Group ID required", http.StatusBadRequest)
		return
	}

	group, err := rt.db.GetGroup(groupID)
	if err != nil {
		http.Error(w, "Group not found", http.StatusNotFound)
		return
	}

	writeJSON(w, ctx, http.StatusOK, group)
}

// setGroupName handles PUT /groups/{group-id}
func (rt *_router) setGroupName(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	groupID := ps.ByName("groupId")
	if groupID == "" {
		http.Error(w, "Group ID required", http.StatusBadRequest)
		return
	}

	var req UpdateGroupRequest
	if !decodeJSONBody(w, r, &req) {
		return
	}

	if req.Name == "" {
		http.Error(w, "Group name required", http.StatusBadRequest)
		return
	}

	if err := rt.db.UpdateGroupName(groupID, req.Name, ctx.UserID); err != nil {
		http.Error(w, "Failed to update group name", http.StatusForbidden)
		return
	}

	writeMessage(w, ctx, http.StatusOK, "Group name updated successfully")
}

// setGroupPhoto handles PUT /groups/{group-id}/photo
func (rt *_router) setGroupPhoto(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	groupID := ps.ByName("groupId")
	if groupID == "" {
		http.Error(w, "Group ID required", http.StatusBadRequest)
		return
	}

	if !rt.requireGroupMember(w, groupID, ctx.UserID) {
		return
	}

	fileStore, filename, ok := receivePhotoUpload(w, r, rt.photos.GroupsPath)
	if !ok {
		return
	}

	// The database stores the filename; the photo itself lives on disk.
	if err := rt.db.SetGroupPhoto(groupID, filename, ctx.UserID); err != nil {
		discardPhoto(ctx, fileStore, filename)
		ctx.Logger.WithError(err).Error("failed to update group photo")
		http.Error(w, "Error updating group photo", http.StatusInternalServerError)
		return
	}

	writeMessage(w, ctx, http.StatusOK, "Group photo updated successfully")
}

// getGroupPhoto handles GET /groups/{group-id}/photo
func (rt *_router) getGroupPhoto(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	groupID := ps.ByName("groupId")
	if groupID == "" {
		http.Error(w, "Group ID required", http.StatusBadRequest)
		return
	}

	if !rt.requireGroupMember(w, groupID, ctx.UserID) {
		return
	}

	group, err := rt.db.GetGroup(groupID)
	if err != nil {
		http.Error(w, "Group not found", http.StatusNotFound)
		return
	}

	servePhotoFile(w, r, rt.photos.GroupsPath, group.PhotoURL)
}

// getGroupMembers handles GET /groups/{group-id}/members
func (rt *_router) getGroupMembers(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	groupID := ps.ByName("groupId")
	if groupID == "" {
		http.Error(w, "Group ID required", http.StatusBadRequest)
		return
	}

	if !rt.requireGroupMember(w, groupID, ctx.UserID) {
		return
	}

	members, err := rt.db.GetGroupMembers(groupID)
	if err != nil {
		http.Error(w, "Group not found", http.StatusNotFound)
		return
	}

	writeJSON(w, ctx, http.StatusOK, map[string]interface{}{"members": members})
}

// addToGroup handles POST /groups/{group-id}/members
func (rt *_router) addToGroup(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	groupID := ps.ByName("groupId")
	if groupID == "" {
		http.Error(w, "Group ID required", http.StatusBadRequest)
		return
	}

	var req AddMemberRequest
	if !decodeJSONBody(w, r, &req) {
		return
	}

	if req.UserID == "" {
		http.Error(w, "UserID required", http.StatusBadRequest)
		return
	}

	if err := rt.db.AddUserToGroup(groupID, req.UserID, ctx.UserID); err != nil {
		http.Error(w, "Failed to add user to group", http.StatusForbidden)
		return
	}

	writeMessage(w, ctx, http.StatusCreated, "User added to group successfully")
}

// leaveGroup handles DELETE /groups/{group-id}/members/{user-id}
func (rt *_router) leaveGroup(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	groupID := ps.ByName("groupId")
	targetUserID := ps.ByName("userId")

	if groupID == "" || targetUserID == "" {
		http.Error(w, "Group ID and User ID required", http.StatusBadRequest)
		return
	}

	// This endpoint is "leave", not "kick": nobody can remove anyone else.
	if ctx.UserID != targetUserID {
		http.Error(w, "Users can only leave groups themselves", http.StatusForbidden)
		return
	}

	if err := rt.db.RemoveUserFromGroup(groupID, targetUserID); err != nil {
		http.Error(w, "Failed to leave group", http.StatusForbidden)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
