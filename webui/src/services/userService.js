import axios from './axios.js';

class UserService {
    async updateUserName(userId, newName) {
        const response = await axios.put(`/users/${userId}`, {
            name: newName
        });
        return response.data;
    }

    async setMyPhoto(userId, photoFile) {
        const response = await axios.put(`/users/${userId}/photo`, photoFile, {
            headers: {
                'Content-Type': photoFile.type
            }
        });
        return response.data;
    }

    async getUserPhoto(userId) {
        const response = await axios.get(`/users/${userId}/photo`, {
            responseType: 'blob'
        });
        return URL.createObjectURL(response.data);
    }

    async searchUsers(query) {
        const response = await axios.get('/users', {
            params: { q: query }
        });
        return response.data.users;
    }
}

export default new UserService();
