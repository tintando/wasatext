// Formatting helpers shared by the conversation list and the message view.

// formatTime renders a conversation-list timestamp: the clock for today,
// the date for anything older.
export function formatTime(timestamp) {
    if (!timestamp) return '';

    const date = new Date(timestamp);
    const now = new Date();

    if (date.toDateString() === now.toDateString()) {
        return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
    }
    return date.toLocaleDateString();
}

// formatMessageTime renders the clock shown under a single message.
export function formatMessageTime(timestamp) {
    if (!timestamp) return '';

    return new Date(timestamp).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
}
