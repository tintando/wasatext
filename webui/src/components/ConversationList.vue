<template>
  <div class="conversations-panel bg-white border-end">
    <div class="p-3 border-bottom">
      <div class="d-flex justify-content-between align-items-center mb-2">
        <h5 class="mb-0">Conversations</h5>
        <div class="dropdown">
          <button
            class="btn btn-sm btn-outline-secondary dropdown-toggle"
            type="button"
            data-bs-toggle="dropdown"
          >
            {{ currentUser?.name || 'User' }}
          </button>
          <ul class="dropdown-menu dropdown-menu-end">
            <li><button class="dropdown-item" type="button" @click="$emit('show-profile')">Profile</button></li>
            <li><hr class="dropdown-divider"></li>
            <li><button class="dropdown-item" type="button" @click="$emit('logout')">Sign Out</button></li>
          </ul>
        </div>
      </div>

      <div class="d-flex gap-2">
        <button
          type="button"
          class="btn btn-sm btn-outline-secondary flex-fill"
          :disabled="loading"
          @click="$emit('refresh')"
        >
          <span v-if="loading" class="spinner-border spinner-border-sm me-1" />
          Refresh
        </button>
        <div class="btn-group">
          <button
            type="button"
            class="btn btn-sm btn-primary dropdown-toggle"
            data-bs-toggle="dropdown"
          >
            New
          </button>
          <ul class="dropdown-menu">
            <li><button class="dropdown-item" type="button" @click="$emit('new-direct')">Direct Message</button></li>
            <li><button class="dropdown-item" type="button" @click="$emit('new-group')">Create Group</button></li>
          </ul>
        </div>
      </div>
    </div>

    <div class="conversations-list flex-grow-1 overflow-auto">
      <ErrorMsg v-if="errorMsg" :msg="errorMsg" class="m-3" />

      <LoadingSpinner :loading="loading && conversations.length === 0">
        <div v-if="conversations.length === 0" class="text-center py-5 px-3">
          <h6 class="text-muted">No conversations yet</h6>
          <p class="text-muted small">Start a new conversation by clicking "New"</p>
        </div>

        <div v-else>
          <div
            v-for="conversation in conversations"
            :key="conversation.id"
            class="conversation-item d-flex align-items-center p-3 border-bottom"
            :class="{ active: selectedConversationId === conversation.id }"
            @click="$emit('select', conversation)"
          >
            <UserAvatar
              :key="avatarKey(conversation)"
              :user-id="conversation.type === 'direct' ? conversation.otherUserId : null"
              :group-id="conversation.type === 'group' ? conversation.id : null"
              :avatar-type="conversation.type === 'direct' ? 'user' : 'group'"
              :photo-url="conversation.photoUrl"
              :name="conversation.name"
              size="medium"
              class="me-3"
            />
            <div class="flex-grow-1 min-width-0">
              <div class="d-flex justify-content-between align-items-start">
                <h6 class="mb-1 text-truncate">{{ conversation.name }}</h6>
                <small class="text-muted flex-shrink-0 ms-2">
                  {{ formatTime(conversation.lastMessage?.timestamp) }}
                </small>
              </div>
              <div class="d-flex justify-content-between align-items-center">
                <small class="text-muted text-truncate">
                  {{ conversation.lastMessage?.content || 'No messages yet' }}
                </small>
                <span
                  v-if="conversation.unreadCount > 0"
                  class="badge bg-primary rounded-pill flex-shrink-0 ms-2"
                >
                  {{ conversation.unreadCount }}
                </span>
              </div>
            </div>
          </div>
        </div>
      </LoadingSpinner>
    </div>
  </div>
</template>

<script>
import ErrorMsg from './ErrorMsg.vue';
import LoadingSpinner from './LoadingSpinner.vue';
import UserAvatar from './UserAvatar.vue';
import { formatTime } from '../utils/format.js';

// ConversationList is the left panel: the signed-in user's menu, the
// refresh/new controls, and the list of conversations to pick from.
export default {
    name: 'ConversationList',
    components: {
        ErrorMsg,
        LoadingSpinner,
        UserAvatar
    },
    props: {
        conversations: {
            type: Array,
            required: true
        },
        loading: {
            type: Boolean,
            default: false
        },
        errorMsg: {
            type: String,
            default: null
        },
        selectedConversationId: {
            type: String,
            default: null
        },
        currentUser: {
            type: Object,
            default: null
        }
    },
    emits: ['refresh', 'select', 'logout', 'show-profile', 'new-direct', 'new-group'],
    methods: {
        formatTime,

        // avatarKey forces a fresh avatar whenever the conversation's photo
        // changes, since UserAvatar fetches the image once per instance.
        avatarKey(conversation) {
            return `${conversation.type}-${conversation.id}-${conversation.photoUrl || 'no-photo'}`;
        }
    }
};
</script>

<style scoped>
.conversations-panel {
    width: 350px;
    min-width: 300px;
    display: flex;
    flex-direction: column;
}

.conversations-list {
    height: calc(100vh - 200px);
}

.conversation-item {
    cursor: pointer;
}

.conversation-item:hover {
    background-color: #f8f9fa;
}

.conversation-item.active {
    background-color: #e3f2fd;
    border-left: 4px solid #007bff;
}

.min-width-0 {
    min-width: 0;
}
</style>
