<template>
  <div class="message-input-container bg-white border-top p-3">
    <div v-if="replyingTo" class="reply-preview bg-light p-2 rounded mb-2 d-flex justify-content-between">
      <small>Replying to: {{ replyingTo.content || 'Photo' }}</small>
      <button class="btn btn-sm p-0" @click="$emit('cancel-reply')">×</button>
    </div>

    <form class="d-flex align-items-end" @submit.prevent="submit">
      <div class="flex-grow-1 me-2">
        <textarea
          v-model="text"
          class="form-control no-resize"
          placeholder="Type a message..."
          :disabled="sending"
          rows="1"
          @keydown.enter.exact.prevent="submit"
        />
      </div>

      <div class="d-flex">
        <label class="btn btn-outline-secondary me-2">
          <svg width="16" height="16" fill="currentColor" viewBox="0 0 16 16">
            <path d="M6.002 5.5a1.5 1.5 0 1 1-3 0 1.5 1.5 0 0 1 3 0z" />
            <path d="M1.5 2A1.5 1.5 0 0 0 0 3.5v9A1.5 1.5 0 0 0 1.5 14h13a1.5 1.5 0 0 0 1.5-1.5v-9A1.5 1.5 0 0 0 14.5 2h-13zm13 1a.5.5 0 0 1 .5.5v6l-3.775-1.947a.5.5 0 0 0-.577.093l-3.71 3.71-2.66-1.772a.5.5 0 0 0-.63.062L1.002 12v.54A.505.505 0 0 1 1.5 13h13a.5.5 0 0 1 .5-.5v-9a.5.5 0 0 1-.5-.5h-13z" />
          </svg>
          <input
            ref="fileInput"
            type="file"
            hidden
            accept="image/*"
            :disabled="sending"
            @change="selectPhoto"
          >
        </label>

        <button
          type="submit"
          class="btn btn-primary"
          :disabled="sending || !hasContent"
        >
          <span v-if="sending" class="spinner-border spinner-border-sm me-1" />
          Send
        </button>
      </div>
    </form>

    <div v-if="photo" class="photo-preview mt-2">
      <div class="d-flex align-items-center">
        <img :src="photo.preview" alt="Photo preview" class="rounded preview-image">
        <button class="btn btn-sm btn-outline-danger ms-2" @click="removePhoto">
          Remove
        </button>
      </div>
    </div>
  </div>
</template>

<script>
// MessageComposer owns the draft message: its text, the attached photo and the
// preview URL for that photo. The parent sends it and then calls clear().
export default {
    name: 'MessageComposer',
    props: {
        sending: {
            type: Boolean,
            default: false
        },
        replyingTo: {
            type: Object,
            default: null
        }
    },
    emits: ['send', 'cancel-reply'],
    data() {
        return {
            text: '',
            photo: null
        };
    },
    computed: {
        hasContent() {
            return this.text.trim().length > 0 || this.photo !== null;
        }
    },
    beforeUnmount() {
        this.releasePreview();
    },
    methods: {
        submit() {
            if (!this.hasContent || this.sending) return;

            this.$emit('send', {
                text: this.text.trim(),
                file: this.photo?.file || null,
                preview: this.photo?.preview || null
            });
        },

        selectPhoto(event) {
            const file = event.target.files[0];
            if (!file) return;

            this.releasePreview();
            this.photo = { file, preview: URL.createObjectURL(file) };
        },

        removePhoto() {
            this.releasePreview();
            this.photo = null;
            if (this.$refs.fileInput) {
                this.$refs.fileInput.value = '';
            }
        },

        // clear resets the draft after a successful send. The preview URL is
        // handed to the caller with the send payload, so it is not revoked
        // here: the caller releases it once the real photo has loaded.
        clear() {
            this.text = '';
            this.photo = null;
            if (this.$refs.fileInput) {
                this.$refs.fileInput.value = '';
            }
        },

        releasePreview() {
            if (this.photo?.preview) {
                URL.revokeObjectURL(this.photo.preview);
            }
        }
    }
};
</script>

<style scoped>
.no-resize {
    resize: none;
}

.preview-image {
    max-width: 100px;
    max-height: 100px;
}
</style>
