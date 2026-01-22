# Messaging System - WebSocket vs REST Architecture

## Current Implementation: REST API Only ✅

Your messaging system is currently implemented as **pure REST endpoints**:

```
POST   /api/v1/messages                        - Send message
GET    /api/v1/messages/rides/{rideId}         - Get messages (polling)
GET    /api/v1/messages/rides/{rideId}/unread-count
POST   /api/v1/messages/{messageId}/read       - Mark as read
DELETE /api/v1/messages/{messageId}            - Delete message
```

### Advantages of REST-Only ✅
- ✅ Simple, stateless architecture
- ✅ Easy to test with any HTTP client
- ✅ Works with traditional load balancers
- ✅ Good for low-frequency messages
- ✅ Clients can implement their own polling strategy
- ✅ No connection state to manage

### Disadvantages ❌
- ❌ Polling causes lag (not real-time)
- ❌ Higher latency for message delivery
- ❌ More server load (constant polling requests)
- ❌ Battery drain on mobile clients
- ❌ Not truly "real-time"

---

## Option 1: Keep REST + Add WebSocket for Real-Time (RECOMMENDED)

**Hybrid approach** - REST for CRUD + WebSocket for real-time delivery:

### Architecture
```
┌─────────────────────────────────────┐
│      REST API (CRUD)                │
├─────────────────────────────────────┤
│  POST   /messages                   │ - Create/Send
│  GET    /messages/rides/{id}        │ - List (history)
│  POST   /messages/{id}/read         │ - Mark read
│  DELETE /messages/{id}              │ - Delete
├─────────────────────────────────────┤
│      WebSocket (Real-Time)          │
├─────────────────────────────────────┤
│  ws://localhost:8080/ws/connect     │ - Real-time delivery
│  Events:                            │
│    - message:new                    │ - New message arrived
│    - message:read                   │ - Message read
│    - typing:indicator               │ - User typing
│    - presence:online/offline        │ - User status
└─────────────────────────────────────┘
```

### Flow Example
```
1. Rider sends message via REST
   POST /api/v1/messages
   ↓
2. Server saves to DB
   ↓
3. Server broadcasts via WebSocket
   → Driver receives real-time notification
   ↓
4. Driver fetches message history via REST (optional)
   GET /api/v1/messages/rides/{rideId}
```

### Implementation Status
```
✅ REST endpoints - COMPLETE
⏳ WebSocket integration - READY (infrastructure exists)
⏳ Real-time delivery - NEEDS IMPLEMENTATION
```

---

## Option 2: WebSocket-Only (Advanced)

**Full real-time** approach - all communication via WebSocket:

```
ws://localhost:8080/ws/chat/{rideId}

Commands:
{
  "type": "send_message",
  "content": "Hello!",
  "metadata": {}
}

Events (received):
{
  "type": "message",
  "data": {
    "id": "msg_123",
    "content": "Hello!",
    "sender": "driver",
    "timestamp": "2026-01-22T22:45:00Z"
  }
}
```

### Pros
- ✅ True real-time, bi-directional
- ✅ Lower latency
- ✅ Persistent connection = less overhead
- ✅ Can handle typing indicators, presence, etc.

### Cons
- ❌ More complex to implement
- ❌ Requires connection state management
- ❌ Harder to scale across multiple servers
- ❌ Needs Redis Pub/Sub or similar for multi-server deployments

---

## Current Codebase: WebSocket Infrastructure Exists ✅

Your backend **already has WebSocket infrastructure**:

```
✅ internal/websocket/manager.go     - Connection management
✅ internal/websocket/hub.go         - Message broadcasting
✅ internal/websocket/handlers.go    - Event handlers
✅ internal/websocket/handlers/      - Ride-specific handlers
   - ride_location.go
   - ride_status_update.go
   - driver_availability.go
   - ride_accept.go
   - ride_cancel.go
```

### Existing WebSocket Features
1. **Presence Detection** - User online/offline status
2. **Location Tracking** - Real-time driver location
3. **Ride Status** - Real-time ride state updates
4. **Broadcasting** - Send to multiple users

---

## Recommendation: Hybrid Approach (REST + WebSocket)

### Implementation Plan

#### Phase 1: REST API (✅ DONE)
- ✅ Send message endpoint
- ✅ Get messages history
- ✅ Mark as read
- ✅ Delete message
- ✅ Unread count

#### Phase 2: WebSocket Integration (⏳ TODO)
1. Create message event handler in WebSocket
2. Emit `message:new` event when message is sent
3. Emit `message:read` event when marked as read
4. Broadcast to all connected users in the ride

#### Phase 3: Enhanced Features (⏳ FUTURE)
- Typing indicators
- Presence detection
- Delivery confirmations
- Message reactions

