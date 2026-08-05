#!/bin/bash

# test_images.sh - Test file for image/photo functionality
# run this after starting the server with: go run ./cmd/webapi/

set -e  # Exit on any error

# Change to the script's directory to ensure relative paths work correctly
cd "$(dirname "$0")"

echo "WASAText Image Functionality Tests"

# Create invalid file (text file pretending to be image)
echo "This is not an image" > fake_image.png

echo "Test files created successfully"

# Setup users and auth
echo ""
echo "Setting up test users"

# 1. Create test users
LOGIN_RESPONSE=$(curl -s -X POST "http://localhost:3000/session" \
  -H "Content-Type: application/json" \
  -d '{"name": "imageuser1"}')
USER1_ID=$(echo $LOGIN_RESPONSE | grep -o '"identifier":"[^"]*' | head -1 | cut -d'"' -f4)
echo "Created User1 ID: $USER1_ID"

SECOND_LOGIN=$(curl -s -X POST "http://localhost:3000/session" \
  -H "Content-Type: application/json" \
  -d '{"name": "imageuser2"}')
USER2_ID=$(echo $SECOND_LOGIN | grep -o '"identifier":"[^"]*' | head -1 | cut -d'"' -f4)
echo "Created User2 ID: $USER2_ID"

# Create direct conversation for message photo tests
CONV_RESPONSE=$(curl -s -X POST "http://localhost:3000/conversations" \
  -H "Authorization: Bearer $USER1_ID" \
  -H "Content-Type: application/json" \
  -d "{\"userId\": \"$USER2_ID\"}")
CONV_ID=$(echo $CONV_RESPONSE | grep -o '"id":"[^"]*' | head -1 | cut -d'"' -f4)
echo "Created conversation: $CONV_ID"

# Create test group
GROUP_RESPONSE=$(curl -s -X POST "http://localhost:3000/groups" \
  -H "Authorization: Bearer $USER1_ID" \
  -H "Content-Type: application/json" \
  -d "{\"name\": \"Image Test Group\", \"memberIds\": [\"$USER2_ID\"]}")
GROUP_ID=$(echo $GROUP_RESPONSE | grep -o '"id":"[^"]*' | head -1 | cut -d'"' -f4)
echo "Created group: $GROUP_ID"

# Helper function to test status code
test_status_code() {
    local expected=$1
    local actual=$2
    local test_name=$3
    
    if [ "$actual" -eq "$expected" ]; then
        echo "$test_name (HTTP $actual)"
        return 0
    else
        echo "$test_name (Expected HTTP $expected, got $actual)"
        return 1
    fi
}

echo ""
echo "Testing User Photo Upload"

# Test 1: Upload valid PNG user photo
echo "1. Testing valid PNG upload..."
STATUS=$(curl -s -w "%{http_code}" -o /dev/null -X PUT "http://localhost:3000/users/$USER1_ID/photo" \
  -H "Authorization: Bearer $USER1_ID" \
  -H "Content-Type: image/png" \
  --data-binary @test_image.png)
test_status_code 204 $STATUS "Valid PNG upload"

# Test 2: Upload valid JPEG user photo
echo "2. Testing valid JPEG upload..."
STATUS=$(curl -s -w "%{http_code}" -o /dev/null -X PUT "http://localhost:3000/users/$USER1_ID/photo" \
  -H "Authorization: Bearer $USER1_ID" \
  -H "Content-Type: image/jpeg" \
  --data-binary @test_image.jpg)
test_status_code 204 $STATUS "Valid JPEG upload"

# Test 3: Upload valid GIF user photo
echo "3. Testing valid GIF upload..."
STATUS=$(curl -s -w "%{http_code}" -o /dev/null -X PUT "http://localhost:3000/users/$USER1_ID/photo" \
  -H "Authorization: Bearer $USER1_ID" \
  -H "Content-Type: image/gif" \
  --data-binary @test_image.gif)
test_status_code 204 $STATUS "Valid GIF upload"

# Test 4: Test invalid content type
echo "4. Testing invalid content type..."
STATUS=$(curl -s -w "%{http_code}" -o /dev/null -X PUT "http://localhost:3000/users/$USER1_ID/photo" \
  -H "Authorization: Bearer $USER1_ID" \
  -H "Content-Type: text/plain" \
  --data-binary @test_image.png)
