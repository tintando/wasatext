package api

import (
	"errors"
	"net/http"

	"github.com/julienschmidt/httprouter"

	"wasatext/service/api/reqcontext"
	"wasatext/service/database"
)

type UpdateUserRequest struct {
	Name string `json:"name"`
}

// searchUsers handles GET /users?q=query
func (rt *_router) searchUsers(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, "Query parameter 'q' required", http.StatusBadRequest)
		return
	}

	users, err := rt.db.SearchUsers(query)
	if err != nil {
		ctx.Logger.WithError(err).Error("user search failed")
		http.Error(w, "Search failed", http.StatusInternalServerError)
		return
	}

	writeJSON(w, ctx, http.StatusOK, map[string]interface{}{"users": users})
}

// getUserProfile handles GET /users/{user-id}
func (rt *_router) getUserProfile(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	targetUserID := ps.ByName("userId")
	if targetUserID == "" {
		http.Error(w, "User ID required", http.StatusBadRequest)
		return
	}

	user, err := rt.db.GetUserByID(targetUserID)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	writeJSON(w, ctx, http.StatusOK, user)
}

// setMyUserName handles PUT /users/{user-id}
func (rt *_router) setMyUserName(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	if ps.ByName("userId") != ctx.UserID {
		http.Error(w, "Can only update your own profile", http.StatusForbidden)
		return
	}

	var req UpdateUserRequest
	if !decodeJSONBody(w, r, &req) {
		return
	}

	if !isValidUsername(req.Name) {
		http.Error(w, usernameLengthMessage, http.StatusBadRequest)
		return
	}

	if err := rt.db.UpdateUserName(ctx.UserID, req.Name); err != nil {
		if errors.Is(err, database.ErrUsernameTaken) {
			http.Error(w, "Username already exists", http.StatusConflict)
			return
		}
		ctx.Logger.WithError(err).Error("failed to update username")
		http.Error(w, "Failed to update username", http.StatusInternalServerError)
		return
	}

	// Get updated user data to return
	user, err := rt.db.GetUserByID(ctx.UserID)
	if err != nil {
		http.Error(w, "Failed to retrieve updated user", http.StatusInternalServerError)
		return
	}

	writeJSON(w, ctx, http.StatusOK, user)
}

// setMyPhoto handles PUT /users/{user-id}/photo
func (rt *_router) setMyPhoto(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	if ps.ByName("userId") != ctx.UserID {
		http.Error(w, "Can only update your own photo", http.StatusForbidden)
		return
	}

	fileStore, filename, ok := receivePhotoUpload(w, r, rt.photos.UsersPath)
	if !ok {
		return
	}

	// The database stores the filename; the photo itself lives on disk.
	if err := rt.db.SetUserPhoto(ctx.UserID, filename); err != nil {
		discardPhoto(ctx, fileStore, filename)
		ctx.Logger.WithError(err).Error("failed to update user photo")
		http.Error(w, "Error updating user photo", http.StatusInternalServerError)
		return
	}

	writeMessage(w, ctx, http.StatusOK, "Photo updated successfully")
}

// getUserPhoto handles GET /users/{user-id}/photo
func (rt *_router) getUserPhoto(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	userID := ps.ByName("userId")
	if userID == "" {
		http.Error(w, "User ID required", http.StatusBadRequest)
		return
	}

	user, err := rt.db.GetUserByID(userID)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	servePhotoFile(w, r, rt.photos.UsersPath, user.PhotoURL)
}
