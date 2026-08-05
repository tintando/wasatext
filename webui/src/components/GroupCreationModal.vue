<template>
  <div class="modal d-block" tabindex="-1" style="background-color: rgba(0,0,0,0.5)">
    <div class="modal-dialog">
      <div class="modal-content">
        <div class="modal-header">
          <h5 class="modal-title">Create New Group</h5>
          <button type="button" class="btn-close" @click="$emit('close')" />
        </div>
        <div class="modal-body">
          <form @submit.prevent="createGroup">
            <div class="mb-3">
              <label class="form-label">Group Name</label>
              <input 
                v-model="groupName" 
                type="text" 
                class="form-control"
                placeholder="Enter group name..."
                minlength="1"
                maxlength="50"
                required
              >
            </div>
                        
            <div class="mb-3">
              <label class="form-label">Search Users to Add</label>
              <input 
                v-model="searchQuery" 
                type="text" 
                class="form-control mb-3"
                placeholder="Search users..."
                @input="searchUsers"
              >
                            
              <div v-if="searchResults.length > 0" class="search-results mb-3">
                <div 
                  v-for="user in searchResults" 
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
                    class="btn btn-sm"
                    :class="isUserSelected(user) ? 'btn-danger' : 'btn-primary'"
                    @click="toggleUser(user)"
                  >
                    {{ isUserSelected(user) ? 'Remove' : 'Add' }}
                  </button>
                </div>
              </div>
            </div>
                        
            <div v-if="selectedUsers.length > 0" class="mb-3">
              <label class="form-label">Selected Members ({{ selectedUsers.length }})</label>
              <div class="selected-users">
                <span 
                  v-for="user in selectedUsers" 
                  :key="user.id"
                  class="badge bg-primary me-2 mb-2"
                >
                  {{ user.name }}
                  <button 
                    type="button" 
                    class="btn-close btn-close-white ms-1" 
                    style="font-size: 0.7em;"
                    @click="removeUser(user)"
                  />
                </span>
              </div>
            </div>
                        
            <ErrorMsg v-if="errorMsg" :msg="errorMsg" />
                        
            <div class="d-flex justify-content-end">
              <button 
                type="button" 
                class="btn btn-secondary me-2" 
                @click="$emit('close')"
              >
                Cancel
              </button>
              <button 
                type="submit" 
                class="btn btn-primary"
                :disabled="!groupName.trim() || selectedUsers.length === 0 || loading"
              >
                <span v-if="loading" class="spinner-border spinner-border-sm me-2" />
                Create Group
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import userService from '../services/userService.js';
import groupService from '../services/groupService.js';
import UserAvatar from './UserAvatar.vue';
import ErrorMsg from './ErrorMsg.vue';
import authService from '../services/authService.js';

export default {
    name: 'GroupCreationModal',
    components: {
        ErrorMsg,
        UserAvatar
    },
    emits: ['close', 'created'],
    data() {
        return {
            groupName: '',
            searchQuery: '',
            searchResults: [],
            selectedUsers: [],
            loading: false,
            errorMsg: null,
            searchLoading: false,
            currentUser: null
        };
    },

    mounted() {
        this.currentUser = authService.getCurrentUser();
    },
    methods: {
        async searchUsers() {
            if (this.searchQuery.trim().length < 1) {
                this.searchResults = [];
                return;
            }

            this.searchLoading = true;
            try {
                this.searchResults = await userService.searchUsers(this.searchQuery.trim());
                // Filter out current user and already selected users
                this.searchResults = this.searchResults.filter(user => 
                    user.id !== this.currentUser.id && 
                    !this.selectedUsers.find(selected => selected.id === user.id)
                );
            } catch (error) {
                this.searchResults = [];
            } finally {
                this.searchLoading = false;
            }
        },

        toggleUser(user) {
            if (this.isUserSelected(user)) {
                this.removeUser(user);
            } else {
                this.selectedUsers.push(user);
                // Remove from search results
                this.searchResults = this.searchResults.filter(u => u.id !== user.id);
            }
        },

        isUserSelected(user) {
            return this.selectedUsers.find(selected => selected.id === user.id) !== undefined;
        },

        removeUser(user) {
            this.selectedUsers = this.selectedUsers.filter(selected => selected.id !== user.id);
            // Add back to search results if search query matches
            if (this.searchQuery && user.name.toLowerCase().includes(this.searchQuery.toLowerCase())) {
                this.searchResults.push(user);
            }
        },

        async createGroup() {
            if (!this.groupName.trim() || this.selectedUsers.length === 0) return;

            this.loading = true;
            this.errorMsg = null;
            
            try {
                const memberIds = this.selectedUsers.map(user => user.id);
                const group = await groupService.createGroup(this.groupName.trim(), memberIds);
                
                this.$emit('created', group);
                this.$emit('close');
            } catch (error) {
                this.errorMsg = error.response?.data?.message || 'Failed to create group';
            } finally {
                this.loading = false;
            }
        }
    }
};
</script>

<style scoped>
.avatar {
    width: 30px;
    height: 30px;
    background-color: #007bff;
    color: white;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    font-weight: bold;
    font-size: 0.8em;
}

.modal {
    display: block !important;
}

.selected-users .badge {
    font-size: 0.9em;
}
</style>