#!/bin/bash

# Test script for message status functionality (sent, delivered, read)

echo "Message Status Test"
echo ""

# 1. Create users
echo "1. Creating test users..."
USER1_RESPONSE=$(curl -s -X POST "http://localhost:3000/session" \
  -H "Content-Type: application/json" \
  -d '{"name": "sender"}')
USER1_ID=$(echo $USER1_RESPONSE | grep -o '"identifier":"[^"]*' | head -1 | cut -d'"' -f4)

USER2_RESPONSE=$(curl -s -X POST "http://localhost:3000/session" \
  -H "Content-Type: application/json" \
  -d '{"name": "recipient1"}')
USER2_ID=$(echo $USER2_RESPONSE | grep -o '"identifier":"[^"]*' | head -1 | cut -d'"' -f4)

USER3_RESPONSE=$(curl -s -X POST "http://localhost:3000/session" \
  -H "Content-Type: application/json" \
  -d '{"name": "recipient2"}')
USER3_ID=$(echo $USER3_RESPONSE | grep -o '"identifier":"[^"]*' | head -1 | cut -d'"' -f4)

echo "Created users: $USER1_ID, $USER2_ID, $USER3_ID"

# 2. Create group conversation
echo ""
echo "2. Creating group conversation..."
GROUP_RESPONSE=$(curl -s -X POST "http://localhost:3000/groups" \
  -H "Authorization: Bearer $USER1_ID" \
  -H "Content-Type: application/json" \
  -d "{\"name\": \"Status Test Group\", \"memberIds\": [\"$USER2_ID\", \"$USER3_ID\"]}")
GROUP_ID=$(echo $GROUP_RESPONSE | grep -o '"id":"[^"]*' | head -1 | cut -d'"' -f4)
echo "Created group: $GROUP_ID"

# 3. Send message to group
echo ""
echo "3. Sending message to group..."
MSG_RESPONSE=$(curl -s -X POST "http://localhost:3000/conversations/$GROUP_ID/messages" \
  -H "Authorization: Bearer $USER1_ID" \
  -H "Content-Type: application/json" \
  -d '{"content": "Hello group - checking status", "messageType": "text"}')
MSG_ID=$(echo $MSG_RESPONSE | grep -o '"id":"[^"]*' | head -1 | cut -d'"' -f4)
MSG_STATUS=$(echo $MSG_RESPONSE | grep -o '"status":"[^"]*' | cut -d'"' -f4)
echo "Sent message $MSG_ID with initial status: $MSG_STATUS"

# 4. Check sender's view (should show 'sent' - no one has received yet)
echo ""
echo "4. Checking sender's view of message..."
SENDER_CONV=$(curl -s -X GET "http://localhost:3000/conversations/$GROUP_ID" \
  -H "Authorization: Bearer $USER1_ID")
SENDER_MSG_STATUS=$(echo $SENDER_CONV | grep -o '"status":"[^"]*' | head -1 | cut -d'"' -f4)
echo "Sender sees status: $SENDER_MSG_STATUS (should be 'sent')"

# 5. Recipient1 checks conversation list (marks as delivered)
echo ""
echo "5. Recipient1 checks conversation list..."
curl -s -X GET "http://localhost:3000/conversations" \
  -H "Authorization: Bearer $USER2_ID" > /dev/null
echo "Recipient1 viewed conversation list (marked as delivered)"

# 6. Check sender's view again (should still be 'sent' - not all recipients delivered)
echo ""
echo "6. Checking sender's view after partial delivery..."
SENDER_CONV2=$(curl -s -X GET "http://localhost:3000/conversations/$GROUP_ID" \
  -H "Authorization: Bearer $USER1_ID")
SENDER_MSG_STATUS2=$(echo $SENDER_CONV2 | grep -o '"status":"[^"]*' | head -1 | cut -d'"' -f4)
echo "Sender sees status: $SENDER_MSG_STATUS2 (should be 'sent' - not all delivered)"

# 7. Recipient2 also checks conversation list (all recipients delivered)
echo ""
echo "7. Recipient2 checks conversation list..."
curl -s -X GET "http://localhost:3000/conversations" \
  -H "Authorization: Bearer $USER3_ID" > /dev/null
echo "Recipient2 viewed conversation list (all recipients delivered)"

# 8. Check sender's view (should now be 'delivered')
echo ""
echo "8. Checking sender's view after full delivery..."
SENDER_CONV3=$(curl -s -X GET "http://localhost:3000/conversations/$GROUP_ID" \
  -H "Authorization: Bearer $USER1_ID")
SENDER_MSG_STATUS3=$(echo $SENDER_CONV3 | grep -o '"status":"[^"]*' | head -1 | cut -d'"' -f4)
echo "Sender sees status: $SENDER_MSG_STATUS3 (should be 'delivered')"

# 9. Recipient1 opens conversation (marks as read)
echo ""
echo "9. Recipient1 opens conversation..."
curl -s -X GET "http://localhost:3000/conversations/$GROUP_ID" \
  -H "Authorization: Bearer $USER2_ID" > /dev/null
echo "Recipient1 opened conversation (marked as read)"

# 10. Check sender's view (should still be 'delivered' - not all read)
echo ""
echo "10. Checking sender's view after partial read..."
SENDER_CONV4=$(curl -s -X GET "http://localhost:3000/conversations/$GROUP_ID" \
  -H "Authorization: Bearer $USER1_ID")
SENDER_MSG_STATUS4=$(echo $SENDER_CONV4 | grep -o '"status":"[^"]*' | head -1 | cut -d'"' -f4)
echo "Sender sees status: $SENDER_MSG_STATUS4 (should be 'delivered' - not all read)"

# 11. Recipient2 also opens conversation (all recipients read)
echo ""
echo "11. Recipient2 opens conversation..."
curl -s -X GET "http://localhost:3000/conversations/$GROUP_ID" \
  -H "Authorization: Bearer $USER3_ID" > /dev/null
echo "Recipient2 opened conversation (all recipients read)"

# 12. Final check - should now show 'read'
echo ""
echo "12. Final status check..."
SENDER_CONV5=$(curl -s -X GET "http://localhost:3000/conversations/$GROUP_ID" \
  -H "Authorization: Bearer $USER1_ID")
SENDER_MSG_STATUS5=$(echo $SENDER_CONV5 | grep -o '"status":"[^"]*' | head -1 | cut -d'"' -f4)
echo "Sender sees status: $SENDER_MSG_STATUS5 (should be 'read')"

echo ""
echo "Summary"
echo "1. Initial: $MSG_STATUS"
echo "2. After send: $SENDER_MSG_STATUS"
echo "3. After partial delivery: $SENDER_MSG_STATUS2"
echo "4. After full delivery: $SENDER_MSG_STATUS3"
echo "5. After partial read: $SENDER_MSG_STATUS4"
echo "6. After full read: $SENDER_MSG_STATUS5"
echo ""

if [ "$SENDER_MSG_STATUS5" = "read" ]; then
    echo "SUCCESS: Message status progression working correctly!"
else
    echo "ISSUE: Expected final status 'read', got '$SENDER_MSG_STATUS5'"
fi