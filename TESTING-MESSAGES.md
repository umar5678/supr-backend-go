# Messaging System - Testing Guide

## Quick Start Testing

### **Option 1: Using Postman (Recommended)**

#### Step 1: Start the Server
```bash
cd e:\final_go_backend\supr-backend-go
go run ./cmd/api
```

#### Step 2: Get an Authentication Token
First, you need to authenticate as a user (rider or driver).

**Sign up as Rider:**
```
POST http://localhost:8080/api/v1/auth/phone/signup
Content-Type: application/json

{
  "phone": "+1234567890",
  "password": "password123",
  "name": "John Rider"
}
```

**Sign up as Driver:**
```
POST http://localhost:8080/api/v1/drivers/register
Content-Type: application/json
Authorization: Bearer {your_token}

{
  "phone": "+1987654321",
  "name": "Jane Driver",
  "licenseNumber": "DL123456",
  "vehicleType": "cab_economy"
}
```

**Login:**
```
POST http://localhost:8080/api/v1/auth/phone/login
Content-Type: application/json

{
  "phone": "+1234567890",
  "password": "password123"
}

Response:
{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "refreshToken": "eyJhbGciOiJIUzI1NiIs..."
}
```

Copy the `token` - you'll use this in the Authorization header.

---

### **Testing the Messaging Endpoints**

#### **1. Create a Ride (Prerequisite)**

```bash
POST http://localhost:8080/api/v1/rides
Content-Type: application/json
Authorization: Bearer {rider_token}

{
  "pickupLocation": {
    "latitude": 40.7128,
    "longitude": -74.0060,
    "address": "Times Square, NYC"
  },
  "dropoffLocation": {
    "latitude": 40.7580,
    "longitude": -73.9855,
    "address": "Central Park, NYC"
  },
  "vehicleType": "cab_economy"
}

Response:
{
  "id": "ride_abc123",
  "status": "requested",
  ...
}
```

Save the `ride_id` from the response - you'll need it for messaging.

---

#### **2. Send a Message**

```bash
POST http://localhost:8080/api/v1/messages
Content-Type: application/json
Authorization: Bearer {rider_token}

{
  "rideId": "ride_abc123",
  "content": "Hi, I'm ready to go!",
  "metadata": {
    "type": "text",
    "priority": "normal"
  }
}

Response (201 Created):
{
  "success": true,
  "data": {
    "id": "msg_1234567890",
    "rideId": "ride_abc123",
    "senderId": "user_xyz",
    "senderName": "John Rider",
    "senderType": "rider",
    "messageType": "text",
    "content": "Hi, I'm ready to go!",
    "isRead": false,
    "createdAt": "2026-01-22T22:45:00Z"
  },
  "message": "Message sent successfully"
}
```

---

#### **3. Get All Messages for a Ride**

```bash
GET http://localhost:8080/api/v1/messages/rides/ride_abc123?limit=10&offset=0
Content-Type: application/json
Authorization: Bearer {rider_token}

Response (200 OK):
{
  "success": true,
  "data": [
    {
      "id": "msg_1234567890",
      "rideId": "ride_abc123",
      "senderId": "user_xyz",
      "senderName": "John Rider",
      "senderType": "rider",
      "messageType": "text",
      "content": "Hi, I'm ready to go!",
      "isRead": false,
      "createdAt": "2026-01-22T22:45:00Z"
    }
  ],
  "message": "Messages retrieved successfully"
}
```

---

#### **4. Get Unread Message Count**

```bash
GET http://localhost:8080/api/v1/messages/rides/ride_abc123/unread-count
Authorization: Bearer {driver_token}

Response (200 OK):
{
  "success": true,
  "data": {
    "rideId": "ride_abc123",
    "unreadCount": 1
  },
  "message": "Unread count retrieved successfully"
}
```

---

#### **5. Mark Message as Read**

```bash
POST http://localhost:8080/api/v1/messages/msg_1234567890/read
Authorization: Bearer {driver_token}

Response (200 OK):
{
  "success": true,
  "data": null,
  "message": "Message marked as read"
}
```

---

#### **6. Delete a Message**

> ⚠️ Can only delete within 5 minutes of creation by the sender

```bash
DELETE http://localhost:8080/api/v1/messages/msg_1234567890
Authorization: Bearer {rider_token}

Response (200 OK):
{
  "success": true,
  "data": null,
  "message": "Message deleted successfully"
}
```

