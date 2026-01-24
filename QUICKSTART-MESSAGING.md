# Real-Time Messaging - Quick Start Guide

Get up and running with real-time messaging in 5 minutes.

## 1. Start the Server

```bash
cd e:\final_go_backend\supr-backend-go

# Option A: Run compiled binary
.\bin\go-backend

# Option B: Build and run
go build -o bin/go-backend ./cmd/api && .\bin\go-backend
```

Expected output:
```
2025-01-22T10:00:00Z INFO websocket system initialized successfully
2025-01-22T10:00:00Z INFO server started on port 8080
```

## 2. Get Authentication Token

```bash
# Login (adjust email/password for your setup)
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123"
  }'

# Copy the token from response
# Export as environment variable (for convenience)
export TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VySWQiOiI5YjhjZTIxYy01YTk5LTRiNGItYmZkNi1mNGMzZjY0ODc5MDUiLCJyb2xlIjoicmlkZXIiLCJleHAiOjE3Njk0ODIzMTUsIm5iZiI6MTc2OTEwNDMxNSwiaWF0IjoxNzY5MTA0MzE1fQ.v-iGjo1-9jSNlh76qaIabU2xSL3ckgWnlx9B6iKq-e0"
```

## 3. Test REST API Messaging (Polling)

```bash
# 1. Send a message
curl -X POST http://localhost:8080/api/v1/messages/send \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "rideId": "ride-test-123",
    "content": "Hello, this is a test message!",
    "messageType": "text",
    "metadata": {"from": "postman"}
  }'

# Response includes messageId:
# {
#   "id": "msg-abc123...",
#   "rideId": "ride-test-123",
#   "content": "Hello, this is a test message!",
#   ...
# }
```

Save the `id` as MESSAGE_ID for later use.

```bash
# 2. Get messages in a ride
curl http://localhost:8080/api/v1/messages/ride/ride-test-123 \
  -H "Authorization: Bearer $TOKEN"

# 3. Get unread count
curl http://localhost:8080/api/v1/messages/unread/count \
  -H "Authorization: Bearer $TOKEN"

# 4. Mark message as read
curl -X POST http://localhost:8080/api/v1/messages/$MESSAGE_ID/read \
  -H "Authorization: Bearer $TOKEN"

# 5. Delete message (within 5 minutes of sending)
curl -X DELETE http://localhost:8080/api/v1/messages/$MESSAGE_ID \
  -H "Authorization: Bearer $TOKEN"
```

## 4. Test WebSocket Real-Time Messaging

### Option A: Using wscat (Easiest)

```bash
# Install wscat if not already installed
npm install -g wscat

# Connect to WebSocket
wscat -c "ws://localhost:8080/ws/connect?token=$TOKEN"

# Wait for connection message, then in the wscat prompt (>) send JSON:

# Send a message
> {"type":"message:send","data":{"rideId":"ride-test-123","content":"Hello from WebSocket!","messageType":"text"}}

# Mark as read
> {"type":"message:read","data":{"messageId":"msg-abc123","rideId":"ride-test-123"}}

# Show typing indicator
> {"type":"message:typing","data":{"rideId":"ride-test-123","isTyping":true}}

# Stop typing
> {"type":"message:typing","data":{"rideId":"ride-test-123","isTyping":false}}

# Go online
> {"type":"presence:online","data":{"rideId":"ride-test-123"}}

# Go offline
> {"type":"presence:offline","data":{"rideId":"ride-test-123"}}

# Delete message (within 5 minutes)
> {"type":"message:delete","data":{"messageId":"msg-abc123","rideId":"ride-test-123"}}

# Press Ctrl+C to disconnect
```

**IMPORTANT for PowerShell users:** The `>` is the **wscat prompt** (appears automatically), not a shell operator.

```powershell
# Example for PowerShell:
$TOKEN = "your_jwt_token"
wscat -c "ws://localhost:8080/ws/connect?token=$TOKEN"

# Once connected, paste the JSON (wscat shows > automatically):
{"type":"message:send","data":{"rideId":"ride-test-123","content":"Hello!","messageType":"text"}}
```

# Press Ctrl+C to disconnect
```

### Option B: Using Node.js

Create file `test-ws.js`:

```javascript
const WebSocket = require('ws');
const token = process.env.TOKEN;

console.log('Connecting to WebSocket...');
const ws = new WebSocket(`ws://localhost:8080/ws/connect?token=${token}`);

ws.on('open', () => {
  console.log('✓ Connected to WebSocket\n');
  
  // Send a message
  console.log('→ Sending message...');
  ws.send(JSON.stringify({
    type: 'message:send',
    data: {
      rideId: 'ride-test-123',
      content: 'Hello from Node.js WebSocket!',
      messageType: 'text'
    }
  }));
});

ws.on('message', (data) => {
  const msg = JSON.parse(data);
  console.log('← Received:', msg.type);
  if (msg.data?.message?.content) {
    console.log('  Message:', msg.data.message.content);
  }
});

ws.on('error', (error) => {
  console.error('✗ Error:', error.message);
});

