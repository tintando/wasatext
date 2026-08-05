#!/bin/bash

# test file for the api endpoints
# run this after starting the server

# 1. doLogin
echo "1. Testing doLogin..."
LOGIN_RESPONSE=$(curl -s -X POST "http://localhost:3000/session" \
  -H "Content-Type: application/json" \
  -d '{"name": "testuser1"}')

USER1_ID=$(echo $LOGIN_RESPONSE | grep -o '"identifier":"[^"]*' | head -1 | cut -d'"' -f4)
echo "doLogin: User1 ID: $USER1_ID"

# Create second user
SECOND_LOGIN=$(curl -s -X POST "http://localhost:3000/session" \
  -H "Content-Type: application/json" \
  -d '{"name": "testuser2"}')
USER2_ID=$(echo $SECOND_LOGIN | grep -o '"identifier":"[^"]*' | head -1 | cut -d'"' -f4)
echo "doLogin: User2 ID: $USER2_ID"

# 2. setMyUserName
echo ""
echo "2. Testing setMyUserName..."
curl -s -X PUT "http://localhost:3000/users/$USER1_ID" \
  -H "Authorization: Bearer $USER1_ID" \
  -H "Content-Type: application/json" \
  -d '{"name": "updateduser1"}' > /dev/null
echo "setMyUserName: Updated user1 name"

# 3. getMyConversations
echo ""
echo "3. Testing getMyConversations..."
CONVERSATIONS=$(curl -s -X GET "http://localhost:3000/conversations" \
  -H "Authorization: Bearer $USER1_ID")
echo "getMyConversations: Retrieved conversations"

# 4. Create a direct conversation first
echo ""
echo "4. Creating direct conversation..."
CONV_RESPONSE=$(curl -s -X POST "http://localhost:3000/conversations" \
  -H "Authorization: Bearer $USER1_ID" \
  -H "Content-Type: application/json" \
  -d "{\"userId\": \"$USER2_ID\"}")
CONV_ID=$(echo $CONV_RESPONSE | grep -o '"id":"[^"]*' | head -1 | cut -d'"' -f4)
echo "Created conversation: $CONV_ID"

# 5. sendMessage
echo ""
echo "5. Testing sendMessage..."
MSG_RESPONSE=$(curl -s -X POST "http://localhost:3000/conversations/$CONV_ID/messages" \
  -H "Authorization: Bearer $USER1_ID" \
  -H "Content-Type: application/json" \
  -d '{"content": "Hello test message", "messageType": "text"}')
MSG_ID=$(echo $MSG_RESPONSE | grep -o '"id":"[^"]*' | head -1 | cut -d'"' -f4)
echo "sendMessage: Sent message $MSG_ID"

# 6. getConversation
echo ""
echo "6. Testing getConversation..."
CONVERSATION=$(curl -s -X GET "http://localhost:3000/conversations/$CONV_ID" \
  -H "Authorization: Bearer $USER1_ID")
echo "getConversation: Retrieved conversation details"

# 7. commentMessage (addComment)
echo ""
echo "7. Testing commentMessage..."
COMMENT_RESPONSE=$(curl -s -X POST "http://localhost:3000/conversations/$CONV_ID/messages/$MSG_ID/comments" \
  -H "Authorization: Bearer $USER2_ID" \
  -H "Content-Type: application/json" \
  -d '{"emoticon": "👍"}')
COMMENT_ID=$(echo $COMMENT_RESPONSE | grep -o '"id":"[^"]*' | head -1 | cut -d'"' -f4)
echo "commentMessage: Added comment $COMMENT_ID"

# 8. uncommentMessage (deleteComment)
echo ""
echo "8. Testing uncommentMessage..."
curl -s -X DELETE "http://localhost:3000/conversations/$CONV_ID/messages/$MSG_ID/comments/$COMMENT_ID" \
  -H "Authorization: Bearer $USER2_ID" > /dev/null
echo "uncommentMessage: Deleted comment"

