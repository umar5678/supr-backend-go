# SOS Module Development Guide

## Overview

The SOS Module implements emergency and safety features including emergency contact management, panic button functionality, and incident reporting for user safety.

## Key Responsibilities

1. Emergency Contact Management - Manage emergency contacts
2. Panic Button - Quick emergency activation
3. Emergency Alerts - Distribute alerts
4. Incident Reporting - Report safety incidents
5. Safety Features - Location sharing during emergencies

## Data Transfer Objects

### EmergencyContactRequest

```go
type EmergencyContactRequest struct {
    Name          string `json:"name" binding:"required"`
    Phone         string `json:"phone" binding:"required"`
    Relationship  string `json:"relationship"`
    IsPrimary     bool   `json:"is_primary,omitempty"`
}
```

### EmergencyContactResponse

```go
type EmergencyContactResponse struct {
    ID            string    `json:"id"`
    UserID        string    `json:"user_id"`
    Name          string    `json:"name"`
    Phone         string    `json:"phone"`
    Relationship  string    `json:"relationship"`
    IsPrimary     bool      `json:"is_primary"`
    VerifiedAt    *time.Time `json:"verified_at,omitempty"`
    CreatedAt     time.Time `json:"created_at"`
}
```

### TriggerSOSRequest

```go
type TriggerSOSRequest struct {
    EmergencyType   string `json:"emergency_type"` // harassment, accident, medical, other
    Location        Location `json:"location"`
    Description     string `json:"description,omitempty"`
    RideID          string `json:"ride_id,omitempty"`
}
```

### SOSResponse

```go
type SOSResponse struct {
    ID              string    `json:"id"`
    UserID          string    `json:"user_id"`
    EmergencyType   string    `json:"emergency_type"`
    Status          string    `json:"status"` // active, resolved, false_alarm
    Location        Location  `json:"location"`
    AlertsSent      []string  `json:"alerts_sent"`
    ShareWithPolice bool      `json:"share_with_police"`
    CreatedAt       time.Time `json:"created_at"`
    ResolvedAt      *time.Time `json:"resolved_at,omitempty"`
}
```

### IncidentReportRequest

```go
type IncidentReportRequest struct {
    Type            string   `json:"type"` // harassment, accident, property_damage, other
    OtherPartyID    string   `json:"other_party_id,omitempty"`
    Description     string   `json:"description" binding:"required"`
    Location        Location `json:"location"`
    PhotoURLs       []string `json:"photo_urls,omitempty"`
    Witnesses       []string `json:"witnesses,omitempty"`
    ReportToPolice  bool     `json:"report_to_police"`
}
```

## Handler Methods

```
AddEmergencyContact(c *gin.Context)      // POST /sos/contacts
GetEmergencyContacts(c *gin.Context)     // GET /sos/contacts
UpdateEmergencyContact(c *gin.Context)   // PUT /sos/contacts/{id}
DeleteEmergencyContact(c *gin.Context)   // DELETE /sos/contacts/{id}
TriggerSOS(c *gin.Context)               // POST /sos/trigger
GetSOSStatus(c *gin.Context)             // GET /sos/{id}
ResolveSOS(c *gin.Context)               // POST /sos/{id}/resolve
ReportIncident(c *gin.Context)           // POST /sos/report
GetIncidentReport(c *gin.Context)        // GET /sos/report/{id}
ShareLocation(c *gin.Context)            // POST /sos/{id}/share-location
```

## Service Methods

```
AddEmergencyContact(ctx context.Context, userID string, req EmergencyContactRequest) (*EmergencyContactResponse, error)
GetEmergencyContacts(ctx context.Context, userID string) ([]EmergencyContactResponse, error)
UpdateEmergencyContact(ctx context.Context, contactID string, updates map[string]interface{}) error
DeleteEmergencyContact(ctx context.Context, contactID string) error
TriggerSOS(ctx context.Context, userID string, req TriggerSOSRequest) (*SOSResponse, error)
GetSOSStatus(ctx context.Context, sosID string) (*SOSResponse, error)
ResolveSOS(ctx context.Context, sosID string) error
ReportIncident(ctx context.Context, userID string, req IncidentReportRequest) (*IncidentReportResponse, error)
SendEmergencyAlert(ctx context.Context, sosID string) error
ShareLocationWithContacts(ctx context.Context, sosID string) error
NotifyPolice(ctx context.Context, sosID string) error
GetIncidentHistory(ctx context.Context, userID string) ([]IncidentReportResponse, error)
```

## Safety Features

### Emergency Contact Management

```
- Up to 5 emergency contacts per user
- Contacts verified by SMS/call
- Quick-dial capability
- Contact relationship tracking
- Priority ordering
```

### Panic Button

