import axios from './axios.js';

class ConversationService {
    async getMyConversations() {
        const response = await axios.get('/conversations');
        return response.data;
    }

    async getConversation(conversationId) {
        const response = await axios.get(`/conversations/${conversationId}`);
        return response.data;
    }

    async createDirectConversation(userId) {
        const response = await axios.post('/conversations', {
            userId: userId
        });
        return response.data;
    }

    async sendMessage(conversationId, messageData) {
        const response = await axios.post(`/conversations/${conversationId}/messages`, messageData);
        return response.data;
    }

    async sendPhotoMessageWithCaption(conversationId, photoFile, caption = '', replyToMessageId = null) {
        const formData = new FormData();
        formData.append('messageType', 'photo');
        formData.append('photo', photoFile);
        if (caption && caption.trim()) {
            formData.append('content', caption.trim());
        }
        if (replyToMessageId) {
            formData.append('replyToMessageId', replyToMessageId);
        }

        const response = await axios.post(`/conversations/${conversationId}/messages`, formData, {
            headers: {
                'Content-Type': 'multipart/form-data'
            }
        });

        return response.data;
    }

    async forwardMessage(fromConversationId, messageId, targetConversationId) {
        const response = await axios.post(
            `/conversations/${targetConversationId}/forwarded_messages`,
            {
                sourceConversationId: fromConversationId,
                sourceMessageId: messageId
            }
        );
        return response.data;
    }

    async forwardMessageToUser(fromConversationId, messageId, targetUserId) {
        // There is no conversation to forward into until one exists, so make
        // sure the direct conversation with the target is there first.
        const conversation = await this.createDirectConversation(targetUserId);

        return await this.forwardMessage(fromConversationId, messageId, conversation.id);
    }

    async deleteMessage(conversationId, messageId) {
        await axios.delete(`/conversations/${conversationId}/messages/${messageId}`);
    }

    async getMessagePhoto(conversationId, messageId) {
        const response = await axios.get(`/conversations/${conversationId}/messages/${messageId}/photo`, {
            responseType: 'blob'
        });
        return URL.createObjectURL(response.data);
    }

    async commentMessage(conversationId, messageId, emoticon) {
        const response = await axios.post(
            `/conversations/${conversationId}/messages/${messageId}/comments`,
            { emoticon }
        );
        return response.data;
    }

    async removeUserReaction(conversationId, messageId) {
        await axios.delete(`/conversations/${conversationId}/messages/${messageId}/user-reaction`);
    }

    async setGroupName(conversationId, newName) {
        const response = await axios.put(`/conversations/${conversationId}`, {
            name: newName
        });
        return response.data;
    }
}

export default new ConversationService();
