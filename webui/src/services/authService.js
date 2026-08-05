import axios from './axios.js';

class AuthService {
    constructor() {
        this.token = localStorage.getItem('wasatext_token');
        this.user = JSON.parse(localStorage.getItem('wasatext_user') || 'null');
    }

    async doLogin(username) {
        try {
            const response = await axios.post('/session', {
                name: username.trim()
            });
            
            this.token = response.data.identifier;
            this.user = {
                id: response.data.identifier,
                name: username.trim()
            };
            
            localStorage.setItem('wasatext_token', this.token);
            localStorage.setItem('wasatext_user', JSON.stringify(this.user));
            
            // Set default authorization header
            axios.defaults.headers.common['Authorization'] = `Bearer ${this.token}`;
            
            return this.user;
        } catch (error) {
            throw new Error(error.response?.data?.message || 'Login failed');
        }
    }

    logout() {
        this.token = null;
        this.user = null;
        localStorage.removeItem('wasatext_token');
        localStorage.removeItem('wasatext_user');
        delete axios.defaults.headers.common['Authorization'];
    }

    isAuthenticated() {
        return !!this.token;
    }

    getCurrentUser() {
        return this.user;
    }

    getToken() {
        return this.token;
    }

    updateCurrentUser(updatedUser) {
        this.user = updatedUser;
        localStorage.setItem('wasatext_user', JSON.stringify(updatedUser));
    }

    // Initialize axios with token if available
    init() {
        if (this.token) {
            axios.defaults.headers.common['Authorization'] = `Bearer ${this.token}`;
        }
    }
}

export default new AuthService();