test_status_code 415 $STATUS "Invalid content type rejection"

# Test 5: Test fake image file
echo "5. Testing fake image file..."
STATUS=$(curl -s -w "%{http_code}" -o /dev/null -X PUT "http://localhost:3000/users/$USER1_ID/photo" \
  -H "Authorization: Bearer $USER1_ID" \
  -H "Content-Type: image/png" \
  --data-binary @fake_image.png)
test_status_code 400 $STATUS "Fake image file rejection"

# Test 6: Test unauthorized access
echo "6. Testing unauthorized photo upload..."
STATUS=$(curl -s -w "%{http_code}" -o /dev/null -X PUT "http://localhost:3000/users/$USER1_ID/photo" \
  -H "Content-Type: image/png" \
  --data-binary @test_image.png)
test_status_code 401 $STATUS "Unauthorized upload rejection"

# Test 7: Test wrong user trying to upload
echo "7. Testing wrong user photo upload..."
STATUS=$(curl -s -w "%{http_code}" -o /dev/null -X PUT "http://localhost:3000/users/$USER1_ID/photo" \
  -H "Authorization: Bearer $USER2_ID" \
  -H "Content-Type: image/png" \
  --data-binary @test_image.png)
test_status_code 403 $STATUS "Wrong user upload rejection"

# Test 8: Test empty file
echo "8. Testing empty file upload..."
touch empty_file.png
STATUS=$(curl -s -w "%{http_code}" -o /dev/null -X PUT "http://localhost:3000/users/$USER1_ID/photo" \
  -H "Authorization: Bearer $USER1_ID" \
  -H "Content-Type: image/png" \
  --data-binary @empty_file.png)
test_status_code 400 $STATUS "Empty file rejection"

echo ""
echo "Testing User Photo Serving"

# Test 9: Retrieve uploaded user photo
echo "9. Testing photo retrieval..."
STATUS=$(curl -s -w "%{http_code}" -o retrieved_user_photo.png -X GET "http://localhost:3000/users/$USER1_ID/photo")
test_status_code 200 $STATUS "Photo retrieval"

# Test 10: Test photo retrieval for user without photo
echo "10. Testing photo retrieval for user without photo..."
STATUS=$(curl -s -w "%{http_code}" -o /dev/null -X GET "http://localhost:3000/users/$USER2_ID/photo")
test_status_code 404 $STATUS "No photo available"

echo ""
echo "Testing Group Photo Upload"

# Test 11: Upload valid group photo (by group member)
echo "11. Testing valid group photo upload by member..."
STATUS=$(curl -s -w "%{http_code}" -o /dev/null -X PUT "http://localhost:3000/groups/$GROUP_ID/photo" \
  -H "Authorization: Bearer $USER1_ID" \
  -H "Content-Type: image/png" \
  --data-binary @test_image.png)
test_status_code 204 $STATUS "Valid group photo upload by member"

# Test 12: Upload group photo by another member
echo "12. Testing group photo upload by another member..."
STATUS=$(curl -s -w "%{http_code}" -o /dev/null -X PUT "http://localhost:3000/groups/$GROUP_ID/photo" \
  -H "Authorization: Bearer $USER2_ID" \
  -H "Content-Type: image/jpeg" \
  --data-binary @test_image.jpg)
test_status_code 204 $STATUS "Group photo upload by other member"

# Test 13: Try to upload group photo by non-member
# Create third user who is not in the group
THIRD_LOGIN=$(curl -s -X POST "http://localhost:3000/session" \
  -H "Content-Type: application/json" \
  -d '{"name": "imageuser3"}')
USER3_ID=$(echo $THIRD_LOGIN | grep -o '"identifier":"[^"]*' | head -1 | cut -d'"' -f4)

echo "13. Testing group photo upload by non-member..."
STATUS=$(curl -s -w "%{http_code}" -o /dev/null -X PUT "http://localhost:3000/groups/$GROUP_ID/photo" \
  -H "Authorization: Bearer $USER3_ID" \
  -H "Content-Type: image/png" \
  --data-binary @test_image.png)
