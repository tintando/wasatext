// Client-side checks for photo uploads. The backend validates uploads again by
// sniffing the file, so these only exist to fail fast with a clear message.

export const ALLOWED_PHOTO_TYPES = ['image/png', 'image/jpeg', 'image/gif'];

export const MAX_PHOTO_SIZE = 5 * 1024 * 1024;

// validatePhotoFile returns an error message for an unusable file, or null when
// the file is acceptable.
export function validatePhotoFile(file) {
    if (!ALLOWED_PHOTO_TYPES.includes(file.type)) {
        return 'Please select a PNG, JPEG, or GIF image';
    }

    if (file.size > MAX_PHOTO_SIZE) {
        return 'Image must be smaller than 5MB';
    }

    return null;
}
