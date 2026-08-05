<template>
  <div class="message-item mb-3" :class="{ 'message-sent': isOwn }">
    <div class="message-content">
      <div v-if="!isOwn" class="message-sender small text-muted mb-1">
        {{ message.senderName }}
      </div>

      <div class="message-bubble" :class="isOwn ? 'bg-primary text-white' : 'bg-light'">
        <div v-if="message.replyToMessageId" class="reply-indicator mb-2">
          <div class="reply-border">
            <div class="reply-content">
              <div class="reply-author">{{ replyAuthor }}</div>
              <div class="reply-text">{{ replyPreview }}</div>
            </div>
          </div>
        </div>

        <div v-if="message.forwardFromMessageId" class="forwarded-indicator mb-2">
          <div class="forwarded-border">
            <div class="forwarded-content">
              <div class="forwarded-icon">
                <svg width="14" height="14" fill="currentColor" viewBox="0 0 16 16" class="text-muted me-1">
                  <path d="M6.776 1.553a.5.5 0 0 1 .671.223l3 6a.5.5 0 0 1 0 .448l-3 6a.5.5 0 1 1-.894-.448L9.44 8 6.553 2.224a.5.5 0 0 1 .223-.671z" />
                  <path d="M2.176 1.553a.5.5 0 0 1 .671.223l3 6a.5.5 0 0 1 0 .448l-3 6a.5.5 0 1 1-.894-.448L4.84 8 1.953 2.224a.5.5 0 0 1 .223-.671z" />
                </svg>
                <small class="text-muted fw-bold">Forwarded</small>
              </div>
            </div>
          </div>
        </div>

        <div v-if="message.messageType === 'text'" class="message-text">
          {{ message.content }}
        </div>

        <div v-else-if="message.messageType === 'photo'" class="message-photo">
          <img
            v-if="message.photoObjectUrl"
            :src="message.photoObjectUrl"
            alt="Shared photo"
            class="img-fluid rounded mb-2 message-image"
          >
          <div v-else class="d-flex align-items-center justify-content-center bg-light rounded mb-2 photo-placeholder">
            <div class="text-center">
              <div class="spinner-border spinner-border-sm mb-2" />
              <div class="small text-muted">Loading image...</div>
            </div>
          </div>
          <div v-if="message.content && message.content.trim()" class="message-caption">
            {{ message.content }}
          </div>
        </div>
      </div>

      <div class="message-meta d-flex justify-content-between align-items-center mt-1">
        <small class="text-muted">
          {{ formatMessageTime(message.timestamp) }}
          <span v-if="isOwn" class="ms-1">
            <span v-if="message.status === 'read'">✓✓</span>
            <span v-else-if="message.status === 'delivered'">✓</span>
          </span>
        </small>

        <div class="message-actions">
          <div class="dropdown">
            <button class="btn btn-sm btn-link text-muted p-0" data-bs-toggle="dropdown">
              ⋯
            </button>
            <ul class="dropdown-menu dropdown-menu-end">
              <li><button class="dropdown-item" type="button" @click.prevent="$emit('reply', message)">Reply</button></li>
              <li v-if="!isOwn"><button class="dropdown-item" type="button" @click.prevent="$emit('react', message)">React</button></li>
              <li><button class="dropdown-item" type="button" @click.prevent="$emit('forward', message)">Forward</button></li>
              <li v-if="isOwn">
                <button class="dropdown-item text-danger" type="button" @click.prevent="$emit('delete', message)">Delete</button>
              </li>
            </ul>
          </div>
        </div>
      </div>

      <div v-if="message.comments && message.comments.length > 0" class="message-comments mt-2">
        <div class="comments-list">
          <span
            v-for="comment in message.comments"
            :key="comment.id"
            class="comment-item me-2"
            :class="{ 'user-reaction': comment.userId === currentUserId }"
            :title="comment.userName"
            @click="$emit('toggle-reaction', message, comment.emoticon)"
          >
            {{ comment.emoticon }}
          </span>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { formatMessageTime } from '../utils/format.js';

// Longest reply preview shown inside a bubble before it is cut short.
const REPLY_PREVIEW_LENGTH = 50;

// MessageBubble renders one message: its reply and forward banners, its text
// or photo body, the delivery ticks, the action menu and the reactions.
export default {
    name: 'MessageBubble',
    props: {
        message: {
            type: Object,
            required: true
        },
        currentUserId: {
            type: String,
            default: null
        },
        // The message this one replies to, already resolved by the parent, or
        // null when the original is not in the loaded history.
        replyTo: {
            type: Object,
            default: null
        }
    },
    emits: ['reply', 'react', 'forward', 'delete', 'toggle-reaction'],
    computed: {
        isOwn() {
            return this.message.senderId === this.currentUserId;
        },

        replyAuthor() {
            return this.replyTo?.senderName || 'Unknown';
        },

        replyPreview() {
            if (!this.replyTo) return 'Message not found';

            if (this.replyTo.messageType === 'photo') {
                return '📷 Photo' + (this.replyTo.content ? ': ' + this.replyTo.content : '');
            }

            if (this.replyTo.messageType === 'text') {
                const content = this.replyTo.content || '';
                return content.length > REPLY_PREVIEW_LENGTH
                    ? content.substring(0, REPLY_PREVIEW_LENGTH) + '...'
                    : content;
            }

            return 'Message';
        }
    },
    methods: {
        formatMessageTime
    }
};
</script>

<style scoped>
.message-item {
    max-width: 70%;
}

.message-sent {
    margin-left: auto;
    text-align: right;
}

.message-bubble {
    padding: 10px 15px;
    border-radius: 18px;
    word-wrap: break-word;
}

.message-image {
    max-width: 300px;
    max-height: 300px;
}

.photo-placeholder {
    width: 200px;
    height: 150px;
}

.reply-indicator,
.forwarded-indicator {
    font-size: 0.875em;
}

.reply-border {
    border-left: 3px solid rgba(255, 255, 255, 0.5);
    background-color: rgba(255, 255, 255, 0.1);
    border-radius: 4px;
    padding: 6px 8px;
}

.message-bubble.bg-light .reply-border {
    border-left-color: #007bff;
    background-color: rgba(0, 123, 255, 0.1);
}

.reply-author {
    font-weight: 600;
    color: inherit;
    margin-bottom: 2px;
    opacity: 0.9;
}

.reply-text {
    opacity: 0.8;
    font-style: italic;
}

.forwarded-border {
    border-left: 3px solid rgba(255, 255, 255, 0.5);
    background-color: rgba(255, 255, 255, 0.1);
    border-radius: 4px;
    padding: 4px 8px;
}

.message-bubble.bg-light .forwarded-border {
    border-left-color: #28a745;
    background-color: rgba(40, 167, 69, 0.1);
}

.forwarded-content,
.forwarded-icon {
    display: flex;
    align-items: center;
}

.comment-item {
    font-size: 1.2em;
    cursor: pointer;
    padding: 2px 4px;
    border-radius: 4px;
    transition: background-color 0.2s ease;
}

.comment-item:hover {
    background-color: rgba(0, 0, 0, 0.1);
}

.user-reaction {
    background-color: #007bff;
    color: white;
    font-weight: bold;
}

.user-reaction:hover {
    background-color: #0056b3;
}
</style>
