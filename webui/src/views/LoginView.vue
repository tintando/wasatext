<template>
  <div class="container-fluid d-flex align-items-center justify-content-center min-vh-100 bg-light">
    <div class="row justify-content-center w-100">
      <div class="col-md-4 col-lg-3">
        <div class="card shadow">
          <div class="card-body p-4">
            <div class="text-center mb-4">
              <h2 class="h4 text-primary">WASAText</h2>
              <p class="text-muted">Connect with your friends</p>
            </div>

            <form @submit.prevent="handleLogin">
              <div class="mb-3">
                <label for="username" class="form-label">Username</label>
                <input 
                  id="username" 
                  v-model="username" 
                  type="text"
                  class="form-control"
                  :disabled="loading"
                  placeholder="Enter your username"
                  minlength="3"
                  maxlength="16"
                  required
                >
                <div class="form-text">3-16 characters</div>
              </div>

              <div class="d-grid">
                <button 
                  type="submit" 
                  class="btn btn-primary"
                  :disabled="loading || !isValidUsername"
                >
                  <span v-if="loading" class="spinner-border spinner-border-sm me-2" />
                  {{ loading ? 'Signing in...' : 'Sign In' }}
                </button>
              </div>
            </form>

            <ErrorMsg v-if="errorMsg" :msg="errorMsg" class="mt-3" />

            <div class="text-center mt-3">
              <small class="text-muted">
                Enter any username to sign in or create a new account
              </small>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import authService from '../services/authService.js';
import ErrorMsg from '../components/ErrorMsg.vue';

export default {
    name: 'LoginView',
    components: {
        ErrorMsg
    },
    data() {
        return {
            username: '',
            loading: false,
            errorMsg: null
        };
    },
    computed: {
        isValidUsername() {
            return this.username.trim().length >= 3 && this.username.trim().length <= 16;
        }
    },
    mounted() {
        // Redirect if already logged in
        if (authService.isAuthenticated()) {
            this.$router.push('/chat');
        }
    },
    methods: {
        async handleLogin() {
            if (!this.isValidUsername) return;

            this.loading = true;
            this.errorMsg = null;

            try {
                await authService.doLogin(this.username.trim());
                this.$router.push('/chat');
            } catch (error) {
                this.errorMsg = error.message;
            } finally {
                this.loading = false;
            }
        }
    }
};
</script>

<style scoped>
.min-vh-100 {
    min-height: 100vh;
}
</style>