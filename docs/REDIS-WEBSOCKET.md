# Redis and WebSocket Integration Guide

## Overview

Redis and WebSocket are core components of the SUPR backend's real-time capabilities. Redis provides caching and pub/sub messaging, while WebSocket enables bidirectional communication for features like live location tracking, instant notifications, and real-time chat.

## Redis Architecture

### Purpose & Usage

Redis serves three critical functions in SUPR:

1. **Session & Cache Layer** - Store frequently accessed data
2. **Pub/Sub Messaging** - Broadcast events across services
3. **Rate Limiting** - Track API usage per user/IP
4. **Background Job Queue** - Queue tasks for async processing

### Redis Configuration

```go
// internal/config/redis.go
type RedisConfig struct {
    Host     string        // localhost
    Port     string        // 6379
    Password string        // "" or password
    DB       int           // 0-15
    MaxConn  int           // 10
    MaxIdle  int           // 5
    TTL      time.Duration // Default: 1 hour
}

// Connection pool setup
rdb := redis.NewClient(&redis.Options{
    Addr:         "localhost:6379",
    MaxRetries:   3,
    PoolSize:     10,
    PoolTimeout:  4 * time.Second,
    ReadTimeout:  3 * time.Second,
    WriteTimeout: 3 * time.Second,
})
```

### Cache Key Namespacing

All cache keys follow a consistent pattern for organization and collision prevention:

```
Pattern: {service}:{entity}:{identifier}:{attribute}

Examples:
  ride:ride_uuid123
  ride:ride_uuid123:driver_location
  user:user_uuid456:profile
  user:user_uuid456:preferences
  driver:driver_uuid789:availability
  driver:driver_uuid789:rating
  session:session_token_xyz
  pricing:surge:city:new_york
  pricing:surge:city:los_angeles:timestamp
  wallet:user_uuid456:balance
  notification:user_uuid456:unread_count
```

### Cache TTL Strategy

```go
const (
    // Short-lived: Real-time data that changes frequently
    TTL_REALTIME = 30 * time.Second      // Location, active status
    TTL_VOLATILE = 1 * time.Minute       // Session temp data
    TTL_SHORT    = 5 * time.Minute       // API responses, counts
    
    // Medium-lived: User data
    TTL_MEDIUM   = 30 * time.Minute      // User profiles, preferences
    TTL_STANDARD = 1 * time.Hour         // Ride details, transactions
    
    // Long-lived: Static/reference data
    TTL_LONG     = 24 * time.Hour        // Vehicle types, pricing rules
    TTL_STATIC   = 7 * 24 * time.Hour    // Rarely changing data
)
```

### Cache Usage Patterns

#### Pattern 1: Cache-Aside (Lazy Loading)

```go
func (s *DriverService) GetDriver(ctx context.Context, driverID string) (*Driver, error) {
    // Try cache first
    cacheKey := fmt.Sprintf("driver:%s:profile", driverID)
    
    cachedDriver, err := s.cache.Get(ctx, cacheKey)
    if err == nil && cachedDriver != nil {
        // Cache hit - return cached data
        return cachedDriver.(*Driver), nil
    }
    
    // Cache miss - fetch from database
    driver, err := s.repo.GetByID(ctx, driverID)
    if err != nil {
        return nil, err
    }
    
    // Store in cache for future requests
    s.cache.Set(ctx, cacheKey, driver, TTL_MEDIUM)
    
    return driver, nil
}
```

#### Pattern 2: Write-Through Cache

```go
func (s *DriverService) UpdateDriverStatus(ctx context.Context, driverID string, status string) error {
    // Update in database first
    if err := s.repo.UpdateStatus(ctx, driverID, status); err != nil {
        return err
    }
    
    // Update cache immediately after
    cacheKey := fmt.Sprintf("driver:%s:status", driverID)
    s.cache.Set(ctx, cacheKey, status, TTL_VOLATILE)
    
    // Invalidate related caches
    s.cache.Delete(ctx, fmt.Sprintf("driver:%s:profile", driverID))
    
    return nil
}
```