---

## **Option 2: Using cURL (Command Line)**

### **Full Testing Flow**

#### Step 1: Start Server
```bash
cd e:\final_go_backend\supr-backend-go
go run ./cmd/api
```

#### Step 2: Create Users & Get Tokens

**Rider Token:**
```bash
curl -X POST http://localhost:8080/api/v1/auth/phone/signup \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "+1234567890",
    "password": "password123",
    "name": "John Rider"
  }' | jq '.token' > rider_token.txt
```

**Driver Token:**
```bash
curl -X POST http://localhost:8080/api/v1/auth/phone/signup \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "+1987654321",
    "password": "password123",
    "name": "Jane Driver"
  }' | jq '.token' > driver_token.txt
```

#### Step 3: Create a Ride

```bash
RIDER_TOKEN=$(cat rider_token.txt)

curl -X POST http://localhost:8080/api/v1/rides \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $RIDER_TOKEN" \
  -d '{
    "pickupLocation": {
      "latitude": 40.7128,
      "longitude": -74.0060,
      "address": "Times Square"
    },
    "dropoffLocation": {
      "latitude": 40.7580,
      "longitude": -73.9855,
      "address": "Central Park"
    },
    "vehicleType": "cab_economy"
  }' | jq '.data.id' > ride_id.txt
```

#### Step 4: Send Messages

```bash
RIDER_TOKEN=$(cat rider_token.txt)
RIDE_ID=$(cat ride_id.txt)

# Rider sends message
curl -X POST http://localhost:8080/api/v1/messages \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $RIDER_TOKEN" \
  -d "{
    \"rideId\": \"$RIDE_ID\",
    \"content\": \"Where are you?\",
    \"metadata\": {\"type\": \"text\"}
  }"
```

#### Step 5: Get Messages

```bash
RIDER_TOKEN=$(cat rider_token.txt)
RIDE_ID=$(cat ride_id.txt)

curl -X GET "http://localhost:8080/api/v1/messages/rides/$RIDE_ID?limit=10&offset=0" \
  -H "Authorization: Bearer $RIDER_TOKEN" | jq '.'
```

#### Step 6: Get Unread Count

```bash
DRIVER_TOKEN=$(cat driver_token.txt)
RIDE_ID=$(cat ride_id.txt)

curl -X GET "http://localhost:8080/api/v1/messages/rides/$RIDE_ID/unread-count" \
  -H "Authorization: Bearer $DRIVER_TOKEN" | jq '.'
```

---

## **Option 3: Using REST Client (VS Code Extension)**

### Create `test-messages.http` file

```http
### Variables
@host = http://localhost:8080
@riderToken = your_rider_token_here
@driverToken = your_driver_token_here
@rideId = your_ride_id_here
@messageId = your_message_id_here

### 1. Send Message (Rider)
POST {{host}}/api/v1/messages
Content-Type: application/json
Authorization: Bearer {{riderToken}}

{
  "rideId": "{{rideId}}",
  "content": "I'm waiting at the pickup location",
  "metadata": {
    "type": "text",
    "priority": "high"
  }
}

### 2. Get Messages for Ride
GET {{host}}/api/v1/messages/rides/{{rideId}}?limit=20&offset=0
Authorization: Bearer {{riderToken}}

### 3. Get Unread Count (Driver)
GET {{host}}/api/v1/messages/rides/{{rideId}}/unread-count
Authorization: Bearer {{driverToken}}

### 4. Mark Message as Read
POST {{host}}/api/v1/messages/{{messageId}}/read
Authorization: Bearer {{driverToken}}

### 5. Delete Message (within 5 minutes)
DELETE {{host}}/api/v1/messages/{{messageId}}
Authorization: Bearer {{riderToken}}

### 6. Send Message (Driver)
POST {{host}}/api/v1/messages
Content-Type: application/json
Authorization: Bearer {{driverToken}}

{
  "rideId": "{{rideId}}",
  "content": "I'm 2 minutes away, arriving soon!",
  "metadata": {
    "type": "status"
  }
}
```

---

## **Test Scenarios**

### **Scenario 1: Basic Message Exchange**
1. ✅ Rider sends "Where are you?"
2. ✅ Driver receives message (unread count = 1)
3. ✅ Driver marks as read
4. ✅ Verify unread count = 0

