<template>
  <BaseModal title="React to Message" dialog-class="modal-sm" @close="$emit('close')">
    <div class="emoji-grid">
      <button
        v-for="emoji in emojis"
        :key="emoji"
        class="btn btn-outline-secondary emoji-btn"
        @click="$emit('select', emoji)"
      >
        {{ emoji }}
      </button>
    </div>
  </BaseModal>
</template>

<script>
import BaseModal from './BaseModal.vue';

// The reactions offered when responding to a message.
const COMMON_EMOJIS = ['😀', '😂', '❤️', '👍', '👎', '😊', '😢', '😮', '😡', '🎉', '🔥', '💯'];

// EmojiPickerModal picks the emoticon to react to a message with.
export default {
    name: 'EmojiPickerModal',
    components: {
        BaseModal
    },
    emits: ['close', 'select'],
    data() {
        return {
            emojis: COMMON_EMOJIS
        };
    }
};
</script>

<style scoped>
.emoji-grid {
    display: grid;
    grid-template-columns: repeat(4, 1fr);
    gap: 10px;
}

.emoji-btn {
    font-size: 1.5em;
    padding: 10px;
}
</style>