#### Pattern 3: Cache Invalidation on Update

```go
func (s *RiderService) UpdatePreferences(ctx context.Context, riderID string, prefs Preferences) error {
    // Update database
    if err := s.repo.UpdatePreferences(ctx, riderID, prefs); err != nil {
        return err
    }
    
    // Delete old cache entry - will be repopulated on next read
    cacheKey := fmt.Sprintf("user:%s:preferences", riderID)
    s.cache.Delete(ctx, cacheKey)
    
    return nil
}
```

### Redis Pub/Sub System

Pub/Sub is used for broadcasting events across the system:

#### Topics (Channels)

```
ride:{ride_id}:updates        # Ride status changes
ride:{ride_id}:location       # Location updates
ride:{ride_id}:chat           # Ride chat messages
user:{user_id}:notifications  # User notifications
driver:{driver_id}:requests   # Driver ride requests
pricing:surge:updates         # Surge pricing changes
admin:activity:log            # Admin activity logging
system:alerts                 # System-wide alerts
```

#### Publishing Events

```go
// In RideService - when ride status changes
func (s *RideService) UpdateRideStatus(ctx context.Context, rideID string, newStatus string) error {
    ride := &Ride{ID: rideID, Status: newStatus}
    
    // Update database
    if err := s.repo.Update(ctx, ride); err != nil {
        return err
    }
    
    // Publish event for subscribers
    event := map[string]interface{}{
        "type":       "ride_status_updated",
        "ride_id":    rideID,
        "status":     newStatus,
        "timestamp":  time.Now(),
    }
    
    eventJSON, _ := json.Marshal(event)
    channel := fmt.Sprintf("ride:%s:updates", rideID)
    
    s.redis.Publish(ctx, channel, string(eventJSON))
    
    return nil
}

// In PricingService - when surge pricing updates
func (s *PricingService) UpdateSurgePricing(ctx context.Context, city string, multiplier float64) error {
    // Update pricing in database/cache
    
    // Publish to all subscribers
    event := map[string]interface{}{
        "city":       city,
        "multiplier": multiplier,
        "timestamp":  time.Now(),
    }
    
    eventJSON, _ := json.Marshal(event)
    s.redis.Publish(ctx, "pricing:surge:updates", string(eventJSON))
    
    return nil
}
```

#### Subscribing to Events

```go
// Listener goroutine in RideService initialization
func (s *RideService) SubscribeToRideUpdates(rideID string) {
    pubsub := s.redis.Subscribe(context.Background(), fmt.Sprintf("ride:%s:updates", rideID))
    defer pubsub.Close()
    
    ch := pubsub.Channel()
    for msg := range ch {
        var event RideUpdateEvent
        json.Unmarshal([]byte(msg.Payload), &event)
        
        // Handle event (update cache, notify websocket clients, etc)
        s.handleRideUpdate(event)
    }
}
```

## WebSocket Architecture

### WebSocket Server Setup

```go
// internal/websocket/server.go
type WSServer struct {
    router    *gin.Engine
    hub       *Hub
    redis     *redis.Client
    logger    *logrus.Logger
}

type Hub struct {
    clients    map[string]*Client  // Map of user_id -> client
    broadcast  chan Message
    register   chan *Client
    unregister chan *Client
}

type Client struct {
    id       string          // user_id
    ride_id  string          // current ride (if applicable)
    conn     *websocket.Conn
    send     chan Message
}

type Message struct {
    Type      string          `json:"type"`
    Payload   json.RawMessage `json:"payload"`
    Timestamp int64           `json:"timestamp"`
}
```

### WebSocket Routes

```go
// Route setup in main routes file
router.GET("/ws/ride/:ride_id", authMiddleware, wsHandler.HandleRideConnection)
router.GET("/ws/driver/:driver_id", authMiddleware, wsHandler.HandleDriverConnection)
router.GET("/ws/notifications", authMiddleware, wsHandler.HandleNotificationConnection)
```

