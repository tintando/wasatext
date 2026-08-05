package api

import (
	"bytes"
	"io"
	"net/http"
	"os"

	"wasatext/service/api/reqcontext"
	"wasatext/service/filestore"
)

// maxPhotoSize is the largest photo upload accepted, in bytes.
const maxPhotoSize = 10 * 1024 * 1024

// receivePhotoUpload validates and stores an image sent as the raw request body.
// On failure it writes the HTTP error itself and returns ok=false; on success it
// returns the store the photo landed in, so the caller can discard it if a later
// step fails.
func receivePhotoUpload(w http.ResponseWriter, r *http.Request, basePath string) (*filestore.FileStore, string, bool) {
	contentType := r.Header.Get("Content-Type")
	if contentType == "" {
		http.Error(w, "Content-Type header required", http.StatusBadRequest)
		return nil, "", false
	}

	if !filestore.IsValidImageType(contentType) {
		http.Error(w, "Invalid image type. Supported: JPEG, PNG, GIF", http.StatusUnsupportedMediaType)
		return nil, "", false
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxPhotoSize)
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading file", http.StatusBadRequest)
		return nil, "", false
	}

	if len(body) == 0 {
		http.Error(w, "Empty file", http.StatusBadRequest)
		return nil, "", false
	}

	// The sniffed type wins over the declared one, so a mislabelled body cannot
	// get a wrong extension on disk.
	detectedType, err := filestore.ValidateImageFile(bytes.NewReader(body))
	if err != nil {
		http.Error(w, "Invalid image file", http.StatusBadRequest)
		return nil, "", false
	}

	fileStore := filestore.NewFileStore(basePath)
	filename, err := fileStore.SavePhoto(bytes.NewReader(body), detectedType)
	if err != nil {
		http.Error(w, "Error saving photo", http.StatusInternalServerError)
		return nil, "", false
	}

	return fileStore, filename, true
}

// discardPhoto removes an already stored photo after a later step failed, so a
// failed upload does not leave an orphan file behind.
func discardPhoto(ctx reqcontext.RequestContext, fileStore *filestore.FileStore, filename string) {
	if err := fileStore.DeletePhoto(filename); err != nil {
		ctx.Logger.WithError(err).Warn("failed to delete photo file during cleanup")
	}
}

// servePhotoFile serves a stored photo, replying 404 when no photo is recorded
// or the file is missing from disk.
func servePhotoFile(w http.ResponseWriter, r *http.Request, basePath string, storedName *string) {
	if storedName == nil || *storedName == "" {
		http.Error(w, "No photo available", http.StatusNotFound)
		return
	}

	filePath := filestore.NewFileStore(basePath).GetPhotoPath(*storedName)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		http.Error(w, "Photo file not found", http.StatusNotFound)
		return
	}

	http.ServeFile(w, r, filePath)
}
