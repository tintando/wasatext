package api

import (
	"net/http"

	"github.com/julienschmidt/httprouter"

	"wasatext/service/api/reqcontext"
)

// Username bounds, as defined by the OpenAPI spec.
const (
	minUsernameLength = 3
	maxUsernameLength = 16

	usernameLengthMessage = "Username must be between 3 and 16 characters"
)

// isValidUsername reports whether name is within the length the spec allows.
func isValidUsername(name string) bool {
	return len(name) >= minUsernameLength && len(name) <= maxUsernameLength
}

type LoginRequest struct {
	Name string `json:"name"`
}

type LoginResponse struct {
	Identifier string `json:"identifier"`
}

// doLogin handles POST /session - simplified login with username only.
// An unknown username creates the account, so there is no separate sign-up.
func (rt *_router) doLogin(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ctx reqcontext.RequestContext) {
	var req LoginRequest
	if !decodeJSONBody(w, r, &req) {
		return
	}

	if !isValidUsername(req.Name) {
		http.Error(w, usernameLengthMessage, http.StatusBadRequest)
		return
	}

	if existingUser, err := rt.db.GetUserByName(req.Name); err == nil {
		writeJSON(w, ctx, http.StatusCreated, LoginResponse{Identifier: existingUser.ID})
		return
	}

	newUser, err := rt.db.CreateUser(req.Name)
	if err != nil {
		ctx.Logger.WithError(err).Error("failed to create user")
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	writeJSON(w, ctx, http.StatusCreated, LoginResponse{Identifier: newUser.ID})
}
