<template>
  <div class="chat-container d-flex">
    <ConversationList
      :conversations="conversations"
      :loading="loading"
      :error-msg="errorMsg"
      :selected-conversation-id="selectedConversationId"
      :current-user="currentUser"
      @refresh="refreshConversations"
      @select="selectConversation"
      @logout="handleLogout"
      @show-profile="showProfile = true"
      @new-direct="showUserSearch = true"
      @new-group="showGroupCreation = true"
    />

    <div class="conversation-panel flex-grow-1 d-flex flex-column">
      <div v-if="!selectedConversationId" class="d-flex align-items-center justify-content-center h-100 bg-light">
        <div class="text-center">
          <h4 class="text-muted mb-3">Select a conversation</h4>
          <p class="text-muted">Choose a conversation from the left to start messaging</p>
        </div>
      </div>

      <div v-else class="d-flex flex-column h-100">
        <div class="conversation-header bg-white border-bottom p-3 d-flex align-items-center">
          <UserAvatar
            :key="`header-${selectedConversation?.type}-${selectedConversation?.id}-${selectedConversation?.photoUrl || 'no-photo'}`"
            :user-id="selectedConversation?.type === 'direct' ? selectedConversation?.otherUserId : null"
            :group-id="selectedConversation?.type === 'group' ? selectedConversation?.id : null"
            :avatar-type="selectedConversation?.type === 'direct' ? 'user' : 'group'"
            :photo-url="selectedConversation?.photoUrl"
            :name="conversationName || 'Conversation'"
            size="medium"
            class="me-3"
          />
          <div class="flex-grow-1">
            <h5 class="mb-0">{{ conversationName }}</h5>
            <small class="text-muted">{{ conversationInfo }}</small>
          </div>
          <div class="dropdown">
            <button class="btn btn-outline-secondary btn-sm dropdown-toggle" type="button" data-bs-toggle="dropdown">
              Options
            </button>
            <ul class="dropdown-menu dropdown-menu-end">
              <li v-if="selectedConversation?.type === 'group'">
                <button class="dropdown-item" type="button" @click.prevent="showGroupSettings = true">Group Settings</button>
              </li>
              <li><button class="dropdown-item" type="button" @click.prevent="refreshMessages">Refresh</button></li>
            </ul>
          </div>
        </div>

        <MessageList
          ref="messageList"
          :messages="messages"
          :current-user-id="currentUser?.id"
          :loading="messagesLoading"
          @reply="replyingTo = $event"
          @react="openEmojiPicker"
          @forward="forwardingMessage = $event"
          @delete="deleteMessage"
          @toggle-reaction="toggleReaction"
        />

        <MessageComposer
          ref="composer"
          :sending="sending"
          :replying-to="replyingTo"
          @send="sendMessage"
          @cancel-reply="replyingTo = null"
        />
      </div>
    </div>

    <UserSearchModal
      v-if="showUserSearch"
      :current-user-id="currentUser?.id"
      @close="showUserSearch = false"
      @select="startConversation"
    />

    <GroupCreationModal
      v-if="showGroupCreation"
      @close="showGroupCreation = false"
      @created="refreshConversations"
    />

    <ProfileModal
      v-if="showProfile"
      :user="currentUser"
      @close="showProfile = false"
      @renamed="currentUser = $event"
    />

    <ForwardMessageModal
      v-if="forwardingMessage"
      :message="forwardingMessage"
      :source-conversation-id="selectedConversationId"
      :current-user-id="currentUser?.id"
      @close="forwardingMessage = null"
      @forwarded="onMessageForwarded"
    />

    <GroupSettingsModal
      v-if="showGroupSettings && selectedConversation"
      :conversation="selectedConversation"
      :current-user="currentUser"
      @close="showGroupSettings = false"
      @updated="onGroupUpdated"
      @left="onGroupLeft"
    />

    <EmojiPickerModal
      v-if="reactingToMessage"
      @close="reactingToMessage = null"
      @select="addReaction"
    />
  </div>
</template>