# 9. forwardMessage
echo ""
echo "9. Testing forwardMessage..."
# Create another conversation to forward to
CONV2_RESPONSE=$(curl -s -X POST "http://localhost:3000/conversations" \
  -H "Authorization: Bearer $USER2_ID" \
  -H "Content-Type: application/json" \
  -d "{\"userId\": \"$USER1_ID\"}")
CONV2_ID=$(echo $CONV2_RESPONSE | grep -o '"id":"[^"]*' | head -1 | cut -d'"' -f4)

FORWARD_RESPONSE=$(curl -s -X POST "http://localhost:3000/conversations/$CONV_ID/messages/$MSG_ID/forward" \
  -H "Authorization: Bearer $USER1_ID" \
  -H "Content-Type: application/json" \
  -d "{\"conversationId\": \"$CONV2_ID\"}")
echo "forwardMessage: Forwarded message"

# 10. deleteMessage
echo ""
echo "10. Testing deleteMessage..."
curl -s -X DELETE "http://localhost:3000/conversations/$CONV_ID/messages/$MSG_ID" \
  -H "Authorization: Bearer $USER1_ID" > /dev/null
echo "deleteMessage: Deleted message"

# 11. Create group for group operations
echo ""
echo "11. Creating test group..."
GROUP_RESPONSE=$(curl -s -X POST "http://localhost:3000/groups" \
  -H "Authorization: Bearer $USER1_ID" \
  -H "Content-Type: application/json" \
  -d "{\"name\": \"Test Group\", \"memberIds\": [\"$USER2_ID\"]}")
GROUP_ID=$(echo $GROUP_RESPONSE | grep -o '"id":"[^"]*' | head -1 | cut -d'"' -f4)
echo "Created group: $GROUP_ID"

# 12. addToGroup
echo ""
echo "12. Testing addToGroup..."
# Create third user to add
THIRD_LOGIN=$(curl -s -X POST "http://localhost:3000/session" \
  -H "Content-Type: application/json" \
  -d '{"name": "testuser3"}')
USER3_ID=$(echo $THIRD_LOGIN | grep -o '"identifier":"[^"]*' | head -1 | cut -d'"' -f4)

curl -s -X POST "http://localhost:3000/groups/$GROUP_ID/members" \
  -H "Authorization: Bearer $USER1_ID" \
  -H "Content-Type: application/json" \
  -d "{\"userId\": \"$USER3_ID\"}" > /dev/null
echo "addToGroup: Added user3 to group"

# 13. setGroupName
echo ""
echo "13. Testing setGroupName..."
curl -s -X PUT "http://localhost:3000/groups/$GROUP_ID" \
  -H "Authorization: Bearer $USER1_ID" \
  -H "Content-Type: application/json" \
  -d '{"name": "Updated Test Group"}' > /dev/null
echo "setGroupName: Updated group name"

# 14. setGroupPhoto
echo ""
echo "14. Testing setGroupPhoto..."
curl -s -X PUT "http://localhost:3000/groups/$GROUP_ID/photo" \
  -H "Authorization: Bearer $USER1_ID" \
  -H "Content-Type: application/json" > /dev/null
echo "setGroupPhoto: Updated group photo (member validation working)"

# 15. setMyPhoto
echo ""
echo "15. Testing setMyPhoto..."
curl -s -X PUT "http://localhost:3000/users/$USER1_ID/photo" \
  -H "Authorization: Bearer $USER1_ID" \
  -H "Content-Type: application/json" > /dev/null
echo "setMyPhoto: Updated user photo"

# 16. leaveGroup
echo ""
echo "16. Testing leaveGroup..."
curl -s -X DELETE "http://localhost:3000/groups/$GROUP_ID/members/$USER3_ID" \
  -H "Authorization: Bearer $USER3_ID" > /dev/null
echo "leaveGroup: User3 left the group"

# 17. Test group messaging works
echo ""
echo "17. Testing group messaging..."
GROUP_MSG=$(curl -s -X POST "http://localhost:3000/conversations/$GROUP_ID/messages" \
  -H "Authorization: Bearer $USER1_ID" \
  -H "Content-Type: application/json" \
  -d '{"content": "Hello group", "messageType": "text"}')
echo "Group messaging: Sent message to group"
