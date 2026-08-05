#!/bin/bash

# Test message status for direct (1-on-1) conversations

echo "Direct Message Status Test"
echo ""

# 1. Create users
USER1_RESPONSE=$(curl -s -X POST "http://localhost:3000/session" \
  -H "Content-Type: application/json" \
  -d '{"name": "alice"}')
USER1_ID=$(echo $USER1_RESPONSE | grep -o '"identifier":"[^"]*' | head -1 | cut -d'"' -f4)

USER2_RESPONSE=$(curl -s -X POST "http://localhost:3000/session" \
  -H "Content-Type: application/json" \
  -d '{"name": "bob"}')
USER2_ID=$(echo $USER2_RESPONSE | grep -o '"identifier":"[^"]*' | head -1 | cut -d'"' -f4)

echo "Created users: Alice ($USER1_ID), Bob ($USER2_ID)"

# 2. Create direct conversation
CONV_RESPONSE=$(curl -s -X POST "http://localhost:3000/conversations" \
  -H "Authorization: Bearer $USER1_ID" \
  -H "Content-Type: application/json" \
  -d "{\"userId\": \"$USER2_ID\"}")
CONV_ID=$(echo $CONV_RESPONSE | grep -o '"id":"[^"]*' | head -1 | cut -d'"' -f4)
echo "Created direct conversation: $CONV_ID"

# 3. Alice sends message to Bob
MSG_RESPONSE=$(curl -s -X POST "http://localhost:3000/conversations/$CONV_ID/messages" \
  -H "Authorization: Bearer $USER1_ID" \
  -H "Content-Type: application/json" \
  -d '{"content": "Hi Bob!", "messageType": "text"}')
MSG_STATUS=$(echo $MSG_RESPONSE | grep -o '"status":"[^"]*' | cut -d'"' -f4)
echo "Alice sent message with status: $MSG_STATUS"

# 4. Check Alice's view (should be 'sent')
ALICE_VIEW1=$(curl -s -X GET "http://localhost:3000/conversations/$CONV_ID" \
  -H "Authorization: Bearer $USER1_ID")
ALICE_STATUS1=$(echo $ALICE_VIEW1 | grep -o '"status":"[^"]*' | head -1 | cut -d'"' -f4)
echo "Alice sees: $ALICE_STATUS1 (should be 'sent')"

# 5. Bob checks conversation list (delivered)
curl -s -X GET "http://localhost:3000/conversations" \
  -H "Authorization: Bearer $USER2_ID" > /dev/null
echo "Bob checked conversation list"

# 6. Check Alice's view (should be 'delivered')
ALICE_VIEW2=$(curl -s -X GET "http://localhost:3000/conversations/$CONV_ID" \
  -H "Authorization: Bearer $USER1_ID")
ALICE_STATUS2=$(echo $ALICE_VIEW2 | grep -o '"status":"[^"]*' | head -1 | cut -d'"' -f4)
echo "Alice sees: $ALICE_STATUS2 (should be 'delivered')"

# 7. Bob opens conversation (read)
curl -s -X GET "http://localhost:3000/conversations/$CONV_ID" \
  -H "Authorization: Bearer $USER2_ID" > /dev/null
echo "Bob opened conversation"

# 8. Final check - should be 'read'
ALICE_VIEW3=$(curl -s -X GET "http://localhost:3000/conversations/$CONV_ID" \
  -H "Authorization: Bearer $USER1_ID")
ALICE_STATUS3=$(echo $ALICE_VIEW3 | grep -o '"status":"[^"]*' | head -1 | cut -d'"' -f4)
echo "Alice sees: $ALICE_STATUS3 (should be 'read')"

echo ""
echo "Summary"
echo "1. After send: $ALICE_STATUS1"
echo "2. After Bob's delivery: $ALICE_STATUS2" 
echo "3. After Bob's read: $ALICE_STATUS3"

if [ "$ALICE_STATUS3" = "read" ]; then
    echo "SUCCESS: Direct message status working!"
else
    echo "ISSUE: Expected 'read', got '$ALICE_STATUS3'"
fi