# Frontend Integration Guide: Admin Support Chat & SOS Location Tracking

## Overview

This guide explains how to integrate two key features:
1. **Admin Support Chat** - Any user role can connect to admin for chat support
2. **SOS Live Location Tracking** - Users can trigger SOS and stream live location to admin

---

## 1. Admin Support Chat Integration

### 1.1 WebSocket Connection

First, establish a WebSocket connection with auth token:

```javascript
// Connect to WebSocket
const token = localStorage.getItem('authToken');
const ws = new WebSocket(`ws://localhost:8080/ws/connect?token=${token}`);

ws.onopen = () => {
    console.log('WebSocket connected');
};

ws.onmessage = (event) => {
    const message = JSON.parse(event.data);
    handleIncomingMessage(message);
};

ws.onerror = (error) => {
    console.error('WebSocket error:', error);
};
```

### 1.2 Send Chat Message to Admin

Send a chat message from any role (driver, rider, etc.) to admin:

```javascript
// Send admin support chat message
function sendChatToAdmin(content, metadata = {}) {
    const message = {
        type: 'admin_support_chat',
        data: {
            content: content,
            metadata: metadata
        }
    };
    
    ws.send(JSON.stringify(message));
    console.log('Chat sent to admin');
}

// Example Usage
sendChatToAdmin('I need help with my ride', {
    issue_type: 'ride_issue',
    ride_id: 'ride-123'
});
```

### 1.3 Receive Chat Messages from Admin

Admin messages are received via WebSocket with type `chat_message`:

```javascript
function handleIncomingMessage(message) {
    if (message.type === 'chat_message' && message.data.adminSupport) {
        // This is a message from admin in support chat
        const { senderId, senderRole, content, timestamp } = message.data;
        
        console.log(`Admin (${senderRole}): ${content}`);
        updateChatUI(message.data);
    }
}
```

### 1.4 Complete Chat Example

```javascript
class AdminSupportChat {
    constructor(authToken) {
        this.token = authToken;
        this.ws = null;
        this.setupWebSocket();
    }
    
    setupWebSocket() {
        this.ws = new WebSocket(`ws://localhost:8080/ws/connect?token=${this.token}`);
        this.ws.onopen = () => { console.log('Connected'); };
        this.ws.onmessage = (e) => this.onMessage(e);
    }
    
    onMessage(event) {
        const msg = JSON.parse(event.data);
        if (msg.type === 'chat_message' && msg.data.adminSupport) {
            this.displayMessage(msg.data);
        }
    }
    
    sendMessage(content) {
        this.ws.send(JSON.stringify({
            type: 'admin_support_chat',
            data: { content: content }
        }));
    }
    
    displayMessage(data) {
        const chatDiv = document.getElementById('chat-messages');
        const msgElement = document.createElement('div');
        msgElement.innerHTML = `${data.senderRole}: ${data.content}`;
        chatDiv.appendChild(msgElement);
    }
}

