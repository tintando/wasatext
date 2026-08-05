package api

import (
	"errors"
	"net/http"
	"strings"
)

// errNoAuth reports a missing or malformed Authorization header.
var errNoAuth = errors.New("missing or malformed Authorization header")

// getUserIDFromAuth extracts and validates the user ID from the Authorization
// header. The bearer token is the user ID itself, so the only check possible is
// that it names an existing user.
func (rt *_router) getUserIDFromAuth(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", errNoAuth
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" || parts[1] == "" {
		return "", errNoAuth
	}
	userID := parts[1]

	if _, err := rt.db.GetUserByID(userID); err != nil {
		return "", err
	}

	return userID, nil
}