### **Scenario 2: Pagination**
1. ✅ Send 10 messages
2. ✅ Get first 5 with `limit=5&offset=0`
3. ✅ Get next 5 with `limit=5&offset=5`

### **Scenario 3: Error Handling**
1. ✅ Send message without rideId → Should fail with error
2. ✅ Send message without content → Should fail with error
3. ✅ Try to delete message after 5 minutes → Should fail
4. ✅ Try to mark non-existent message as read → Should fail
5. ✅ Send message without token → Should get 401 Unauthorized

### **Scenario 4: Security**
1. ✅ Rider can only delete their own messages
2. ✅ Driver cannot delete rider's message
3. ✅ Can only access messages from rides they're in

---

## **Expected Response Formats**

### Success Response (200/201)
```json
{
  "success": true,
  "data": { /* response data */ },
  "message": "Operation successful"
}
```

### Error Response (400/500)
```json
{
  "success": false,
  "error": "Error message",
  "code": "ERROR_CODE"
}
```

### Empty List Response
```json
{
  "success": true,
  "data": [],
  "message": "Messages retrieved successfully"
}
```

---

## **Database Verification (SQL)**

### Check if table was created
```sql
SELECT * FROM information_schema.tables 
WHERE table_name = 'ride_messages';
```

### Check messages in DB
```sql
SELECT id, ride_id, sender_id, content, is_read, created_at 
FROM ride_messages 
ORDER BY created_at DESC;
```

### Check unread messages for a ride
```sql
SELECT COUNT(*) as unread_count
FROM ride_messages
WHERE ride_id = 'ride_abc123'
  AND is_read = false
  AND deleted_at IS NULL;
```

---

## **Debugging Tips**

### Check Server Logs
Watch the terminal where the server is running for:
- ✅ "failed to send message" → Service error
- ✅ "failed to get messages" → Repository error
- ✅ Error in middleware → Authentication issue

### Check Database Connection
```bash
psql -h localhost -d go_backend -U postgres -c "SELECT 1"
```

### Test with Invalid Token
```bash
curl -X GET "http://localhost:8080/api/v1/messages/rides/ride_123/unread-count" \
  -H "Authorization: Bearer invalid_token"

# Should return 401 Unauthorized
```

### Test Pagination Edge Cases
```bash
# Test limit > 100 (should cap at 100)
GET /api/v1/messages/rides/ride_123?limit=200&offset=0

# Test negative offset (should start at 0)
GET /api/v1/messages/rides/ride_123?limit=10&offset=-5

# Test huge offset
GET /api/v1/messages/rides/ride_123?limit=10&offset=999999
```

---

## **Performance Testing**

### Send 100 messages rapidly
```bash
for i in {1..100}; do
  curl -X POST http://localhost:8080/api/v1/messages \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $RIDER_TOKEN" \
    -d "{
      \"rideId\": \"$RIDE_ID\",
      \"content\": \"Message $i\",
      \"metadata\": {\"index\": $i}
    }"
  echo "Sent message $i"
done
```

### Measure response time
```bash
time curl -X GET "http://localhost:8080/api/v1/messages/rides/ride_123?limit=50&offset=0" \
  -H "Authorization: Bearer $TOKEN"
```

---

## **Troubleshooting**

| Issue | Solution |
|-------|----------|
| 401 Unauthorized | Verify token is valid and not expired |
| 400 Bad Request | Check required fields (rideId, content) |
| 404 Not Found | Verify ride_id and message_id exist |
| 500 Internal Error | Check server logs for detailed error |
| Messages not appearing | Verify ride exists in rides table |
| Sender name is empty | Check users table has correct user data |

---

## **Next: WebSocket Real-Time Testing**

Once REST API is working, you can add WebSocket for real-time updates:

```javascript
// Connect to WebSocket
const ws = new WebSocket('ws://localhost:8080/ws/connect?token=<token>');

// Listen for messages
ws.onmessage = (event) => {
  const msg = JSON.parse(event.data);
  if (msg.type === 'message') {
    console.log('New message:', msg.data);
  }
};

// Send message via WebSocket
ws.send(JSON.stringify({
  type: 'message',
  rideId: 'ride_123',
  content: 'Hello!'
}));
```
