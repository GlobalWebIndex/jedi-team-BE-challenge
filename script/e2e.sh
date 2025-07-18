#!/bin/bash

# Define the backend API URL
BASE_URL="http://localhost:8080"
TOKEN_ENDPOINT="/token"
USER_ID="12345678-0000-0000-0000-000000000000"

# Step 1: Get the token
response_token=$(curl -s --location "$BASE_URL$TOKEN_ENDPOINT" \
  --header 'Content-Type: application/json' \
  --data '{"username": "user", "password": "password"}')

token=$(echo "$response_token" | jq -r '.token')

if [ -z "$token" ]; then
  echo "Error: Failed to retrieve token."
  exit 1
fi

echo "Token retrieved successfully: $token"
echo "----------------------------------"

# Step 2: Create 3 chat sessions for the user and get the first session's ID
for i in {1..3}; do
  response=$(curl -s --location --request POST "$BASE_URL/users/$USER_ID/chat-sessions" \
    --header "Authorization: Bearer $token")

  if [ $i -eq 1 ]; then
    CHAT_SESSION_ID=$(echo "$response" | jq -r '.id')
    echo "First chat session created with ID: $CHAT_SESSION_ID"
  fi
done

echo "----------------------------------"

# Step 3: Send the first message
message_content_1="what do you know about latino mobile gamers"
response_message_1=$(curl -s --location "$BASE_URL/users/$USER_ID/chat-sessions/$CHAT_SESSION_ID/messages" \
  --header "Content-Type: application/json" \
  --header "Authorization: Bearer $token" \
  --data "{\"content\":\"$message_content_1\"}")

# Step 4: Send the second message
message_content_2="do they use social media"
response_message_2=$(curl -s --location "$BASE_URL/users/$USER_ID/chat-sessions/$CHAT_SESSION_ID/messages" \
  --header "Content-Type: application/json" \
  --header "Authorization: Bearer $token" \
  --data "{\"content\":\"$message_content_2\"}")

# Step 5: Send the third message
message_content_3="what social media do they use the most"
response_message_3=$(curl -s --location "$BASE_URL/users/$USER_ID/chat-sessions/$CHAT_SESSION_ID/messages" \
  --header "Content-Type: application/json" \
  --header "Authorization: Bearer $token" \
  --data "{\"content\":\"$message_content_3\"}")

system_message_id_3=$(echo "$response_message_3" | jq -r '.systemMessage.id')

echo "Three messages were sent in the same discussion and title was generated"
echo "----------------------------------"

# Step 5.5: Send three messages with websockets

messages=(
  "what do you know about australian millennial parents"
  "do they use social media"
  "what social media do they use the most"
)

echo "Connecting to WebSocket..."
for msg in "${messages[@]}"; do
  echo "Sending message: $msg"

  # Send JSON and read one response line
  response=$(echo "{\"content\":\"$msg\"}" |
    websocat --header="Authorization: Bearer $token" --text "ws://localhost:8080/ws/users/$USER_ID/chat-sessions/$CHAT_SESSION_ID/messages")

  echo "Received response:"
  echo "$response"
done

echo "----------------------------------"

# Step 6: Fetch and display the whole chat session
echo "Fetching the entire chat session $CHAT_SESSION_ID..."
chat_session_response=$(curl -s --location "$BASE_URL/chat-sessions/$CHAT_SESSION_ID" \
  --header "Authorization: Bearer $token")

echo "Full Chat Session $CHAT_SESSION_ID:"
echo "$chat_session_response"

echo "----------------------------------"

# Step 7: Submit feedback for the third message
feedback_message="The force is not strong with that message."
feedback_response=$(curl -s --location --write-out "%{http_code}" --request POST "$BASE_URL/users/$USER_ID/chat-sessions/$CHAT_SESSION_ID/messages/$system_message_id_3/feedback" \
  --header "Authorization: Bearer $token" \
  --data "{\"feedback\":\"$feedback_message\"}")

if [ "$feedback_response" -eq 201 ]; then
  echo "Negative feedback OK: received 201 status code"
else
  echo "Error: Feedback submission failed. Status code: $feedback_response"
  exit 1
fi

echo "----------------------------------"

# Step 8: Send the final and failed message
message_content_4="what are butterflies"
response_message_4=$(curl -s --location "$BASE_URL/users/$USER_ID/chat-sessions/$CHAT_SESSION_ID/messages" \
  --header "Content-Type: application/json" \
  --header "Authorization: Bearer $token" \
  --data "{\"content\":\"$message_content_4\"}")

echo "Message 4 that's supposed to be not answered sent to chat session $CHAT_SESSION_ID:"
echo "$response_message_4"
