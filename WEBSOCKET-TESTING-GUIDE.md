# WebSocket Real-Time Messaging Testing Guide

This guide explains how to test the real-time messaging system via WebSocket connections.

## Architecture Overview

The messaging system now has two layers:

1. **REST API** - For CRUD operations and polling
   - `POST /api/v1/messages/send` - Send a message
   - `GET /api/v1/messages/ride/{rideId}` - Get messages
   - `POST /api/v1/messages/{messageId}/read` - Mark as read
   - `DELETE /api/v1/messages/{messageId}` - Delete message
   - `GET /api/v1/messages/unread/count` - Get unread count

2. **WebSocket** - For real-time event delivery
   - Message sent/received
   - Read receipts
   - Typing indicators
   - Presence (online/offline)

## WebSocket Connection

### Connect to WebSocket

```bash
# With JWT token
ws://localhost:8080/ws/connect?token=YOUR_JWT_TOKEN
```

### Message Format

All WebSocket messages follow this format:

```json
{
  "type": "string",
  "data": {
    // event-specific data
  },
  "timestamp": "2025-01-22T10:00:00Z",
  "requestId": "optional-request-id"
}
```

## Event Types

### 1. Send Message Event

**Type:** `message:send`

Send a message via WebSocket:

```json
{
  "type": "message:send",
  "data": {
    "rideId": "ride-123",
    "content": "Hello, driver!",
    "messageType": "text",
    "metadata": {
      "location": "37.7749,-122.4194"
    }
  }
}
```

**Response (Broadcast to all clients):**

```json
{
  "type": "message:new",
  "data": {
    "message": {
      "id": "msg-456",
      "rideId": "ride-123",
      "senderId": "user-789",
      "senderName": "John Doe",
      "senderType": "rider",
      "content": "Hello, driver!",
      "messageType": "text",
      "metadata": { "location": "37.7749,-122.4194" },
      "isRead": false,
      "timestamp": "2025-01-22T10:00:00Z"
    },
    "rideId": "ride-123"
  }
}
```

### 2. Mark as Read Event

**Type:** `message:read`

Mark a message as read:

```json
{
  "type": "message:read",
  "data": {
    "messageId": "msg-456",
    "rideId": "ride-123"
  }
}
```

**Response (Broadcast):**

```json
{
  "type": "message:read",
  "data": {
    "messageId": "msg-456",
    "rideId": "ride-123",
    "readBy": "user-123",
    "timestamp": "2025-01-22T10:00:05Z"
  }
}
```

### 3. Delete Message Event

**Type:** `message:delete`

Delete a message (5-minute window, same sender only):

```json
{
  "type": "message:delete",
  "data": {
    "messageId": "msg-456",
    "rideId": "ride-123"
  }
}
```

**Response (Broadcast):**

```json
{
  "type": "message:delete",
  "data": {
    "messageId": "msg-456",
    "rideId": "ride-123",
    "deletedBy": "user-789",
    "timestamp": "2025-01-22T10:00:10Z"
  }
}
```

### 4. Typing Indicator Event

**Type:** `message:typing`

Show/hide typing indicator:

```json
{
  "type": "message:typing",
  "data": {
    "rideId": "ride-123",
    "isTyping": true
  }
}
```

**Response (Broadcast):**

```json
{
  "type": "message:typing",
  "data": {
    "rideId": "ride-123",
    "userId": "user-789",
    "isTyping": true
  }
}
```

### 5. Presence Events

**Type:** `presence:online` / `presence:offline`

Send when user comes online:

```json
{
  "type": "presence:online",
  "data": {
    "rideId": "ride-123"
  }
}
```

**Response:**

```json
{
  "type": "presence:online",
  "data": {
    "rideId": "ride-123",
    "userId": "user-789",
    "status": "online",
    "timestamp": "2025-01-22T10:00:00Z"
  }
}
```

## Testing Tools

### 1. WebSocket Test with JavaScript

Create a file `test-websocket.js`:

```javascript
// Get JWT token (from login API first)
const token = 'YOUR_JWT_TOKEN';

// Connect to WebSocket
const ws = new WebSocket(`ws://localhost:8080/ws/connect?token=${token}`);

ws.onopen = () => {
  console.log('Connected to WebSocket');
  
  // Send a message
  ws.send(JSON.stringify({
    type: 'message:send',
    data: {
      rideId: 'ride-123',
      content: 'Hello from WebSocket!',
      messageType: 'text'
    }
  }));
};

ws.onmessage = (event) => {
  console.log('Message received:', JSON.parse(event.data));
};

ws.onerror = (error) => {
  console.error('WebSocket error:', error);
};

ws.onclose = () => {
  console.log('Disconnected from WebSocket');
};