<script>
import authService from '../services/authService.js';
import conversationService from '../services/conversationService.js';
import ConversationList from '../components/ConversationList.vue';
import EmojiPickerModal from '../components/EmojiPickerModal.vue';
import ForwardMessageModal from '../components/ForwardMessageModal.vue';
import GroupCreationModal from '../components/GroupCreationModal.vue';
import GroupSettingsModal from '../components/GroupSettingsModal.vue';
import MessageComposer from '../components/MessageComposer.vue';
import MessageList from '../components/MessageList.vue';
import ProfileModal from '../components/ProfileModal.vue';
import UserAvatar from '../components/UserAvatar.vue';
import UserSearchModal from '../components/UserSearchModal.vue';

// There is no push channel, so the view polls. Conversations change less often
// than the messages of the one being read, hence the two intervals.
const CONVERSATIONS_POLL_MS = 3000;
const MESSAGES_POLL_MS = 2000;

// ChatView owns the conversation and message state and the polling that keeps
// them fresh; presentation lives in the components it composes.
export default {
    name: 'ChatView',
    components: {
        ConversationList,
        EmojiPickerModal,
        ForwardMessageModal,
        GroupCreationModal,
        GroupSettingsModal,
        MessageComposer,
        MessageList,
        ProfileModal,
        UserAvatar,
        UserSearchModal
    },
    props: {
        id: {
            type: String,
            required: false,
            default: null
        }
    },
    data() {
        return {
            currentUser: null,

            conversations: [],
            loading: false,
            errorMsg: null,

            selectedConversationId: null,
            selectedConversation: null,
            messages: [],
            messagesLoading: false,

            sending: false,
            replyingTo: null,

            // A modal is open when its flag is set, or - for the ones that act
            // on a specific message - when that message is set.
            showUserSearch: false,
            showGroupCreation: false,
            showProfile: false,
            showGroupSettings: false,
            reactingToMessage: null,
            forwardingMessage: null
        };
    },

    computed: {
        conversationName() {
            if (!this.selectedConversation) return 'Loading...';

            if (this.selectedConversation.type === 'group') {
                return this.selectedConversation.name || 'Group Chat';
            }

            const other = this.selectedConversation.participants?.find(p => p.id !== this.currentUser?.id);
            return other?.name || 'Unknown User';
        },

        conversationInfo() {
            if (!this.selectedConversation) return '';

            if (this.selectedConversation.type === 'group') {
                return `${this.selectedConversation.participants?.length || 0} members`;
            }
            return 'Direct message';
        }
    },

    watch: {
        // Keeps the view in step with the URL when the router navigates without
        // a click on the list, e.g. the back button.
        '$route'(to, from) {
            if (to.params.id && to.params.id !== from.params.id) {
                const conversation = this.conversations.find(c => c.id === to.params.id);
                if (conversation && conversation.id !== this.selectedConversationId) {
                    this.openConversation(conversation);
                }
            } else if (!to.params.id && this.selectedConversationId) {
                this.clearSelection();
            }
        }
    },

    mounted() {
        this.currentUser = authService.getCurrentUser();

        this.refreshConversations().then(() => {
            if (!this.id) return;

            const conversation = this.conversations.find(c => c.id === this.id);
            if (conversation) {
                this.selectConversation(conversation);
            }
        });

        this.conversationsPoll = setInterval(() => {
            if (!this.loading) {
                this.refreshConversations();
            }
        }, CONVERSATIONS_POLL_MS);

        this.messagesPoll = setInterval(() => {
            if (this.selectedConversationId && !this.messagesLoading && !this.sending) {
                this.refreshMessages();
            }
        }, MESSAGES_POLL_MS);
    },

    beforeUnmount() {
        clearInterval(this.conversationsPoll);
        clearInterval(this.messagesPoll);
        this.releasePhotoUrls();
    },

    methods: {
        async refreshConversations() {
            this.loading = true;
            this.errorMsg = null;

            try {
                const result = await conversationService.getMyConversations();
                this.conversations = Array.isArray(result) ? result : [];
            } catch (error) {
                this.errorMsg = error.response?.data?.message || 'Failed to load conversations';
                this.conversations = [];
            } finally {
                this.loading = false;
            }
        },

        selectConversation(conversation) {
            this.openConversation(conversation);

            if (this.$route.params.id !== conversation.id) {
                this.$router.replace(`/chat/${conversation.id}`);
            }
        },

        openConversation(conversation) {
            this.releasePhotoUrls();
            this.selectedConversationId = conversation.id;
            this.selectedConversation = conversation;
            this.messages = [];
            this.replyingTo = null;
            this.loadConversation();
        },

        clearSelection() {
            this.releasePhotoUrls();
            this.selectedConversationId = null;
            this.selectedConversation = null;
            this.messages = [];
            this.replyingTo = null;
        },

        async loadConversation() {
            if (!this.selectedConversationId) return;

            this.messagesLoading = true;
            this.errorMsg = null;

            try {
                this.selectedConversation = await conversationService.getConversation(this.selectedConversationId);
                // The API returns newest first; the view reads oldest first.
                this.messages = (this.selectedConversation.messages || []).slice().reverse();

                await this.loadPhotoMessages();
                this.$refs.messageList?.scrollToBottom();
            } catch (error) {
                this.errorMsg = error.response?.data?.message || 'Failed to load conversation';
            } finally {
                this.messagesLoading = false;
            }
        },

        async refreshMessages() {
            if (!this.selectedConversationId) return;

            try {
                const wasAtBottom = this.$refs.messageList?.isAtBottom() ?? true;

                const conversation = await conversationService.getConversation(this.selectedConversationId);
                const updated = (conversation.messages || []).slice().reverse();

                if (fingerprint(this.messages) === fingerprint(updated)) return;

                this.adoptPhotoUrls(updated);
                this.messages = updated;
                await this.loadPhotoMessages();

                if (wasAtBottom) {
                    this.$refs.messageList?.scrollToBottom();
                }
            } catch (error) {
                this.errorMsg = 'Failed to refresh messages';
            }
        },

        // loadPhotoMessages fetches the image behind every photo message that
        // does not have a real one yet, replacing the local preview shown
        // straight after sending.
        async loadPhotoMessages() {
            const conversationId = this.selectedConversationId;

            for (const message of this.messages) {
                if (message.messageType !== 'photo') continue;
                if (message.photoObjectUrl && !message.isPreviewUrl) continue;

                try {
                    const url = await conversationService.getMessagePhoto(conversationId, message.id);

                    // A refresh may have replaced the list while this was in
                    // flight; that copy of the message is no longer on screen.
                    if (!this.messages.includes(message)) {
                        URL.revokeObjectURL(url);
                        continue;
                    }

                    if (message.isPreviewUrl && message.photoObjectUrl) {
                        URL.revokeObjectURL(message.photoObjectUrl);
                    }
                    message.photoObjectUrl = url;
                    message.isPreviewUrl = false;
                } catch (error) {
                    // Leave the placeholder in place; the next poll retries.
                }
            }
        },

        // adoptPhotoUrls moves photos already fetched onto the refreshed message
        // objects, and releases the blobs nothing points at any more. Previews
        // are deliberately not carried over, so the real photo replaces them.
        adoptPhotoUrls(updated) {
            const held = new Map();
            for (const message of this.messages) {
                if (message.photoObjectUrl && !message.isPreviewUrl) {
                    held.set(message.id, message.photoObjectUrl);
                }
            }

            for (const message of updated) {
                const url = held.get(message.id);
                if (url) {
                    message.photoObjectUrl = url;
                    held.delete(message.id);
                }
            }

            for (const url of held.values()) {
                URL.revokeObjectURL(url);
            }
            for (const message of this.messages) {
                if (message.isPreviewUrl && message.photoObjectUrl) {
                    URL.revokeObjectURL(message.photoObjectUrl);
                }
            }
        },

        releasePhotoUrls() {
            for (const message of this.messages) {
                if (message.photoObjectUrl) {
                    URL.revokeObjectURL(message.photoObjectUrl);
                }
            }
        },

        async sendMessage({ text, file, preview }) {
            if (!this.selectedConversationId) return;

            this.sending = true;
            try {
                const replyToId = this.replyingTo?.id;
                const message = file
                    ? await conversationService.sendPhotoMessageWithCaption(this.selectedConversationId, file, text, replyToId)
                    : await conversationService.sendMessage(this.selectedConversationId, {
                        messageType: 'text',
                        content: text,
                        replyToMessageId: replyToId
                    });

                // Show the local copy of the photo until the stored one arrives.
                if (message.messageType === 'photo' && preview) {
                    message.photoObjectUrl = preview;
                    message.isPreviewUrl = true;
                }

                this.messages.push(message);
                this.replyingTo = null;
                this.$refs.composer?.clear();
                this.$refs.messageList?.scrollToBottom();

                await this.loadPhotoMessages();
                this.refreshConversations();
            } catch (error) {
                this.errorMsg = 'Failed to send message';
            } finally {
                this.sending = false;
            }
        },

        async deleteMessage(message) {
            if (!confirm('Delete this message?')) return;

            try {
                await conversationService.deleteMessage(this.selectedConversationId, message.id);
                if (message.photoObjectUrl) {
                    URL.revokeObjectURL(message.photoObjectUrl);
                }
                this.messages = this.messages.filter(m => m.id !== message.id);
            } catch (error) {
                this.errorMsg = 'Failed to delete message';
            }
        },

        openEmojiPicker(message) {
            this.reactingToMessage = message;
        },

        async addReaction(emoticon) {
            const message = this.reactingToMessage;
            this.reactingToMessage = null;

            try {
                const comment = await conversationService.commentMessage(
                    this.selectedConversationId,
                    message.id,
                    emoticon
                );
                this.applyReaction(message.id, comment);
            } catch (error) {
                this.errorMsg = 'Failed to add reaction';
            }
        },

        // toggleReaction treats a click on an existing reaction as "take mine
        // back" when it is the one already given, and as a change otherwise.
        async toggleReaction(message, emoticon) {
            const mine = message.comments?.find(c => c.userId === this.currentUser?.id);

            try {
                if (mine && mine.emoticon === emoticon) {
                    await conversationService.removeUserReaction(this.selectedConversationId, message.id);
                    this.applyReaction(message.id, null);
                    return;
                }

                const comment = await conversationService.commentMessage(
                    this.selectedConversationId,
                    message.id,
                    emoticon
                );
                this.applyReaction(message.id, comment);
            } catch (error) {
                this.errorMsg = 'Failed to toggle reaction';
            }
        },

        // applyReaction updates the local copy of a message so the change shows
        // before the next poll confirms it. A null comment removes the
        // reaction. Each user has at most one reaction per message.
        applyReaction(messageId, comment) {
            const message = this.messages.find(m => m.id === messageId);
            if (!message) return;

            const others = (message.comments || []).filter(c => c.userId !== this.currentUser?.id);
            message.comments = comment ? [...others, comment] : others;
        },

        async startConversation(user) {
            this.showUserSearch = false;

            try {
                const conversation = await conversationService.createDirectConversation(user.id);
                await this.refreshConversations();
                this.selectConversation(conversation);
            } catch (error) {
                this.errorMsg = 'Failed to start conversation';
            }
        },

        onMessageForwarded(targetName) {
            this.forwardingMessage = null;
            alert(`Message forwarded to ${targetName}`);
        },

        async onGroupUpdated() {
            await this.loadConversation();
            await this.refreshConversations();
        },

        onGroupLeft() {
            this.showGroupSettings = false;
            this.clearSelection();
            this.refreshConversations();
            alert('You have left the group');
        },

        handleLogout() {
            authService.logout();
            this.$router.push('/login');
        }
    }
};

// fingerprint reduces a message list to the parts the view renders, so a poll
// that brings back identical data does not churn the DOM.
function fingerprint(messages) {
    return JSON.stringify(messages.map(message => ({
        id: message.id,
        status: message.status,
        content: message.content,
        comments: message.comments?.map(c => ({ id: c.id, userId: c.userId, emoticon: c.emoticon })) || []
    })));
}
</script>

<style scoped>
.chat-container {
    height: 100vh;
    background-color: #f8f9fa;
}

.conversation-panel {
    background-color: #ffffff;
}

.conversation-header {
    min-height: 70px;
}
</style>
