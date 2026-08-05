<template>
  <div class="user-avatar" :class="sizeClass">
    <img
      v-if="photoObjectUrl && !imageError"
      :src="photoObjectUrl"
      :alt="name"
      class="avatar-image"
      @error="handleImageError"
    >
    <div v-else class="avatar-initials" :class="{ 'text-white': darkBackground }">
      {{ initials }}
    </div>
  </div>
</template>

<script>
import userService from '../services/userService.js';
import groupService from '../services/groupService.js';

// UserAvatar shows a user's or group's photo, falling back to their initial
// when there is no photo or it cannot be loaded.
export default {
    name: 'UserAvatar',
    props: {
        userId: {
            type: String,
            default: null
        },
        groupId: {
            type: String,
            default: null
        },
        avatarType: {
            type: String,
            default: 'user',
            validator: value => ['user', 'group'].includes(value)
        },
        // An already usable image URL. When given, nothing is fetched.
        photoUrl: {
            type: String,
            default: null
        },
        name: {
            type: String,
            required: true
        },
        size: {
            type: String,
            default: 'medium',
            validator: value => ['small', 'medium', 'large'].includes(value)
        },
        darkBackground: {
            type: Boolean,
            default: false
        }
    },
    data() {
        return {
            photoObjectUrl: null,
            // True when the photo could not be shown, so initials are used.
            imageError: false,
            // Whether photoObjectUrl is a blob this component has to release.
            ownsObjectUrl: false
        };
    },
    computed: {
        initials() {
            return this.name.charAt(0).toUpperCase();
        },
        sizeClass() {
            return `avatar-${this.size}`;
        }
    },
    watch: {
        userId: 'loadPhoto',
        groupId: 'loadPhoto',
        avatarType: 'loadPhoto',
        photoUrl: 'loadPhoto'
    },
    async mounted() {
        await this.loadPhoto();
    },
    beforeUnmount() {
        this.releasePhoto();
    },
    methods: {
        isDisplayableUrl(url) {
            return Boolean(url) && (url.startsWith('http://') || url.startsWith('https://') || url.startsWith('blob:'));
        },

        async loadPhoto() {
            this.releasePhoto();
            this.imageError = false;

            // A URL supplied by the caller is used as is; it is not ours to free.
            if (this.isDisplayableUrl(this.photoUrl)) {
                this.photoObjectUrl = this.photoUrl;
                return;
            }

            const fetchPhoto = this.avatarType === 'group'
                ? this.groupId && (() => groupService.getGroupPhoto(this.groupId))
                : this.userId && (() => userService.getUserPhoto(this.userId));

            if (!fetchPhoto) {
                this.imageError = true;
                return;
            }

            try {
                this.photoObjectUrl = await fetchPhoto();
                this.ownsObjectUrl = true;
            } catch (error) {
                // A 404 just means there is no photo, which is not a problem
                // worth reporting; anything else is worth a warning.
                if (error.response?.status !== 404) {
                    console.warn('Failed to load avatar photo:', error);
                }
                this.imageError = true;
            }
        },

        // releasePhoto frees the blob this component fetched. Avatars are
        // recreated whenever a photo changes, so skipping this leaks one blob
        // per render.
        releasePhoto() {
            if (this.ownsObjectUrl && this.photoObjectUrl) {
                URL.revokeObjectURL(this.photoObjectUrl);
            }
            this.photoObjectUrl = null;
            this.ownsObjectUrl = false;
        },

        handleImageError() {
            this.imageError = true;
            this.releasePhoto();
        }
    }
};
</script>

<style scoped>
.user-avatar {
    display: flex;
    align-items: center;
    justify-content: center;
    border-radius: 50%;
    overflow: hidden;
    background-color: #6c757d;
    color: white;
    font-weight: 500;
    flex-shrink: 0;
}

.avatar-image {
    width: 100%;
    height: 100%;
    object-fit: cover;
}

.avatar-initials {
    width: 100%;
    height: 100%;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: inherit;
}

.avatar-small {
    width: 32px;
    height: 32px;
    font-size: 12px;
}

.avatar-medium {
    width: 40px;
    height: 40px;
    font-size: 14px;
}

.avatar-large {
    width: 80px;
    height: 80px;
    font-size: 24px;
}

.text-white {
    color: white !important;
}
</style>