// Mark message as read
setTimeout(() => {
  ws.send(JSON.stringify({
    type: 'message:read',
    data: {
      messageId: 'msg-456',
      rideId: 'ride-123'
    }
  }));
}, 1000);

// Show typing indicator
setTimeout(() => {
  ws.send(JSON.stringify({
    type: 'message:typing',
    data: {
      rideId: 'ride-123',
      isTyping: true
    }
  }));
}, 2000);

// Stop typing
setTimeout(() => {
  ws.send(JSON.stringify({
    type: 'message:typing',
    data: {
      rideId: 'ride-123',
      isTyping: false
    }
  }));
}, 5000);
```

Run with Node.js:

```bash
node test-websocket.js
```

### 2. WebSocket Test with wscat

Install wscat:

```bash
npm install -g wscat
```

Connect and send messages:

```bash
# Step 1: Connect to WebSocket
wscat -c "ws://localhost:8080/ws/connect?token=YOUR_JWT_TOKEN"

# Step 2: Wait for "Connected" message, then in the wscat prompt (>) send JSON:

# Send a message
> {"type":"message:send","data":{"rideId":"ride-123","content":"Hello!","messageType":"text"}}

# Mark as read
> {"type":"message:read","data":{"messageId":"msg-456","rideId":"ride-123"}}

# Typing indicator
> {"type":"message:typing","data":{"rideId":"ride-123","isTyping":true}}

# Show presence
> {"type":"presence:online","data":{"rideId":"ride-123"}}

# Press Ctrl+C to disconnect
```

**Note for PowerShell users:** The `>` is the wscat prompt, not a shell redirect. Don't include it in your command.

**Example for PowerShell:**
```powershell
# Connect
wscat -c "ws://localhost:8080/ws/connect?token=$TOKEN"

# Once connected, paste the JSON (wscat shows > automatically):
{"type":"message:send","data":{"rideId":"ride-123","content":"Hello!","messageType":"text"}}
```

**Batch multiple commands:**
```powershell
@"
{"type":"message:send","data":{"rideId":"ride-123","content":"Hello!","messageType":"text"}}
{"type":"message:typing","data":{"rideId":"ride-123","isTyping":true}}
{"type":"presence:online","data":{"rideId":"ride-123"}}
"@ | wscat -c "ws://localhost:8080/ws/connect?token=$TOKEN"
```

### 3. WebSocket Test with VS Code REST Client

Create `websocket.http`:

```http
### WebSocket Connection Test
@host = localhost:8080
@token = YOUR_JWT_TOKEN

GET ws://@host/ws/connect?token=@token HTTP/1.1
Connection: Upgrade
Upgrade: websocket
```

**Note:** VS Code REST Client has limited WebSocket support. Use wscat or JavaScript instead.

### 4. Multiple Client Simulation

Create `test-multiple-clients.js`:

```javascript
const token1 = 'DRIVER_JWT_TOKEN';
const token2 = 'RIDER_JWT_TOKEN';
const rideId = 'ride-123';

// Connect first client (driver)
const ws1 = new WebSocket(`ws://localhost:8080/ws/connect?token=${token1}`);
ws1.onmessage = (event) => {
  console.log('[Driver received]:', JSON.parse(event.data));
};

// Connect second client (rider)
const ws2 = new WebSocket(`ws://localhost:8080/ws/connect?token=${token2}`);
ws2.onmessage = (event) => {
  console.log('[Rider received]:', JSON.parse(event.data));
};

// Wait for connections
setTimeout(() => {
  // Rider sends message
  console.log('[Rider] Sending message...');
  ws2.send(JSON.stringify({
    type: 'message:send',
    data: {
      rideId,
      content: 'Hello driver, where are you?',
      messageType: 'text'
    }
  }));

  // Driver types response
  setTimeout(() => {
    console.log('[Driver] Showing typing indicator...');
    ws1.send(JSON.stringify({
      type: 'message:typing',
      data: { rideId, isTyping: true }
    }));

    // Driver sends message
    setTimeout(() => {
      console.log('[Driver] Sending message...');
      ws1.send(JSON.stringify({
        type: 'message:send',
        data: {
          rideId,
          content: 'I am 2 minutes away!',
          messageType: 'text'
        }
      }));
    }, 2000);
  }, 1000);
}, 1000);
```

Run:

```bash
node test-multiple-clients.js
```

## Testing Workflow

### 1. Start the Server

```bash
cd e:\final_go_backend\supr-backend-go
./bin/go-backend
```

Or with Docker:

```bash
docker-compose up
```

### 2. Get JWT Tokens

```bash
# Login as rider
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "rider@example.com",
    "password": "password123"
  }'

