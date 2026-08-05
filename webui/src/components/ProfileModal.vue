<template>
  <BaseModal title="User Profile" @close="$emit('close')">
    <div class="text-center mb-4">
      <div class="position-relative d-inline-block">
        <img
          v-if="photoUrl"
          :src="photoUrl"
          alt="Profile photo"
          class="avatar-large-img rounded-circle"
          @error="photoUrl = null"
        >
        <div v-else class="avatar-large">
          {{ user?.name?.charAt(0).toUpperCase() }}
        </div>
      </div>

      <div class="mt-3 mb-0">
        <div v-if="!editingName" class="d-flex align-items-center justify-content-center">
          <h5 class="mb-0 me-2">{{ user?.name }}</h5>
          <button class="btn btn-sm btn-outline-secondary" @click="startEditingName">
            Edit
          </button>
        </div>

        <div v-else class="d-flex flex-column align-items-center">
          <input
            v-model="newName"
            type="text"
            class="form-control mb-2 name-input"
            placeholder="Enter new username"
            minlength="3"
            maxlength="16"
            @keyup.enter="saveName"
          >
          <div class="d-flex gap-2">
            <button
              class="btn btn-sm btn-success"
              :disabled="nameLoading || !isNameValid"
              @click="saveName"
            >
              <span v-if="nameLoading" class="spinner-border spinner-border-sm me-1" />
              Save
            </button>
            <button
              class="btn btn-sm btn-secondary"
              :disabled="nameLoading"
              @click="cancelEditingName"
            >
              Cancel
            </button>
          </div>
          <small class="text-muted mt-1">3-16 characters</small>
        </div>
      </div>
    </div>

    <ErrorMsg v-if="errorMsg" :msg="errorMsg" class="mb-3" />

    <div class="mb-3">
      <label class="form-label">Profile Photo</label>
      <input
        ref="photoInput"
        type="file"
        class="form-control"
        accept="image/png,image/jpeg,image/gif"
        @change="selectPhoto"
      >
      <div class="form-text">Upload PNG, JPEG, or GIF image</div>
    </div>

    <div class="d-flex justify-content-end">
      <button
        v-if="selectedPhoto"
        type="button"
        class="btn btn-primary"
        :disabled="photoLoading"
        @click="uploadPhoto"
      >
        <span v-if="photoLoading" class="spinner-border spinner-border-sm me-2" />
        Upload Photo
      </button>
      <button
        v-else
        type="button"
        class="btn btn-secondary"
        @click="$emit('close')"
      >
        Close
      </button>
    </div>
  </BaseModal>
</template>

<script>
import authService from '../services/authService.js';
import userService from '../services/userService.js';
import BaseModal from './BaseModal.vue';
import ErrorMsg from './ErrorMsg.vue';
import { validatePhotoFile } from '../utils/photo.js';

const MIN_NAME_LENGTH = 3;
const MAX_NAME_LENGTH = 16;

// ProfileModal shows the signed-in user's avatar and name, and lets them
// change either one.
export default {
    name: 'ProfileModal',
    components: {
        BaseModal,
        ErrorMsg
    },
    props: {
        user: {
            type: Object,
            default: null
        }
    },
    emits: ['close', 'renamed'],
    data() {
        return {
            photoUrl: null,
            selectedPhoto: null,
            photoLoading: false,
            editingName: false,
            newName: '',
            nameLoading: false,
            errorMsg: null
        };
    },
    computed: {
        isNameValid() {
            const length = this.newName.trim().length;
            return length >= MIN_NAME_LENGTH && length <= MAX_NAME_LENGTH;
        }
    },
    async mounted() {
        await this.loadPhoto();
    },
    beforeUnmount() {
        this.releasePhoto();
    },
    methods: {
        async loadPhoto() {
            this.releasePhoto();
            try {
                this.photoUrl = await userService.getUserPhoto(this.user.id);
            } catch (error) {
                this.photoUrl = null;
            }
        },

        // releasePhoto frees the blob URL the avatar was loaded into; without
        // this every reload would leak one.
        releasePhoto() {
            if (this.photoUrl) {
                URL.revokeObjectURL(this.photoUrl);
                this.photoUrl = null;
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
                await userService.setMyPhoto(this.user.id, this.selectedPhoto);
                this.selectedPhoto = null;
                this.$refs.photoInput.value = '';
                this.errorMsg = null;
                await this.loadPhoto();
            } catch (error) {
                this.errorMsg = error.response?.data?.message || 'Failed to upload photo';
            } finally {
                this.photoLoading = false;
            }
        },

        startEditingName() {
            this.editingName = true;
            this.newName = this.user?.name || '';
        },

        cancelEditingName() {
            this.editingName = false;
            this.newName = '';
        },

        async saveName() {
            if (!this.isNameValid) return;

            this.nameLoading = true;
            try {
                const updated = await userService.updateUserName(this.user.id, this.newName.trim());
                const renamed = { id: this.user.id, name: updated.name };

                // Keep the stored session in step with the new name.
                authService.updateCurrentUser(renamed);
                this.$emit('renamed', renamed);

                this.editingName = false;
                this.newName = '';
                this.errorMsg = null;
            } catch (error) {
                // The API reports this failure as plain text, not JSON.
                this.errorMsg = error.response?.data || error.message || 'Failed to update username';
            } finally {
                this.nameLoading = false;
            }
        }
    }
};
</script>

<style scoped>
.avatar-large {
    width: 80px;
    height: 80px;
    background-color: #007bff;
    color: white;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    font-weight: bold;
    font-size: 1.5rem;
    margin: 0 auto;
}

.avatar-large-img {
    width: 80px;
    height: 80px;
    object-fit: cover;
}

.name-input {
    max-width: 250px;
}
</style>