```
One-tap activation:
1. Sends alert to emergency contacts
2. Shares live location
3. Records video/audio (optional)
4. Notifies platform support
5. Option to contact police
```

### Location Sharing

```
During SOS:
- Real-time location shared with contacts
- Updates every 10 seconds
- Stops when SOS resolved
- Contact can see route/status
- Can be cancelled anytime
```

## Emergency Type Handling

```
HARASSMENT:
- Alert contacts immediately
- Share location
- Record incident
- Offer police report

ACCIDENT:
- Alert emergency services
- Share location to nearby users
- Medical assistance check
- Insurance information capture

MEDICAL_EMERGENCY:
- Alert closest hospital
- Share medical info (with consent)
- Emergency contacts notified
- Ambulance dispatch if needed

OTHER:
- General panic alert
- Contact and police notification
- Detailed incident report
- Follow-up support
```

## Incident Reporting

```
Process:
1. User fills incident details
2. Photos/evidence attached
3. Witness contact info collected
4. Report reviewed by admin
5. Police report filed if requested
6. Follow-up communication

Types:
- Harassment/abuse
- Accident
- Property damage
- Safety concerns
- Driver misconduct
```

## Typical Use Cases

### 1. Add Emergency Contact

Request:
```
POST /sos/contacts
{
    "name": "Mom",
    "phone": "+1234567890",
    "relationship": "Mother",
    "is_primary": true
}
```

Flow:
1. Validate phone number
2. Send verification SMS
3. Wait for verification
4. Store contact
5. Return confirmation

### 2. Trigger SOS

Request:
```
POST /sos/trigger
{
    "emergency_type": "harassment",
    "location": {
        "latitude": 40.7128,
        "longitude": -74.0060
    },
    "description": "Driver behaving aggressively"
}
```

Flow:
1. Create SOS record
2. Get user's emergency contacts
3. Send SMS/call alerts to contacts
4. Notify platform support (24/7)
5. Enable location sharing
6. Offer police notification
7. Start recording (if enabled)
8. Return SOS confirmation

### 3. Share Location

Request:
```
POST /sos/{sosID}/share-location
```

Flow:
1. Start real-time location updates
2. Send location link to contacts
3. Update every 10 seconds
4. Contacts receive updates
5. Continue until SOS resolved

### 4. Report Incident

Request:
```
POST /sos/report
{
    "type": "driver_misconduct",
    "description": "Driver was rude and took wrong route",
    "location": {
        "latitude": 40.7128,
        "longitude": -74.0060
    },
    "photo_urls": ["url1", "url2"],
    "report_to_police": false
}
```

Flow:
1. Create incident record
2. Attach photos/evidence
3. Admin review flag if serious
4. Send confirmation to user
5. Notify driver if appropriate
6. Log for monitoring
7. File police report if requested

### 5. Resolve SOS

Request:
```
POST /sos/{sosID}/resolve
```

Flow:
1. Find SOS record
2. Stop location sharing
3. Stop recording
4. Mark as resolved
5. Get follow-up options
6. Notify all contacts
7. Offer follow-up support

## Safety Metrics

```
Track:
- SOS incidents per user/driver
- Types of incidents
- Response times
- Police involvement
- Support effectiveness
```

## Privacy and Consent

```
Location Sharing:
- User must explicitly enable
- Can disable anytime
- Auto-disables when resolved
- Can choose who sees location

Data Handling:
- Police reports encrypted
- Incident history confidential
- Emergency contact info protected
- Contact with user before sharing
```

## Database Schema

### Emergency Contacts Table

```sql
CREATE TABLE emergency_contacts (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    name VARCHAR(255),
    phone VARCHAR(20),
    relationship VARCHAR(100),
    is_primary BOOLEAN,
    is_verified BOOLEAN DEFAULT false,
    verified_at TIMESTAMP,
    created_at TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id),
    INDEX (user_id)
);
```

### SOS Records Table

```sql
CREATE TABLE sos_records (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    emergency_type VARCHAR(50),
    status VARCHAR(50),
    location JSON,
    description TEXT,
    ride_id VARCHAR(36),
    is_recording BOOLEAN,
    recording_url VARCHAR(500),
    alerts_sent JSON,
    share_with_police BOOLEAN,
    created_at TIMESTAMP,
    resolved_at TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id),
    INDEX (user_id, created_at),
    INDEX (status)
);
```

### Incident Reports Table

```sql
CREATE TABLE incident_reports (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    type VARCHAR(50),
    other_party_id VARCHAR(36),
    description TEXT,
    location JSON,
    photo_urls JSON,
    witnesses JSON,
    status VARCHAR(50),
    report_to_police BOOLEAN,
    police_reference VARCHAR(100),
    created_at TIMESTAMP,
    resolved_at TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id)
);
```

---
