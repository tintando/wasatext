<template>
  <BaseModal title="Start New Conversation" @close="$emit('close')">
    <input
      v-model="query"
      type="text"
      class="form-control mb-3"
      placeholder="Search users..."
      @input="search"
    >

    <div v-if="results.length > 0">
      <div
        v-for="user in results"
        :key="user.id"
        class="d-flex align-items-center p-2 border-bottom user-item"
        @click="$emit('select', user)"
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
    <div v-else-if="query && !loading" class="text-muted text-center py-3">
      No users found
    </div>

    <div v-if="loading" class="text-center py-3">
      <div class="spinner-border spinner-border-sm" />
    </div>
  </BaseModal>
</template>

<script>
import userService from '../services/userService.js';
import BaseModal from './BaseModal.vue';
import UserAvatar from './UserAvatar.vue';

// UserSearchModal picks the person to open a new direct conversation with.
export default {
    name: 'UserSearchModal',
    components: {
        BaseModal,
        UserAvatar
    },
    props: {
        currentUserId: {
            type: String,
            default: null
        }
    },
    emits: ['close', 'select'],
    data() {
        return {
            query: '',
            results: [],
            loading: false
        };
    },
    methods: {
        async search() {
            const query = this.query.trim();
            if (query.length < 1) {
                this.results = [];
                return;
            }

            this.loading = true;
            try {
                const users = await userService.searchUsers(query);
                this.results = users.filter(user => user.id !== this.currentUserId);
            } catch (error) {
                this.results = [];
            } finally {
                this.loading = false;
            }
        }
    }
};
</script>

<style scoped>
.user-item {
    cursor: pointer;
}

.user-item:hover {
    background-color: #f8f9fa;
}
</style>