// Usage
const chat = new AdminSupportChat(token);
chat.sendMessage('I have a question about my ride');
```

---

## 2. SOS Live Location Tracking

### 2.1 Trigger SOS Alert

First, trigger an SOS using the REST API:

```javascript
async function triggerSOS(latitude, longitude, rideId = null) {
    try {
        const response = await fetch('http://localhost:8080/api/v1/sos/trigger', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${token}`
            },
            body: JSON.stringify({
                latitude: latitude,
                longitude: longitude,
                rideId: rideId
            })
        });
        
        const data = await response.json();
        if (response.ok) {
            console.log('SOS triggered:', data.data);
            return data.data.id; // Return SOS alert ID
        } else {
            console.error('Failed to trigger SOS:', data.message);
        }
    } catch (error) {
        console.error('Error triggering SOS:', error);
    }
}

// Usage
const sosAlertId = await triggerSOS(28.6139, 77.2090, 'ride-123');
```

### 2.2 Stream Live Location to Admin

Once SOS is triggered, continuously update user location:

```javascript
class SOSLocationTracker {
    constructor(authToken, sosAlertId) {
        this.token = authToken;
        this.sosAlertId = sosAlertId;
        this.isTracking = false;
        this.watchtId = null;
    }
    
    // Start continuous location tracking
    startTracking() {
        if (this.isTracking) return;
        
        this.isTracking = true;
        
        // Watch position changes
        this.watchId = navigator.geolocation.watchPosition(
            (position) => this.updateLocation(position),
            (error) => this.handleLocationError(error),
            {
                enableHighAccuracy: true,
                maximumAge: 1000,
                timeout: 5000
            }
        );
        
        console.log('Location tracking started');
    }
    
    // Send location update
    async updateLocation(position) {
        const { latitude, longitude } = position.coords;
        
        try {
            const response = await fetch(
                `http://localhost:8080/api/v1/sos/${this.sosAlertId}/location`,
                {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                        'Authorization': `Bearer ${this.token}`
                    },
                    body: JSON.stringify({
                        latitude: latitude,
                        longitude: longitude
                    })
                }
            );
            
            if (response.ok) {
                console.log(`Location updated: ${latitude}, ${longitude}`);
            } else {
                console.error('Failed to update location');
            }
        } catch (error) {
            console.error('Error updating location:', error);
        }
    }
    
    // Stop tracking when SOS is resolved
    stopTracking() {
        if (this.watchId) {
            navigator.geolocation.clearWatch(this.watchId);
        }
        this.isTracking = false;
        console.log('Location tracking stopped');
    }
    
    handleLocationError(error) {
        console.error('Location error:', error.message);
    }
}

// Usage
const tracker = new SOSLocationTracker(token, sosAlertId);
tracker.startTracking();

// Stop when SOS is resolved
// tracker.stopTracking();
```

### 2.3 Complete SOS Flow Example

```javascript
class SOSManager {
    constructor(authToken) {
        this.token = authToken;
        this.sosAlertId = null;
        this.tracker = null;
    }
    
    // User triggers SOS
    async triggerSOS() {
        try {
            // Get current location
            const position = await this.getCurrentLocation();
            const { latitude, longitude } = position.coords;
            
            // Create SOS alert
            const response = await fetch('http://localhost:8080/api/v1/sos/trigger', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${this.token}`
                },
                body: JSON.stringify({
                    latitude: latitude,
                    longitude: longitude,
                    rideId: this.getCurrentRideId()
                })
            });
            
            const data = await response.json();
            
            if (response.ok) {
                this.sosAlertId = data.data.id;
                console.log('🚨 SOS Alert Triggered!', this.sosAlertId);
                
                // Start location tracking
                this.tracker = new SOSLocationTracker(this.token, this.sosAlertId);
                this.tracker.startTracking();
                
                // Update UI
                this.showSOSStatus(true);
                
                return data.data;
            }
        } catch (error) {
            console.error('Failed to trigger SOS:', error);
        }
    }
    
    // Resolve SOS when safe
    async resolveSOS(notes = '') {
        try {
            const response = await fetch(
                `http://localhost:8080/api/v1/sos/${this.sosAlertId}/resolve`,
                {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                        'Authorization': `Bearer ${this.token}`
                    },
                    body: JSON.stringify({ notes: notes })
                }
            );
            
            if (response.ok) {
                console.log('✅ SOS Resolved');
                
                // Stop tracking
                if (this.tracker) {
                    this.tracker.stopTracking();
                }
                
                this.showSOSStatus(false);
            }
        } catch (error) {
            console.error('Failed to resolve SOS:', error);
        }
    }
    
    getCurrentLocation() {
        return new Promise((resolve, reject) => {
            navigator.geolocation.getCurrentPosition(resolve, reject);
        });
    }
    
    getCurrentRideId() {
        // Get from app state or URL
        return document.querySelector('[data-ride-id]')?.dataset.rideId;
    }
    
    showSOSStatus(isActive) {
        const sosBtn = document.getElementById('sos-button');
        if (isActive) {
            sosBtn.classList.add('active');
            sosBtn.innerHTML = '🚨 SOS ACTIVE - Tap to Resolve';
        } else {
            sosBtn.classList.remove('active');
            sosBtn.innerHTML = '🆘 SOS';
        }
    }
}