# Extract token from response
# Store as RIDER_TOKEN

# Login as driver
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "driver@example.com",
    "password": "password123"
  }'

# Extract token from response
# Store as DRIVER_TOKEN
```

### 3. Create a Ride

```bash
curl -X POST http://localhost:8080/api/v1/rides \
  -H "Authorization: Bearer $RIDER_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "pickupLocation": "37.7749,-122.4194",
    "dropoffLocation": "37.7849,-122.4094",
    "rideType": "economy"
  }'

# Extract rideId from response
# Store as RIDE_ID
```

### 4. Test WebSocket Messages

Open two terminals:

**Terminal 1 (Driver):**
```bash
wscat -c "ws://localhost:8080/ws/connect?token=$DRIVER_TOKEN"
# In prompt:
{"type":"presence:online","data":{"rideId":"$RIDE_ID"}}
{"type":"message:typing","data":{"rideId":"$RIDE_ID","isTyping":true}}
```

**Terminal 2 (Rider):**
```bash
wscat -c "ws://localhost:8080/ws/connect?token=$RIDER_TOKEN"
# In prompt (should see driver's presence and typing):
# Then send message:
{"type":"message:send","data":{"rideId":"$RIDE_ID","content":"Where are you?","messageType":"text"}}
```

Both clients should see all messages in real-time!

### 5. Verify Database Persistence

```bash
# Check if messages are stored
psql postgresql://user:password@localhost:5432/supr_backend -c "
  SELECT id, ride_id, sender_id, content, is_read, created_at 
  FROM ride_messages 
  WHERE ride_id = '$RIDE_ID' 
  ORDER BY created_at DESC;
"
```

## Performance Testing

### Load Test with K6

Create `k6-websocket-test.js`:

```javascript
import ws from 'k6/ws';
import { check } from 'k6';

export const options = {
  vus: 10,
  duration: '30s',
};

const token = __ENV.TOKEN;
const rideId = __ENV.RIDE_ID;

export default function () {
  const url = `ws://localhost:8080/ws/connect?token=${token}`;
  
  const res = ws.connect(url, {}, function (socket) {
    socket.on('open', () => {
      // Send message
      socket.send(JSON.stringify({
        type: 'message:send',
        data: {
          rideId,
          content: `Message from VU ${__VU}`,
          messageType: 'text'
        }
      }));
    });

    socket.on('message', (data) => {
      check(data, {
        'received message': (msg) => msg.length > 0,
      });
    });

    socket.setTimeout(() => {
      socket.close();
    }, 5000);
  });

  check(res, { 'status is 101': (r) => r && r.status === 101 });
}
```

Run:

```bash
k6 run -e TOKEN=$JWT_TOKEN -e RIDE_ID=$RIDE_ID k6-websocket-test.js
```

## Troubleshooting

### 1. Connection Refused

- Ensure server is running on port 8080
- Check firewall rules
- Verify JWT token is valid

### 2. Authentication Error

```json
{
  "type": "error",
  "data": "authentication failed"
}
```

- Check JWT token expiration
- Verify token in query parameter: `?token=YOUR_TOKEN`
- Token must have `user_id` claim

### 3. Message Not Received

- Check if both clients are connected to same WebSocket server
- Verify `rideId` is correct in message
- Check server logs for handler errors

### 4. Database Not Updated

- Verify database connection string is correct
- Check if `ride_messages` table exists
- Ensure migration ran: `000013_create_ride_messages.up.sql`

### 5. Slow Message Delivery

- Check Redis connection (if enabled)
- Monitor network latency
- Verify CPU/memory on server

## Best Practices

1. **Always connect with valid JWT token**
   - Token must have `user_id` and `role` claims
   - Token must not be expired

2. **Send heartbeat/ping messages regularly**
   - Server sends ping every 30 seconds
   - Client should respond with pong
   - Helps detect dead connections

3. **Implement exponential backoff for reconnection**
   - Initial retry: 1 second
   - Max retry: 30 seconds
   - Max attempts: 5

4. **Filter messages by rideId on client side**
   - Server broadcasts to all clients
   - Client should only display messages for current ride

5. **Cache unread count**
   - Use REST API for initial load
   - Update on each new message via WebSocket
   - Re-sync periodically (every 60 seconds)

## Summary

The WebSocket implementation provides real-time messaging with:

- ✅ Event-driven architecture
- ✅ Database persistence
- ✅ Real-time broadcast to all clients in ride
- ✅ Read receipts and typing indicators
- ✅ Presence tracking (online/offline)
- ✅ Cross-server Redis PubSub support
- ✅ Graceful connection handling
- ✅ Message acknowledgment with retry

Start testing now!
