<template>
  <BaseModal title="Group Settings" @close="$emit('close')">
    <div class="mb-4">
      <div class="text-center mb-4">
        <div class="position-relative d-inline-block">
          <UserAvatar
            :key="`group-settings-${conversation?.id}-${conversation?.photoUrl || 'no-photo'}`"
            :group-id="conversation?.id"
            avatar-type="group"
            :photo-url="conversation?.photoUrl"
            :name="conversation?.name || 'Group'"
            size="large"
          />
        </div>
        <h5 class="mt-3 mb-0">{{ conversation?.name || 'Group' }}</h5>
      </div>

      <div class="mb-3">
        <label class="form-label">Group Photo</label>
        <input
          ref="photoInput"
          type="file"
          class="form-control"
          accept="image/png,image/jpeg,image/gif"
          @change="selectPhoto"
        >
        <div class="form-text">Upload PNG, JPEG, or GIF image</div>
      </div>

      <div v-if="selectedPhoto" class="d-flex justify-content-end mb-3">
        <button
          type="button"
          class="btn btn-primary"
          :disabled="photoLoading"
          @click="uploadPhoto"
        >
          <span v-if="photoLoading" class="spinner-border spinner-border-sm me-2" />
          Upload Photo
        </button>
      </div>
    </div>

    <ErrorMsg v-if="errorMsg" :msg="errorMsg" class="mb-3" />

    <div class="mb-3">
      <label class="form-label">Group Name</label>
      <input
        v-model="name"
        type="text"
        class="form-control"
        :placeholder="conversation?.name || 'Group Chat'"
      >
    </div>

    <div class="mb-3">
      <label class="form-label">Members ({{ conversation?.participants?.length || 0 }})</label>
      <div class="list-group mb-3">
        <div
          v-for="participant in conversation?.participants"
          :key="participant.id"
          class="list-group-item d-flex justify-content-between align-items-center"
        >
          <div class="d-flex align-items-center">
            <UserAvatar
              :user-id="participant.id"
              :photo-url="participant.photoUrl"
              :name="participant.name"
              size="small"
              class="me-2"
            />
            <span>{{ participant.name }}</span>
            <span v-if="participant.id === currentUser?.id" class="badge bg-primary ms-2">You</span>
          </div>
        </div>
      </div>

      <div class="add-member-section">
        <label class="form-label">Add New Member</label>
        <input
          v-model="memberQuery"
          type="text"
          class="form-control mb-3"
          placeholder="Search users to add..."
          @input="searchUsers"
        >

        <div v-if="candidates.length > 0" class="search-results">
          <div
            v-for="user in candidates"
            :key="user.id"
            class="d-flex align-items-center justify-content-between p-2 border rounded mb-2"
          >
            <div class="d-flex align-items-center">
              <UserAvatar
                :user-id="user.id"
                :photo-url="user.photoUrl"
                :name="user.name"
                size="small"
                class="me-2"
              />
              <span>{{ user.name }}</span>
            </div>
            <button
              type="button"
              class="btn btn-sm btn-primary"
              :disabled="addingMember"
              @click="addMember(user)"
            >
              <span v-if="addingMember" class="spinner-border spinner-border-sm me-1" />
              Add
            </button>
          </div>
        </div>
      </div>
    </div>

    <div class="d-grid gap-2">
      <button class="btn btn-primary" :disabled="nameLoading" @click="saveName">
        <span v-if="nameLoading" class="spinner-border spinner-border-sm me-1" />
        Update Group
      </button>
      <button class="btn btn-outline-danger" :disabled="leaving" @click="leave">
        <span v-if="leaving" class="spinner-border spinner-border-sm me-1" />
        Leave Group
      </button>
    </div>
  </BaseModal>
</template>

<script>
import conversationService from '../services/conversationService.js';
import groupService from '../services/groupService.js';
import userService from '../services/userService.js';
import BaseModal from './BaseModal.vue';
import ErrorMsg from './ErrorMsg.vue';
import UserAvatar from './UserAvatar.vue';
import { validatePhotoFile } from '../utils/photo.js';

// GroupSettingsModal edits one group: its name, its photo, and its membership.
export default {
    name: 'GroupSettingsModal',
    components: {
        BaseModal,
        ErrorMsg,
        UserAvatar
    },
    props: {
        conversation: {
            type: Object,
            required: true
        },
        currentUser: {
            type: Object,
            default: null
        }
    },
    emits: ['close', 'updated', 'left'],
    data() {
        return {
            name: this.conversation?.name || '',
            nameLoading: false,
            selectedPhoto: null,
            photoLoading: false,
            memberQuery: '',
            candidates: [],
            addingMember: false,
            leaving: false,
            errorMsg: null
        };
    },
    methods: {
        async saveName() {
            if (!this.name.trim()) {
                this.errorMsg = 'Group name cannot be empty';
                return;
            }

            this.nameLoading = true;
            try {
                await conversationService.setGroupName(this.conversation.id, this.name.trim());
                this.$emit('updated');
                this.$emit('close');
            } catch (error) {
                this.errorMsg = 'Failed to update group settings';
            } finally {
                this.nameLoading = false;
            }
        },

        selectPhoto(event) {
            const file = event.target.files[0];
            if (!file) return;

            const error = validatePhotoFile(file);
            if (error) {
                this.errorMsg = error;
                return;
            }

            this.selectedPhoto = file;
            this.errorMsg = null;
        },

        async uploadPhoto() {
            if (!this.selectedPhoto) return;

            this.photoLoading = true;
            try {
                await groupService.setGroupPhoto(this.conversation.id, this.selectedPhoto);
                this.selectedPhoto = null;
                this.$refs.photoInput.value = '';
                this.errorMsg = null;
                this.$emit('updated');
            } catch (error) {
                this.errorMsg = error.response?.data?.message || 'Failed to upload group photo';
            } finally {
                this.photoLoading = false;
            }
        },

        async searchUsers() {
            const query = this.memberQuery.trim();
            if (query.length < 1) {
                this.candidates = [];
                return;
            }

            try {
                const results = await userService.searchUsers(query);
                const memberIds = this.conversation?.participants?.map(p => p.id) || [];
                this.candidates = results.filter(user => !memberIds.includes(user.id));
            } catch (error) {
                this.candidates = [];
            }
        },

        async addMember(user) {
            this.addingMember = true;
            try {
                await groupService.addToGroup(this.conversation.id, user.id);
                this.candidates = this.candidates.filter(candidate => candidate.id !== user.id);
                this.memberQuery = '';
                this.$emit('updated');
            } catch (error) {
                this.errorMsg = 'Failed to add user to group';
            } finally {
                this.addingMember = false;
            }
        },

        async leave() {
            if (!confirm('Are you sure you want to leave this group? You will no longer receive messages from this group.')) {
                return;
            }

            this.leaving = true;
            try {
                await groupService.leaveGroup(this.conversation.id, this.currentUser.id);
                this.$emit('left');
            } catch (error) {
                this.errorMsg = 'Failed to leave group';
            } finally {
                this.leaving = false;
            }
        }
    }
};
</script>