---

## How to Use: REST vs WebSocket

### For Regular Messages (Use REST)
```bash
# Send message
curl -X POST http://localhost:8080/api/v1/messages \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"rideId": "ride_123", "content": "Hello"}'

# Poll for new messages every 2 seconds
curl http://localhost:8080/api/v1/messages/rides/ride_123 \
  -H "Authorization: Bearer $TOKEN"
```

### For Real-Time Updates (Use WebSocket)
```javascript
// Connect to WebSocket
const ws = new WebSocket('ws://localhost:8080/ws/connect?token=' + token);

// Listen for new messages
ws.onmessage = (event) => {
  const msg = JSON.parse(event.data);
  if (msg.type === 'message' && msg.data.rideId === 'ride_123') {
    // New message arrived in real-time!
    console.log('New message:', msg.data.content);
  }
};
```

---

## Testing Strategy

### REST API Testing (Current - Use These)
```
✅ POST /messages              - Send message
✅ GET /messages/rides/{id}   - Get history
✅ POST /messages/{id}/read   - Mark read
```

**See:** `TESTING-MESSAGES.md` for complete REST testing guide

### WebSocket Testing (Future)
```
⏳ Connect to ws://localhost:8080/ws/connect
⏳ Listen for 'message' events
⏳ Send/receive in real-time
```

---

## Migration Path: Add WebSocket Later

### If you want to add WebSocket now:

**Step 1: Create Message Event Handler**
```go
// internal/websocket/handlers/message.go
package handlers

func HandleMessageEvent(manager *websocket.Manager, msg *Message) {
    // Broadcast to all users in ride
    manager.BroadcastToRide(msg.RideID, msg)
}
```

**Step 2: Hook into Message Service**
```go
// Modify internal/modules/messages/service.go
func (s *service) SendMessage(...) {
    // Save to DB
    msg := s.repo.CreateMessage(...)
    
    // Emit WebSocket event (if manager available)
    // wsManager.BroadcastMessageEvent(msg)
}
```

**Step 3: Update Routes**
```go
// Add WebSocket message route
router.GET("/ws/chat/:rideId", handler.HandleChatConnection)
```

---

## Decision Matrix

| Feature | REST Only | REST + WS | WebSocket Only |
|---------|-----------|-----------|----------------|
| Send Message | ✅ | ✅ | ✅ |
| Message History | ✅ | ✅ | ✅ |
| Real-Time Delivery | ❌ | ✅ | ✅ |
| Typing Indicators | ❌ | ✅ (possible) | ✅ |
| Presence Status | ❌ | ✅ | ✅ |
| Complexity | Low | Medium | High |
| Scalability | Easy | Medium | Hard |
| Mobile Friendly | ✅ | ✅ | ❌ (battery) |
| Latency | 0.5-2s | 50-200ms | <50ms |

---

## Current Best Practice: **REST + WebSocket Hybrid**

### Why This Approach?
1. **REST for CRUD** - Simple, stateless, cacheable
2. **WebSocket for Real-Time** - Leverage existing infrastructure
3. **Flexible** - Works in any network (firewalls allow both)
4. **Scalable** - Can load balance REST, fan-out via Redis Pub/Sub
5. **Developer-Friendly** - Easy to test, debug, monitor

### Example Flow with Hybrid Approach
```
Client 1 (Rider)              Server                Client 2 (Driver)
     |                           |                          |
     +--- REST POST /messages ---|                          |
     |    (Send: "Where are you")|                          |
     |                      [Save to DB]                    |
     |                           |                          |
     |                      [Emit WS event]                 |
     |                           +--- WS broadcast -------->|
     |                           |   (Real-time message)   |
     |                           |                    [Update UI]
     |                           |                          |
     |                           |<-- REST POST /read -------|
     |                      [Update is_read]               |
     |                           |                          |
     |                      [Emit WS event]                |
     |<-- WS broadcast ---------|                          |
     |   (Message marked read)   |                          |
```

---

## Summary

**Your messaging system is currently:**
- ✅ **REST-based** - Good for basic communication
- ✅ **Persistent** - Messages saved to database
- ✅ **Secure** - Authentication required
- ❌ **Not real-time** - No live updates

**To make it real-time, you can:**
1. **Keep current REST as-is** (no changes needed)
2. **Add WebSocket handler** for live delivery
3. **Emit events** when messages are sent/read
4. **Clients listen** on WebSocket for real-time updates

**WebSocket infrastructure is already in place:**
- ✅ Connection manager exists
- ✅ Broadcasting mechanism exists
- ✅ Just need message event handler

Would you like me to implement WebSocket real-time delivery?