ws.on('close', () => {
  console.log('✗ Disconnected from WebSocket');
});

// Keep process alive
setTimeout(() => {
  console.log('\n← Showing typing indicator...');
  ws.send(JSON.stringify({
    type: 'message:typing',
    data: { rideId: 'ride-test-123', isTyping: true }
  }));
  
  setTimeout(() => {
    console.log('← Going online...');
    ws.send(JSON.stringify({
      type: 'presence:online',
      data: { rideId: 'ride-test-123' }
    }));
    
    setTimeout(() => {
      ws.close();
    }, 2000);
  }, 2000);
}, 2000);
```

Run it:
```bash
npm install ws
node test-ws.js
```

### Option C: Using Chrome DevTools Console

```javascript
// In browser console (e.g., on any page at localhost:8080)
const token = 'YOUR_TOKEN_HERE';
const ws = new WebSocket(`ws://localhost:8080/ws/connect?token=${token}`);

ws.onopen = () => console.log('Connected');

ws.onmessage = (e) => {
  const msg = JSON.parse(e.data);
  console.log('Message received:', msg);
};

ws.onerror = (e) => console.error('Error:', e);

// Send message
ws.send(JSON.stringify({
  type: 'message:send',
  data: {
    rideId: 'ride-test-123',
    content: 'From browser console!',
    messageType: 'text'
  }
}));
```

## 5. Test Multiple Clients (Real-Time Sync)

Open TWO terminal windows:

**Terminal 1 (Client A):**
```bash
export TOKEN="token_from_user_A"
wscat -c "ws://localhost:8080/ws/connect?token=$TOKEN"

# Wait in the prompt
```

**Terminal 2 (Client B):**
```bash
export TOKEN="token_from_user_B"
wscat -c "ws://localhost:8080/ws/connect?token=$TOKEN"

# Send message
> {"type":"message:send","data":{"rideId":"ride-test-123","content":"Message from Client B","messageType":"text"}}
```

**Back to Terminal 1:**
```bash
# You should see Client B's message appear automatically!
< {"type":"message:new","data":{"message":{...},"rideId":"ride-test-123"}}

# Send a reply
> {"type":"message:send","data":{"rideId":"ride-test-123","content":"Reply from Client A","messageType":"text"}}
```

**Back to Terminal 2:**
```bash
# You should see Client A's reply appear automatically!
< {"type":"message:new","data":{"message":{...},"rideId":"ride-test-123"}}
```

🎉 **Real-time messaging is working!**

## 6. Verify Database Persistence

```bash
# Connect to PostgreSQL
psql postgresql://username:password@localhost:5432/supr_backend

# Check messages were saved
SELECT id, ride_id, sender_id, content, is_read, created_at 
FROM ride_messages 
WHERE ride_id = 'ride-test-123' 
ORDER BY created_at DESC;

# Check for soft-deleted messages
SELECT id, content, deleted_at 
FROM ride_messages 
WHERE deleted_at IS NOT NULL;

# Count unread messages
SELECT COUNT(*) 
FROM ride_messages 
WHERE is_read = false AND ride_id = 'ride-test-123';
```

## Common Commands Reference

### REST API

```bash
# Send message
POST /api/v1/messages/send
Body: { rideId, content, messageType, metadata }

# Get messages
GET /api/v1/messages/ride/{rideId}?page=1&limit=20

# Mark as read
POST /api/v1/messages/{messageId}/read

# Delete message
DELETE /api/v1/messages/{messageId}

# Get unread count
GET /api/v1/messages/unread/count
```

### WebSocket Events

```json
// Send message
{"type":"message:send","data":{"rideId":"...","content":"...","messageType":"text"}}

// Mark as read
{"type":"message:read","data":{"messageId":"...","rideId":"..."}}

// Delete message
{"type":"message:delete","data":{"messageId":"...","rideId":"..."}}

// Typing indicator
{"type":"message:typing","data":{"rideId":"...","isTyping":true/false}}

// Go online
{"type":"presence:online","data":{"rideId":"..."}}

// Go offline
{"type":"presence:offline","data":{"rideId":"..."}}
```

## Troubleshooting

**"Connection refused"**
- Server not running on port 8080
- Start with: `.\bin\go-backend`

**"Authentication failed"**
- Invalid JWT token
- Get new token from login endpoint
- Token expired (need new login)

**"WebSocket connection closed"**
- Server crashed or restarted
- Reconnect with exponential backoff
- Check server logs

**Messages not appearing in real-time**
- Different rideId on client vs server
- Client not properly connected (check wscat)
- Check browser console for JavaScript errors

**Database empty**
- Migration not applied
- Check: `\dt ride_messages` in psql
- Run migration if missing

## Next Steps

1. ✅ Test REST API endpoints
2. ✅ Test WebSocket real-time messaging
3. ✅ Test with multiple clients
4. ✅ Verify database persistence

For detailed testing guide, see: `WEBSOCKET-TESTING-GUIDE.md`

For implementation details, see: `REALTIME-MESSAGING-COMPLETE.md`

---

**Status:** Production Ready ✅

All components implemented and tested. Ready for integration with frontend applications!