// Usage in your app
const sosManager = new SOSManager(token);

// Trigger SOS
document.getElementById('sos-button').addEventListener('click', () => {
    if (!sosManager.sosAlertId) {
        sosManager.triggerSOS();
    } else {
        sosManager.resolveSOS('Safe now');
    }
});
```

---

## 3. WebSocket Message Formats

### 3.1 Admin Support Chat Message (Client → Server)

```json
{
    "type": "admin_support_chat",
    "data": {
        "content": "I need urgent help",
        "metadata": {
            "issue_type": "ride_issue",
            "ride_id": "ride-123",
            "severity": "high"
        }
    }
}
```

### 3.2 Admin Support Chat Message (Server → Admin)

```json
{
    "type": "chat_message",
    "data": {
        "senderId": "user-456",
        "senderRole": "rider",
        "content": "I need urgent help",
        "metadata": {
            "issue_type": "ride_issue",
            "ride_id": "ride-123"
        },
        "timestamp": "2026-03-11T10:30:00Z",
        "adminSupport": true
    }
}
```

### 3.3 SOS Alert Update (Live Location)

```json
{
    "type": "sos_alert",
    "data": {
        "userId": "user-123",
        "location": {
            "latitude": 28.6139,
            "longitude": 77.2090
        },
        "timestamp": "2026-03-11T10:35:00Z",
        "sosActive": true
    }
}
```

---

## 4. REST API Endpoints

### 4.1 Trigger SOS

**POST** `/api/v1/sos/trigger`

```bash
curl -X POST http://localhost:8080/api/v1/sos/trigger \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "latitude": 28.6139,
    "longitude": 77.2090,
    "rideId": "ride-123"
  }'
```

**Response:**
```json
{
  "status": "success",
  "data": {
    "id": "sos-alert-123",
    "userId": "user-456",
    "status": "active",
    "latitude": 28.6139,
    "longitude": 77.2090,
    "triggeredAt": "2026-03-11T10:30:00Z"
  }
}
```

### 4.2 Update SOS Location

**POST** `/api/v1/sos/{id}/location`

```bash
curl -X POST http://localhost:8080/api/v1/sos/sos-alert-123/location \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "latitude": 28.6150,
    "longitude": 77.2100
  }'
```

**Response:**
```json
{
  "status": "success",
  "message": "SOS location updated and broadcast to admin"
}
```

### 4.3 Resolve SOS

**POST** `/api/v1/sos/{id}/resolve`

```bash
curl -X POST http://localhost:8080/api/v1/sos/sos-alert-123/resolve \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "notes": "Safe now, issue resolved"
  }'
```

### 4.4 Get Active SOS

**GET** `/api/v1/sos/active`

```bash
curl -X GET http://localhost:8080/api/v1/sos/active \
  -H "Authorization: Bearer <token>"