### WebSocket Connection Lifecycle

```
┌─────────────────────────────────────────────────┐
│ 1. CLIENT INITIATES CONNECTION                  │
│    ws://localhost:8080/ws/ride/ride123          │
│    Headers: Authorization: Bearer {token}       │
└─────────────────────────────────────────────────┘
              ↓
┌─────────────────────────────────────────────────┐
│ 2. SERVER VALIDATES                             │
│    - Check JWT token                            │
│    - Verify user can access ride                │
│    - Check permissions                          │
└─────────────────────────────────────────────────┘
              ↓
┌─────────────────────────────────────────────────┐
│ 3. UPGRADE TO WEBSOCKET                         │
│    conn, err := upgrader.Upgrade(w, r, nil)    │
└─────────────────────────────────────────────────┘
              ↓
┌─────────────────────────────────────────────────┐
│ 4. CREATE CLIENT & REGISTER WITH HUB            │
│    client := &Client{...}                       │
│    hub.register <- client                       │
└─────────────────────────────────────────────────┘
              ↓
┌─────────────────────────────────────────────────┐
│ 5. START READ/WRITE GOROUTINES                  │
│    go client.readPump()    (client → server)    │
│    go client.writePump()   (server → client)    │
└─────────────────────────────────────────────────┘
              ↓
┌─────────────────────────────────────────────────┐
│ 6. HANDLE MESSAGES                              │
│    Both directions: send/receive in real-time   │
└─────────────────────────────────────────────────┘
              ↓
┌─────────────────────────────────────────────────┐
│ 7. CLEANUP ON DISCONNECT                        │
│    hub.unregister <- client                     │
│    Close connection                             │
└─────────────────────────────────────────────────┘
```

### Handler Implementation

```go
// internal/websocket/handler.go
func (h *WSHandler) HandleRideConnection(c *gin.Context) {
    rideID := c.Param("ride_id")
    userID := c.GetString("user_id")  // From auth middleware
    
    // Verify user can access this ride
    ride, err := h.rideService.GetRide(c.Request.Context(), rideID)
    if err != nil {
        c.JSON(404, gin.H{"error": "Ride not found"})
        return
    }
    
    if ride.DriverID != userID && ride.RiderID != userID {
        c.JSON(403, gin.H{"error": "Unauthorized"})
        return
    }
    
    // Upgrade connection to WebSocket
    conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
    if err != nil {
        h.logger.WithError(err).Error("WebSocket upgrade failed")
        return
    }
    
    // Create client
    client := &Client{
        id:      userID,
        ride_id: rideID,
        conn:    conn,
        send:    make(chan Message, 256),
    }
    
    // Register with hub
    h.hub.register <- client
    
    // Start reading from client
    go client.readPump(h.hub, h.rideService)
    // Start writing to client
    go client.writePump()
}
```

### Message Types & Handlers

#### Location Update Message
```go
// Client sends location
type LocationUpdate struct {
    Latitude  float64 `json:"latitude"`
    Longitude float64 `json:"longitude"`
    Accuracy  float64 `json:"accuracy"`
    Timestamp int64   `json:"timestamp"`
}

// Server receives and processes
func (h *Hub) HandleLocationUpdate(client *Client, update LocationUpdate) {
    // Save to database
    location := &Location{
        RideID:    client.ride_id,
        UserID:    client.id,
        Latitude:  update.Latitude,
        Longitude: update.Longitude,
        Accuracy:  update.Accuracy,
    }
    h.locationService.SaveLocation(context.Background(), location)
    
    // Broadcast to other clients in ride
    response := Message{
        Type: "location_updated",
        Payload: marshalJSON(LocationUpdateResponse{
            UserID:    client.id,
            Latitude:  update.Latitude,
            Longitude: update.Longitude,
        }),
        Timestamp: time.Now().Unix(),
    }
    
    h.broadcast <- response
}
```

