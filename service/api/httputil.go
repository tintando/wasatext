package api

import (
	"bytes"
	"encoding/json"
	"net/http"

	"wasatext/service/api/reqcontext"
)

// writeJSON sends v as a JSON response body with the given status code.
//
// The value is encoded into a buffer before anything is written, so that an
// encoding failure can still be reported as a 500: once the status line is on
// the wire it can no longer be changed.
func writeJSON(w http.ResponseWriter, ctx reqcontext.RequestContext, status int, v interface{}) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(v); err != nil {
		ctx.Logger.WithError(err).Error("error encoding response")
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if _, err := w.Write(buf.Bytes()); err != nil {
		ctx.Logger.WithError(err).Warn("error writing response body")
	}
}

// writeMessage sends a `{"message": ...}` JSON acknowledgement.
func writeMessage(w http.ResponseWriter, ctx reqcontext.RequestContext, status int, message string) {
	writeJSON(w, ctx, status, map[string]string{"message": message})
}

// decodeJSONBody parses the request body into dst, replying with 400 and
// returning false when the body is not valid JSON.
func decodeJSONBody(w http.ResponseWriter, r *http.Request, dst interface{}) bool {
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return false
	}
	return true
}
