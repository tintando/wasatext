<template>
  <div ref="container" class="messages-container flex-grow-1 overflow-auto p-3">
    <LoadingSpinner :loading="loading && messages.length === 0">
      <div v-if="messages.length === 0" class="text-center py-5">
        <p class="text-muted">No messages yet. Start the conversation!</p>
      </div>

      <div v-else class="messages-list">
        <MessageBubble
          v-for="message in messages"
          :key="message.id"
          :message="message"
          :current-user-id="currentUserId"
          :reply-to="replyTargets[message.id] || null"
          @reply="$emit('reply', $event)"
          @react="$emit('react', $event)"
          @forward="$emit('forward', $event)"
          @delete="$emit('delete', $event)"
          @toggle-reaction="(msg, emoticon) => $emit('toggle-reaction', msg, emoticon)"
        />
      </div>
    </LoadingSpinner>
  </div>
</template>

<script>
import LoadingSpinner from './LoadingSpinner.vue';
import MessageBubble from './MessageBubble.vue';

// How close to the bottom still counts as "scrolled to the bottom", in pixels.
const AT_BOTTOM_SLACK = 10;

// MessageList is the scrolling message history. It owns the scroll container,
// so callers ask it whether the user is at the bottom rather than reaching for
// the DOM themselves.
export default {
    name: 'MessageList',
    components: {
        LoadingSpinner,
        MessageBubble
    },
    props: {
        messages: {
            type: Array,
            required: true
        },
        currentUserId: {
            type: String,
            default: null
        },
        loading: {
            type: Boolean,
            default: false
        }
    },
    emits: ['reply', 'react', 'forward', 'delete', 'toggle-reaction'],
    computed: {
        // replyTargets maps a message ID to the message it replies to, so each
        // bubble gets its original without searching the history itself.
        replyTargets() {
            const byId = new Map(this.messages.map(message => [message.id, message]));

            return this.messages.reduce((targets, message) => {
                if (message.replyToMessageId) {
                    targets[message.id] = byId.get(message.replyToMessageId) || null;
                }
                return targets;
            }, {});
        }
    },
    methods: {
        isAtBottom() {
            const container = this.$refs.container;
            if (!container) return true;

            return container.scrollTop + container.clientHeight >= container.scrollHeight - AT_BOTTOM_SLACK;
        },

        scrollToBottom() {
            this.$nextTick(() => {
                const container = this.$refs.container;
                if (container) {
                    container.scrollTop = container.scrollHeight;
                }
            });
        }
    }
};
</script>

<style scoped>
.messages-container {
    background-color: #f8f9fa;
}
</style>