#### Chat Message
```go
// Client sends chat message
type ChatMessage struct {
    Text      string `json:"text"`
    Timestamp int64  `json:"timestamp"`
}

// Server receives and processes
func (h *Hub) HandleChatMessage(client *Client, msg ChatMessage) {
    // Save to database
    dbMsg := &Message{
        RideID:    client.ride_id,
        SenderID:  client.id,
        Text:      msg.Text,
        CreatedAt: time.Now(),
    }
    h.messageService.SaveMessage(context.Background(), dbMsg)
    
    // Broadcast to ride participants
    response := Message{
        Type: "chat_message",
        Payload: marshalJSON(ChatMessageResponse{
            SenderID:  client.id,
            Text:      msg.Text,
            Timestamp: msg.Timestamp,
        }),
        Timestamp: time.Now().Unix(),
    }
    
    h.broadcast <- response
}
```

#### Status Update
```go
// Server initiates status change
type StatusUpdate struct {
    Status    string `json:"status"`
    Reason    string `json:"reason,omitempty"`
    Timestamp int64  `json:"timestamp"`
}

// Broadcast to all ride participants
func (h *Hub) BroadcastStatusChange(rideID string, status string) {
    response := Message{
        Type: "status_changed",
        Payload: marshalJSON(StatusUpdate{
            Status:    status,
            Timestamp: time.Now().Unix(),
        }),
        Timestamp: time.Now().Unix(),
    }
    
    // Send to all clients in this ride
    for _, client := range h.clients {
        if client.ride_id == rideID {
            select {
            case client.send <- response:
            default:
                // Client's send channel is full
                go h.removeClient(client)
            }
        }
    }
}
```

### Client Read/Write Pumps

```go
// readPump pumps messages from the WebSocket connection to the hub
func (c *Client) readPump(hub *Hub, service RideService) {
    defer func() {
        hub.unregister <- c
        c.conn.Close()
    }()
    
    c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
    c.conn.SetReadLimit(512 * 1024) // Max message size: 512KB
    
    // Configure pong handler (heartbeat)
    c.conn.SetPongHandler(func(string) error {
        c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
        return nil
    })
    
    for {
        var msg Message
        err := c.conn.ReadJSON(&msg)
        if err != nil {
            if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
                log.Printf("WebSocket error: %v", err)
            }
            return
        }
        
        // Route message to appropriate handler
        switch msg.Type {
        case "location_update":
            var update LocationUpdate
            json.Unmarshal(msg.Payload, &update)
            hub.HandleLocationUpdate(c, update)
            
        case "chat_message":
            var chatMsg ChatMessage
            json.Unmarshal(msg.Payload, &chatMsg)
            hub.HandleChatMessage(c, chatMsg)
            
        default:
            log.Printf("Unknown message type: %s", msg.Type)
        }
    }
}

// writePump pumps messages from the hub to the WebSocket connection
func (c *Client) writePump() {
    ticker := time.NewTicker(54 * time.Second)
    defer func() {
        ticker.Stop()
        c.conn.Close()
    }()
    
    for {
        select {
        case message, ok := <-c.send:
            c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
            if !ok {
                // Hub closed the channel
                c.conn.WriteMessage(websocket.CloseMessage, []byte{})
                return
            }
            
            if err := c.conn.WriteJSON(message); err != nil {
                return
            }
            
        case <-ticker.C:
            // Send ping to keep connection alive
            c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
            if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
                return
            }
        }
    }
}
```

## Redis + WebSocket Integration

### Broadcasting via Pub/Sub to WebSocket