```

---

## 5. Error Handling

### 5.1 Common Error Responses

**Already has active SOS:**
```json
{
  "statusCode": 400,
  "status": "error",
  "message": "You already have an active SOS alert"
}
```

**Invalid location:**
```json
{
  "statusCode": 400,
  "status": "error",
  "message": "Latitude must be between -90 and 90"
}
```

**SOS not found:**
```json
{
  "statusCode": 404,
  "status": "error",
  "message": "SOS alert not found"
}
```

### 5.2 Error Handling in Frontend

```javascript
async function safeSOSRequest(endpoint, method = 'POST', body = null) {
    try {
        const response = await fetch(endpoint, {
            method,
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${token}`
            },
            body: body ? JSON.stringify(body) : null
        });
        
        const data = await response.json();
        
        if (!response.ok) {
            // Handle specific errors
            switch (response.status) {
                case 400:
                    console.error('Bad Request:', data.message);
                    showToast(data.message, 'error');
                    break;
                case 401:
                    console.error('Unauthorized - token expired');
                    redirectToLogin();
                    break;
                case 404:
                    console.error('Not Found:', data.message);
                    showToast('SOS alert not found', 'error');
                    break;
                default:
                    console.error('Error:', data.message);
                    showToast('Something went wrong', 'error');
            }
            return null;
        }
        
        return data.data;
    } catch (error) {
        console.error('Network error:', error);
        showToast('Network error', 'error');
        return null;
    }
}
```

---

## 6. UI Components Reference

### 6.1 Admin Support Chat Widget

```html
<div class="admin-support-chat">
    <div class="chat-header">
        <h3>Admin Support</h3>
        <span class="status online">Online</span>
    </div>
    
    <div id="chat-messages" class="chat-messages">
        <!-- Messages appear here -->
    </div>
    
    <div class="chat-input">
        <input 
            type="text" 
            id="message-input" 
            placeholder="Type your message..."
        />
        <button id="send-btn">Send</button>
    </div>
</div>

<style>
.admin-support-chat {
    width: 300px;
    height: 400px;
    border: 1px solid #ddd;
    border-radius: 8px;
    display: flex;
    flex-direction: column;
}

.chat-messages {
    flex: 1;
    overflow-y: auto;
    padding: 10px;
    background: #f5f5f5;
}

.chat-input {
    padding: 10px;
    display: flex;
    gap: 5px;
}

.chat-input input {
    flex: 1;
    padding: 8px;
    border: 1px solid #ddd;
}

.status.online {
    color: green;
}
</style>
```

### 6.2 SOS Button

```html
<button id="sos-button" class="sos-btn">
    🆘 SOS
</button>

<style>
.sos-btn {
    width: 100%;
    padding: 15px;
    font-size: 18px;
    font-weight: bold;
    color: white;
    background-color: #ff6b6b;
    border: none;
    border-radius: 8px;
    cursor: pointer;
    transition: all 0.3s;
}

.sos-btn:hover {
    background-color: #ff5252;
}

.sos-btn.active {
    background-color: #ff1744;
    animation: pulse 1s infinite;
}

@keyframes pulse {
    0%, 100% { opacity: 1; }
    50% { opacity: 0.7; }
}
</style>
```

---

## 7. Testing Checklist

- [ ] WebSocket connection with valid token
- [ ] Send chat message to admin
- [ ] Receive chat message from admin
- [ ] Trigger SOS alert
- [ ] Start continuous location tracking
- [ ] Update location multiple times
- [ ] Verify admin receives all location updates
- [ ] Resolve SOS alert
- [ ] Stop location tracking
- [ ] Handle network errors gracefully
- [ ] Handle expired auth token
- [ ] Test on different devices (mobile, web)

---

## 8. Troubleshooting

### WebSocket not connecting
- Verify token is valid
- Check WebSocket URL is correct
- Ensure backend is running

### Location not updating
- Check `enableHighAccuracy` is enabled
- Verify geolocation permission is granted
- Check device has GPS/location services

### Chat not appearing for admin
- Verify user is connected to WebSocket
- Check message type is `admin_support_chat`
- Ensure admin client is listening for chat_message type

### SOS not triggering
- Verify latitude/longitude are valid
- Check for existing active SOS alert
- Ensure ride ID is correct format (if provided)

---

## 9. Production Considerations

1. **Location Privacy**: Inform users that location is shared during SOS
2. **Battery**: Location tracking drains battery - limit update frequency
3. **Network**: Use exponential backoff for failed updates
4. **Notifications**: Show visual/audio alerts when SOS is triggered
5. **Encryption**: Use WSS (WebSocket Secure) in production
6. **Rate Limiting**: Implement request throttling for location updates
7. **Offline Support**: Queue messages if offline, sync when reconnected

---

## 10. Support & Questions

For issues or questions, contact the backend team.

Last Updated: March 11, 2026