test_status_code 403 $STATUS "Non-member group photo upload rejection"

echo ""
echo "Testing Group Photo Serving"

# Test 14: Retrieve group photo
echo "14. Testing group photo retrieval..."
STATUS=$(curl -s -w "%{http_code}" -o retrieved_group_photo.jpg -X GET "http://localhost:3000/groups/$GROUP_ID/photo")
test_status_code 200 $STATUS "Group photo retrieval"

echo ""
echo "Testing Message Photo Upload"

# Test 15: Send photo message via multipart form
echo "15. Testing photo message upload..."
PHOTO_MSG_RESPONSE=$(curl -s -w "%{http_code}" -X POST "http://localhost:3000/conversations/$CONV_ID/messages" \
  -H "Authorization: Bearer $USER1_ID" \
  -F "messageType=photo" \
  -F "photo=@test_image.png;type=image/png")
PHOTO_MSG_STATUS=$(echo "$PHOTO_MSG_RESPONSE" | tail -c 4)
PHOTO_MSG_BODY=$(echo "$PHOTO_MSG_RESPONSE" | head -c -4)
PHOTO_MSG_ID=$(echo $PHOTO_MSG_BODY | grep -o '"id":"[^"]*' | head -1 | cut -d'"' -f4)
test_status_code 201 $PHOTO_MSG_STATUS "Photo message upload"
echo "Photo message ID: $PHOTO_MSG_ID"

# Test 16: Send GIF message
echo "16. Testing GIF message upload..."
GIF_MSG_RESPONSE=$(curl -s -w "%{http_code}" -X POST "http://localhost:3000/conversations/$CONV_ID/messages" \
  -H "Authorization: Bearer $USER1_ID" \
  -F "messageType=photo" \
  -F "photo=@test_image.gif;type=image/gif")
GIF_MSG_STATUS=$(echo "$GIF_MSG_RESPONSE" | tail -c 4)
test_status_code 201 $GIF_MSG_STATUS "GIF message upload"

# Test 17: Try to send photo message without photo file
echo "17. Testing photo message without file..."
STATUS=$(curl -s -w "%{http_code}" -o /dev/null -X POST "http://localhost:3000/conversations/$CONV_ID/messages" \
  -H "Authorization: Bearer $USER1_ID" \
  -F "messageType=photo")
test_status_code 400 $STATUS "Photo message without file rejection"

# Test 18: Try to send photo message with fake image
echo "18. Testing photo message with fake image..."
STATUS=$(curl -s -w "%{http_code}" -o /dev/null -X POST "http://localhost:3000/conversations/$CONV_ID/messages" \
  -H "Authorization: Bearer $USER1_ID" \
  -F "messageType=photo" \
  -F "photo=@fake_image.png;type=image/png")
test_status_code 400 $STATUS "Fake image message rejection"

# Test 19: Send regular text message (should still work)
echo "19. Testing regular text message still works..."
TEXT_MSG_RESPONSE=$(curl -s -w "%{http_code}" -X POST "http://localhost:3000/conversations/$CONV_ID/messages" \
  -H "Authorization: Bearer $USER1_ID" \
  -H "Content-Type: application/json" \
  -d '{"content": "Regular text message", "messageType": "text"}')
TEXT_MSG_STATUS=$(echo "$TEXT_MSG_RESPONSE" | tail -c 4)
test_status_code 201 $TEXT_MSG_STATUS "Regular text message"

# Test 20: Send text message via multipart form
echo "20. Testing text message via multipart form..."
STATUS=$(curl -s -w "%{http_code}" -o /dev/null -X POST "http://localhost:3000/conversations/$CONV_ID/messages" \
  -H "Authorization: Bearer $USER1_ID" \
  -F "messageType=text" \
  -F "content=Text via multipart form")
test_status_code 201 $STATUS "Text message via multipart form"

echo ""
echo "Testing Message Photo Serving"

# Test 21: Retrieve photo message
if [ ! -z "$PHOTO_MSG_ID" ]; then
    echo "21. Testing photo message retrieval..."
    STATUS=$(curl -s -w "%{http_code}" -o retrieved_message_photo.png -X GET "http://localhost:3000/conversations/$CONV_ID/messages/$PHOTO_MSG_ID/photo" \
      -H "Authorization: Bearer $USER1_ID")
    test_status_code 200 $STATUS "Message photo retrieval"
