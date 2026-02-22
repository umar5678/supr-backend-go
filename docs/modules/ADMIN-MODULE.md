# Admin Module Development Guide

## Overview

The Admin Module provides comprehensive administrative functionality for system management. It allows administrators to manage users, approve service providers, suspend accounts, and view system statistics.

## Module Structure

```
admin/
├── handler.go         # HTTP request handlers
├── service.go         # Business logic
├── repository.go      # Database operations
├── routes.go          # Route definitions
└── dto/
    ├── requests.go    # Request payloads
    └── responses.go   # Response structures
```

## Key Responsibilities

1. User Management - List, filter, and manage user accounts
2. Service Provider Approval - Process and approve new provider registrations
3. User Suspension - Suspend problematic accounts with reasons
4. Status Management - Update user and provider statuses
5. Dashboard Statistics - Generate admin dashboard metrics

## Architecture

### Handler Layer (handler.go)

Manages HTTP endpoints and request validation.

Key methods:

```
ListUsers(c *gin.Context)                    // GET /admin/users
ApproveServiceProvider(c *gin.Context)       // POST /admin/service-providers/{id}/approve
SuspendUser(c *gin.Context)                  // POST /admin/users/{id}/suspend
UpdateUserStatus(c *gin.Context)             // PUT /admin/users/{id}/status
GetDashboardStats(c *gin.Context)            // GET /admin/dashboard/stats
DeleteUser(c *gin.Context)                   // DELETE /admin/users/{id}
```

Request flow:
1. Extract parameters from URL/Query/Body
2. Validate request data
3. Call service method
4. Return response via response utility

### Service Layer (service.go)

Contains business logic and orchestration.

Key interface methods:

```
ListUsers(ctx context.Context, role, status, page, limit string) (map[string]interface{}, error)
ApproveServiceProvider(ctx context.Context, providerID string) error
SuspendUser(ctx context.Context, userID, reason string) error
UpdateUserStatus(ctx context.Context, userID string, status models.UserStatus) error
GetDashboardStats(ctx context.Context) (map[string]interface{}, error)
DeleteUser(ctx context.Context, userID string) error
```

Logic flow:
1. Validate inputs and parameters
2. Call repository methods
3. Apply business rules
4. Return aggregated results
5. Log significant actions

### Repository Layer (repository.go)

Handles all database operations.

Key interface methods:

```
FindUserByID(ctx context.Context, id string) (*models.User, error)
ListUsers(ctx context.Context, filters map[string]interface{}, page, limit int) ([]*models.User, int64, error)
UpdateUserStatus(ctx context.Context, userID string, status models.UserStatus) error
DeleteUser(ctx context.Context, userID string) error
GetDashboardStats(ctx context.Context) (map[string]interface{}, error)
```

Database operations:
- Use context for query cancellation
- Implement pagination for large datasets
- Apply filters dynamically
- Use transactions for multi-table operations

## Data Transfer Objects

### ListUsersResponse

```go
type ListUsersResponse struct {
    Users []*models.User `json:"users"`
    Total int64          `json:"total"`
    Page  int            `json:"page"`
    Limit int            `json:"limit"`
}
```

### SuspendUserRequest

```go
type SuspendUserRequest struct {
    Reason string `json:"reason" binding:"required"`
}
```

### DashboardStatsResponse

```go
type DashboardStatsResponse struct {
    TotalUsers      int64                    `json:"total_users"`
    UsersByRole     []map[string]interface{} `json:"users_by_role"`
    UsersByStatus   []map[string]interface{} `json:"users_by_status"`
    PendingApprovals int                     `json:"pending_approvals"`
}
```

## Typical Use Cases

### 1. List All Users with Filtering

Request:
```
GET /admin/users?role=driver&status=active&page=1&limit=20
```

Flow:
1. Extract filters from query parameters
2. Service validates pagination parameters
3. Repository queries database with filters
4. Return paginated results with total count

### 2. Approve Service Provider

Request:
```
POST /admin/service-providers/{providerID}/approve
```

Flow:
1. Extract provider ID from URL
2. Service finds service provider profile
3. Service updates user status to active
4. Service updates provider status to active
5. Log approval action
6. Return success response

### 3. Suspend User Account

Request:
```
POST /admin/users/{userID}/suspend
{
    "reason": "Suspicious activity detected"
}
```

Flow:
1. Extract user ID from URL
2. Extract reason from request body
3. Service finds user to verify existence
4. Service updates user status to suspended
5. Log suspension with reason
6. Return success response

### 4. Get Dashboard Statistics

Request:
```
GET /admin/dashboard/stats
```

Flow:
1. Repository aggregates user counts by role
2. Repository aggregates user counts by status
3. Repository calculates total users
4. Service counts pending approvals
5. Return aggregated statistics

## Error Handling

Common error scenarios:

1. Invalid pagination parameters
   - Validation converts to valid defaults
   - Response: 200 OK with defaults applied

2. User not found
   - Repository returns NotFoundError
   - Response: 404 Not Found

3. Database connection error
   - Repository propagates error
   - Response: 500 Internal Server Error

4. Authorization failure
   - Middleware catches
   - Response: 401 Unauthorized

## Security Considerations

1. Authentication - All endpoints require BearerAuth
2. Authorization - Verify admin role in middleware
3. Input Validation - Validate all query/body parameters
4. SQL Injection - Use parameterized queries (GORM)
5. Data Exposure - Only return necessary fields in response

## Testing Strategy

### Unit Tests (Service Layer)

```go
Test_ListUsers_WithFilters()
Test_ApproveServiceProvider_Success()
Test_SuspendUser_InvalidID()
Test_GetDashboardStats()
```

### Integration Tests (Repository Layer)

```go
Test_Repository_ListUsers_WithDatabase()
Test_Repository_UpdateUserStatus()
Test_Repository_DeleteUser()
```

### End-to-End Tests (Handler Layer)

```go
Test_Handler_ListUsers_ValidRequest()
Test_Handler_ApproveServiceProvider_Success()
Test_Handler_SuspendUser_Unauthorized()
```

## Database Schema

### Users Table

```sql
CREATE TABLE users (
    id VARCHAR(36) PRIMARY KEY,
    email VARCHAR(255),
    phone VARCHAR(20),
    role VARCHAR(50),
    status VARCHAR(50),
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);
```

## Integration Points

1. Service Providers Module - For provider profile access
2. Notification Module - For status change notifications
3. Audit Module - For logging admin actions

## Configuration

Admin module configuration (typically in environment or config file):

```yaml
admin:
  pagination:
    default_limit: 20
    max_limit: 100
  cache:
    ttl: 300s
```

## Future Enhancements

1. Advanced search and filtering
2. Bulk operations support
3. Audit trail with filters
4. Custom dashboards
5. Report generation
6. User activity analytics
7. Automated suspension rules

## Common Pitfalls

1. Not validating pagination parameters - Can cause poor performance
2. Missing context propagation - Can cause hanging requests
3. Insufficient logging - Makes debugging difficult
4. Race conditions in status updates - Use transactions
5. Not handling concurrent requests - Implement proper locking

## Performance Optimization

1. Use database indexes on frequently filtered columns (role, status)
2. Implement pagination to avoid large result sets
3. Cache dashboard statistics with TTL
4. Use batch operations for bulk updates
5. Implement query optimization in repository

## Related Documentation

- See MODULES-OVERVIEW.md for module architecture overview
- See internal/models documentation for data models
- See internal/utils/response for error handling patterns