When a service event occurs (e.g., another driver's location update), it's published to Redis, and all listening services forward it to connected WebSocket clients:

```go
// Service publishes to Redis
func (s *TrackingService) PublishLocationUpdate(ctx context.Context, rideID, userID string, lat, lon float64) {
    event := LocationUpdateEvent{
        RideID:    rideID,
        UserID:    userID,
        Latitude:  lat,
        Longitude: lon,
        Timestamp: time.Now(),
    }
    
    data, _ := json.Marshal(event)
    s.redis.Publish(ctx, fmt.Sprintf("ride:%s:location", rideID), string(data))
}

// WebSocket server subscribes to Redis
func (ws *WSServer) subscribeToRideUpdates(rideID string) {
    pubsub := ws.redis.Subscribe(context.Background(), fmt.Sprintf("ride:%s:location", rideID))
    defer pubsub.Close()
    
    ch := pubsub.Channel()
    for msg := range ch {
        var event LocationUpdateEvent
        json.Unmarshal([]byte(msg.Payload), &event)
        
        // Broadcast to WebSocket clients
        response := Message{
            Type:      "location_updated",
            Payload:   json.RawMessage(msg.Payload),
            Timestamp: time.Now().Unix(),
        }
        
        ws.hub.BroadcastToRide(rideID, response)
    }
}
```

### Session Management with Redis

```go
// Store session in Redis
func (s *SessionService) CreateSession(userID string, token string) error {
    sessionKey := fmt.Sprintf("session:%s", token)
    sessionData := SessionData{
        UserID:    userID,
        CreatedAt: time.Now(),
        ExpiresAt: time.Now().Add(24 * time.Hour),
    }
    
    data, _ := json.Marshal(sessionData)
    return s.redis.Set(context.Background(), sessionKey, string(data), 24*time.Hour).Err()
}

// Retrieve session
func (s *SessionService) GetSession(token string) (*SessionData, error) {
    sessionKey := fmt.Sprintf("session:%s", token)
    data, err := s.redis.Get(context.Background(), sessionKey).Result()
    if err != nil {
        return nil, err
    }
    
    var session SessionData
    json.Unmarshal([]byte(data), &session)
    return &session, nil
}
```

## Performance Considerations

### Redis Optimization
- Use connection pooling (10-30 connections recommended)
- Set appropriate read/write timeouts (3-5 seconds)
- Use pipelining for batch operations
- Monitor Redis memory usage and eviction policy

### WebSocket Optimization
- Set message size limits (512KB recommended)
- Implement heartbeat/ping-pong (54-second interval)
- Use read/write deadlines (60 seconds recommended)
- Buffer channels appropriately (256 message buffer)
- Clean up disconnected clients immediately

### Rate Limiting
```go
// Example: Limit location updates to 1 per second per driver
func (s *TrackingService) UpdateLocation(ctx context.Context, driverID string, lat, lon float64) error {
    key := fmt.Sprintf("rate_limit:location:%s", driverID)
    
    // Check if already updated recently
    count, _ := s.redis.Incr(ctx, key).Result()
    if count == 1 {
        s.redis.Expire(ctx, key, 1*time.Second)
    }
    
    if count > 5 { // Allow max 5 updates per second
        return errors.New("rate limit exceeded")
    }
    
    // Process update
    return s.repo.SaveLocation(ctx, driverID, lat, lon)
}
```

## Monitoring & Debugging

### Redis Monitoring
```bash
# Monitor all Redis commands in real-time
redis-cli MONITOR

# Check memory usage
redis-cli INFO memory

# List all keys (careful in production)
redis-cli KEYS "*"

# Check specific key
redis-cli GET "ride:ride123"
```

### WebSocket Debugging
```go
// Log all WebSocket events
func (c *Client) readPump(hub *Hub) {
    for {
        var msg Message
        err := c.conn.ReadJSON(&msg)
        if err != nil {
            log.Printf("WS Error from client %s: %v", c.id, err)
            return
        }
        
        log.Printf("WS Message from %s: type=%s, payload_len=%d", c.id, msg.Type, len(msg.Payload))
    }
}
```

---

**Document Version:** 1.0
**Last Updated:** February 22, 2026
**Status:** Production Configuration
