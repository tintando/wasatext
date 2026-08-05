<template>
  <BaseModal title="Forward Message" @close="$emit('close')">
    <ul class="nav nav-tabs mb-3">
      <li class="nav-item">
        <button
          class="nav-link"
          :class="{ active: tab === 'conversations' }"
          @click="tab = 'conversations'"
        >
          Conversations
        </button>
      </li>
      <li class="nav-item">
        <button
          class="nav-link"
          :class="{ active: tab === 'users' }"
          @click="tab = 'users'"
        >
          Users
        </button>
      </li>
    </ul>

    <ErrorMsg v-if="errorMsg" :msg="errorMsg" class="mb-3" />

    <div v-if="tab === 'conversations'">
      <p class="small text-muted mb-3">Select a conversation to forward this message to:</p>
      <div v-if="conversations.length > 0">
        <div
          v-for="conversation in conversations"
          :key="conversation.id"
          class="d-flex align-items-center p-2 border rounded mb-2 conversation-option"
          @click="forwardToConversation(conversation)"
        >
          <UserAvatar
            :user-id="conversation.type === 'direct' ? conversation.otherUserId : null"
            :group-id="conversation.type === 'group' ? conversation.id : null"
            :avatar-type="conversation.type === 'direct' ? 'user' : 'group'"
            :photo-url="conversation.photoUrl"
            :name="conversation.name"
            size="medium"
            class="me-3"
          />
          <div>{{ conversation.name }}</div>
        </div>
      </div>
      <div v-else class="text-center text-muted py-3">
        No other conversations available
      </div>
    </div>

    <div v-else>
      <p class="small text-muted mb-3">Search for a user to forward this message to:</p>
      <div class="mb-3">
        <input
          v-model="userQuery"
          type="text"
          class="form-control"
          placeholder="Search users..."
          @input="searchUsers"
        >
      </div>
      <div v-if="users.length > 0">
        <div
          v-for="user in users"
          :key="user.id"
          class="d-flex align-items-center p-2 border rounded mb-2 conversation-option"
          @click="forwardToUser(user)"
        >
          <UserAvatar
            :user-id="user.id"
            :photo-url="user.photoUrl"
            :name="user.name"
            size="medium"
            class="me-3"
          />
          <div>{{ user.name }}</div>
        </div>
      </div>
      <div v-else-if="userQuery.length > 0" class="text-center text-muted py-3">
        No users found
      </div>
      <div v-else class="text-center text-muted py-3">
        Start typing to search for users
      </div>
    </div>
  </BaseModal>
</template>

<script>
import conversationService from '../services/conversationService.js';
import userService from '../services/userService.js';
import BaseModal from './BaseModal.vue';
import ErrorMsg from './ErrorMsg.vue';
import UserAvatar from './UserAvatar.vue';

// ForwardMessageModal picks where to forward a message: either an existing
// conversation, or a user, in which case the direct conversation with them is
// created on the way.
export default {
    name: 'ForwardMessageModal',
    components: {
        BaseModal,
        ErrorMsg,
        UserAvatar
    },
    props: {
        message: {
            type: Object,
            required: true
        },
        sourceConversationId: {
            type: String,
            required: true
        },
        currentUserId: {
            type: String,
            default: null
        }
    },
    emits: ['close', 'forwarded'],
    data() {
        return {
            tab: 'conversations',
            conversations: [],
            users: [],
            userQuery: '',
            errorMsg: null
        };
    },
    async mounted() {
        await this.loadConversations();
    },
    methods: {
        async loadConversations() {
            try {
                const all = await conversationService.getMyConversations();
                this.conversations = all.filter(conversation => conversation.id !== this.sourceConversationId);
            } catch (error) {
                this.errorMsg = 'Failed to load conversations';
                this.conversations = [];
            }
        },

        async searchUsers() {
            const query = this.userQuery.trim();
            if (query.length < 1) {
                this.users = [];
                return;
            }

            try {
                const results = await userService.searchUsers(query);
                this.users = results.filter(user => user.id !== this.currentUserId);
            } catch (error) {
                this.errorMsg = 'Failed to search users';
                this.users = [];
            }
        },

        async forwardToConversation(conversation) {
            try {
                await conversationService.forwardMessage(
                    this.sourceConversationId,
                    this.message.id,
                    conversation.id
                );
                this.$emit('forwarded', conversation.name);
            } catch (error) {
                this.errorMsg = 'Failed to forward message';
            }
        },

        async forwardToUser(user) {
            try {
                await conversationService.forwardMessageToUser(
                    this.sourceConversationId,
                    this.message.id,
                    user.id
                );
                this.$emit('forwarded', user.name);
            } catch (error) {
                this.errorMsg = 'Failed to forward message to user';
            }
        }
    }
};
</script>

<style scoped>
.conversation-option {
    cursor: pointer;
}

.conversation-option:hover {
    background-color: #f8f9fa !important;
}
</style>