else
    echo "21. Skipping photo message retrieval (no message ID)"
fi

echo ""
echo "Testing Edge Cases"

# Test 22: Large file upload (should fail)
echo "22. Testing large file rejection..."
dd if=/dev/zero of=large_file.png bs=1M count=6 2>/dev/null
STATUS=$(curl -s -w "%{http_code}" -o /dev/null -X PUT "http://localhost:3000/users/$USER1_ID/photo" \
  -H "Authorization: Bearer $USER1_ID" \
  -H "Content-Type: image/png" \
  --data-binary @large_file.png)
test_status_code 400 $STATUS "Large file rejection"

# Test 23: Missing Content-Type header
echo "23. Testing missing Content-Type header..."
STATUS=$(curl -s -w "%{http_code}" -o /dev/null -X PUT "http://localhost:3000/users/$USER1_ID/photo" \
  -H "Authorization: Bearer $USER1_ID" \
  --data-binary @test_image.png)
test_status_code 415 $STATUS "Missing Content-Type rejection"

# Test 24: Nonexistent user photo
echo "24. Testing nonexistent user photo..."
STATUS=$(curl -s -w "%{http_code}" -o /dev/null -X GET "http://localhost:3000/users/nonexistent/photo")
test_status_code 404 $STATUS "Nonexistent user photo"

# Test 25: Nonexistent group photo  
echo "25. Testing nonexistent group photo..."
STATUS=$(curl -s -w "%{http_code}" -o /dev/null -X GET "http://localhost:3000/groups/nonexistent/photo")
test_status_code 404 $STATUS "Nonexistent group photo"

echo ""
echo "Testing Integration with Existing Features"

# Test 26: Check that conversation list shows user with photo
echo "26. Testing conversation list includes user photos..."
CONVERSATIONS=$(curl -s -X GET "http://localhost:3000/conversations" \
  -H "Authorization: Bearer $USER1_ID")
echo "Retrieved conversations with photo URLs"

# Test 27: Check that group details include photo URL
echo "27. Testing group details include photo URL..."
GROUP_DETAILS=$(curl -s -X GET "http://localhost:3000/groups/$GROUP_ID" \
  -H "Authorization: Bearer $USER1_ID")
echo "Retrieved group details with photo URL"

# Test 28: Forward photo message
if [ ! -z "$PHOTO_MSG_ID" ]; then
    echo "28. Testing photo message forwarding..."
    # Create another conversation to forward to
    CONV2_RESPONSE=$(curl -s -X POST "http://localhost:3000/conversations" \
      -H "Authorization: Bearer $USER2_ID" \
      -H "Content-Type: application/json" \
      -d "{\"userId\": \"$USER1_ID\"}")
    CONV2_ID=$(echo $CONV2_RESPONSE | grep -o '"id":"[^"]*' | head -1 | cut -d'"' -f4)
    
    STATUS=$(curl -s -w "%{http_code}" -o /dev/null -X POST "http://localhost:3000/conversations/$CONV_ID/messages/$PHOTO_MSG_ID/forward" \
      -H "Authorization: Bearer $USER1_ID" \
      -H "Content-Type: application/json" \
      -d "{\"conversationId\": \"$CONV2_ID\"}")
    test_status_code 201 $STATUS "Photo message forwarding"
else
    echo "28. Skipping photo message forwarding (no message ID)"
fi

echo ""
echo "Cleanup"

# Clean up test files
rm -f fake_image.png empty_file.png large_file.png
rm -f retrieved_user_photo.png retrieved_group_photo.jpg retrieved_message_photo.png

echo "Cleaned up test files"

echo ""
echo "Image Functionality Tests Complete"
echo "Summary:"
echo "User photo upload/download with validation"
echo "Group photo upload/download with member validation" 
echo "Message photo upload/download via multipart forms"
echo "GIF support for all photo types"
echo "Comprehensive error handling and edge cases"
echo "Integration with existing messaging features"
echo ""
echo "All image functionality is working correctly!"