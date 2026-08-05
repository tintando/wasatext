package api

import (
	"net/http"
)

// Handler returns an instance of httprouter.Router that handle APIs registered here.
//
// Routes registered with wrapAuth require a Bearer token and receive the caller
// as ctx.UserID; routes registered with wrap are reachable without one.
func (rt *_router) Handler() http.Handler {
	// Authentication
	rt.router.POST("/session", rt.wrap(rt.doLogin))

	// User endpoints
	rt.router.GET("/users", rt.wrapAuth(rt.searchUsers))
	rt.router.GET("/users/:userId", rt.wrapAuth(rt.getUserProfile))
	rt.router.PUT("/users/:userId", rt.wrapAuth(rt.setMyUserName))
	rt.router.PUT("/users/:userId/photo", rt.wrapAuth(rt.setMyPhoto))
	rt.router.GET("/users/:userId/photo", rt.wrap(rt.getUserPhoto))

	// Conversation endpoints
	rt.router.GET("/conversations", rt.wrapAuth(rt.getMyConversations))
	rt.router.POST("/conversations", rt.wrapAuth(rt.createDirectConversation))
	rt.router.GET("/conversations/:conversationId", rt.wrapAuth(rt.getConversation))
	rt.router.PUT("/conversations/:conversationId", rt.wrapAuth(rt.updateConversation))
	rt.router.GET("/conversations/:conversationId/messages", rt.wrapAuth(rt.getConversationMessages))

	// Message endpoints
	rt.router.POST("/conversations/:conversationId/messages", rt.wrapAuth(rt.sendMessage))
	rt.router.DELETE("/conversations/:conversationId/messages/:messageId", rt.wrapAuth(rt.deleteMessage))
	rt.router.PUT("/conversations/:conversationId/messages/:messageId", rt.wrapAuth(rt.updateMessageStatus))
	rt.router.GET("/conversations/:conversationId/messages/:messageId/photo", rt.wrapAuth(rt.getMessagePhoto))
	rt.router.POST("/conversations/:conversationId/forwarded_messages", rt.wrapAuth(rt.forwardMessage))

	// Comment endpoints
	rt.router.POST("/conversations/:conversationId/messages/:messageId/comments", rt.wrapAuth(rt.commentMessage))
	rt.router.DELETE("/conversations/:conversationId/messages/:messageId/comments/:commentId", rt.wrapAuth(rt.uncommentMessage))
	rt.router.DELETE("/conversations/:conversationId/messages/:messageId/user-reaction", rt.wrapAuth(rt.removeUserReaction))

	// Group endpoints
	rt.router.POST("/groups", rt.wrapAuth(rt.createGroup))
	rt.router.GET("/groups/:groupId", rt.wrapAuth(rt.getGroup))
	rt.router.PUT("/groups/:groupId", rt.wrapAuth(rt.setGroupName))
	rt.router.PUT("/groups/:groupId/photo", rt.wrapAuth(rt.setGroupPhoto))
	rt.router.GET("/groups/:groupId/photo", rt.wrapAuth(rt.getGroupPhoto))
	rt.router.GET("/groups/:groupId/members", rt.wrapAuth(rt.getGroupMembers))
	rt.router.POST("/groups/:groupId/members", rt.wrapAuth(rt.addToGroup))
	rt.router.DELETE("/groups/:groupId/members/:userId", rt.wrapAuth(rt.leaveGroup))

	// Special routes
	rt.router.GET("/liveness", rt.liveness)

	return rt.router
}
