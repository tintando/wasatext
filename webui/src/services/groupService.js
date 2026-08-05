import axios from './axios.js';

class GroupService {
    async createGroup(name, memberIds) {
        const response = await axios.post('/groups', {
            name,
            memberIds
        });
        return response.data;
    }

    async setGroupPhoto(groupId, photoFile) {
        await axios.put(`/groups/${groupId}/photo`, photoFile, {
            headers: {
                'Content-Type': photoFile.type
            }
        });
    }

    async getGroupPhoto(groupId) {
        const response = await axios.get(`/groups/${groupId}/photo`, {
            responseType: 'blob'
        });
        return URL.createObjectURL(response.data);
    }

    async addToGroup(groupId, userId) {
        await axios.post(`/groups/${groupId}/members`, { userId });
    }

    async leaveGroup(groupId, userId) {
        await axios.delete(`/groups/${groupId}/members/${userId}`);
    }
}

export default new GroupService();
