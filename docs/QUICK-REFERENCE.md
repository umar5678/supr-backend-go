# Module Development Quick Reference

Quick lookup guide for common tasks and questions.

## Module at a Glance

| Module | Type | Purpose | Key Feature |
|--------|------|---------|------------|
| Auth | Core | User authentication | JWT tokens |
| Riders | User | Rider profiles | Saved addresses |
| Drivers | User | Driver management | Document verification |
| Rides | Core | Ride management | Real-time tracking |
| Pricing | Core | Fare calculation | Surge pricing |
| Wallet | Core | Payment management | Balance tracking |
| Tracking | Core | Location tracking | WebSocket updates |
| Admin | Admin | System administration | User management |
| Laundry | Feature | Laundry orders | Weight-based pricing |
| Messages | Support | Notifications | Delivery tracking |
| Ratings | Support | Reviews | Aggregation |
| Promotions | Support | Discounts | Code management |
| Vehicles | Feature | Vehicle management | Registration |
| ServiceProviders | User | Provider profiles | Approval workflow |
| HomeServices | Feature | Home services | Provider matching |
| SOS | Safety | Emergency features | Panic button |
| Fraud | Security | Fraud detection | Risk scoring |
| RidePin | Security | PIN verification | OTP-based |
| Profile | User | User profiles | Preferences |
| Batching | Utility | Batch processing | Job scheduling |

## Quick Checklist: Adding New Feature

1. Choose or create module
2. Define DTOs (request/response)
3. Update repository interface
4. Implement repository methods
5. Update service interface
6. Implement service methods
7. Create handler endpoints
8. Add Swagger documentation
9. Write comprehensive tests
10. Update module documentation
11. Get code review
12. Deploy and monitor

## Code Templates

### Create Handler

```go
package modulename

import (
    "github.com/gin-gonic/gin"
    "github.com/umar5678/go-backend/internal/utils/response"
)

type Handler struct {
    service Service
}

func NewHandler(service Service) *Handler {
    return &Handler{service: service}
}

func (h *Handler) MethodName(c *gin.Context) {
    userID, _ := c.Get("userID")
    
    var req RequestDTO
    if err := c.ShouldBindJSON(&req); err != nil {
        c.Error(response.BadRequest("Invalid request"))
        return
    }
    
    result, err := h.service.MethodName(c.Request.Context(), userID.(string))
    if err != nil {
        c.Error(err)
        return
    }
    
    response.Success(c, result, "Success message")
}
```

### Create Service

```go
package modulename

import (
    "context"
    "github.com/umar5678/go-backend/internal/models"
    "github.com/umar5678/go-backend/internal/utils/response"
)

type Service interface {
    MethodName(ctx context.Context, id string) (*ResponseDTO, error)
}

type service struct {
    repo Repository
}

func NewService(repo Repository) Service {
    return &service{repo: repo}
}

func (s *service) MethodName(ctx context.Context, id string) (*ResponseDTO, error) {
    data, err := s.repo.FindByID(ctx, id)
    if err != nil {
        return nil, response.NotFoundError("Resource")
    }
    
    return &ResponseDTO{Data: data}, nil
}
```

### Create Repository

```go
package modulename

import (
    "context"
    "github.com/umar5678/go-backend/internal/models"
    "gorm.io/gorm"
)

type Repository interface {
    FindByID(ctx context.Context, id string) (*models.Model, error)
    Create(ctx context.Context, m *models.Model) error
    Update(ctx context.Context, id string, m *models.Model) error
}

type repository struct {
    db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
    return &repository{db: db}
}

func (r *repository) FindByID(ctx context.Context, id string) (*models.Model, error) {
    var m models.Model
    err := r.db.WithContext(ctx).Where("id = ?", id).First(&m).Error
    return &m, err
}
```

### Swagger Documentation

```go
// MethodName godoc
// @Summary Brief description
// @Description Longer description if needed
// @Tags module_name
// @Accept json
// @Produce json
// @Param id path string true "Resource ID"
// @Param request body RequestDTO true "Request body"
// @Success 200 {object} response.Response{data=ResponseDTO}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /path/{id} [post]
// @Security BearerAuth
func (h *Handler) MethodName(c *gin.Context) {
    // implementation
}
```

## Error Responses Quick Reference

```
response.BadRequest("message")
response.NotFoundError("resource_name")
response.InternalServerError("message", err)
response.UnauthorizedError("message")
response.ForbiddenError("message")
response.ConflictError("message")
```

## Logging Examples

```go
logger.Info("operation completed", "userID", userID, "amount", amount)
logger.Error("operation failed", "error", err)
logger.Debug("debug info", "details", details)
```

## Pagination Pattern

```go
page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

if page < 1 {
    page = 1
}
if limit < 1 || limit > maxLimit {
    limit = defaultLimit
}

offset := (page - 1) * limit
// Use offset in query
```

## Context Usage

```go
// Always propagate context in database operations
func (r *repository) Operation(ctx context.Context) error {
    return r.db.WithContext(ctx)./* operation */
}

// Always pass context to service
result, err := s.service.Method(c.Request.Context(), ...)
```

## Transaction Pattern

```go
func (r *repository) MultiStepOperation(ctx context.Context) error {
    tx := r.db.WithContext(ctx).BeginTx(ctx, nil)
    
    if err := tx.Model(&model1).Update("field", value).Error; err != nil {
        tx.Rollback()
        return err
    }
    
    if err := tx.Model(&model2).Update("field", value).Error; err != nil {
        tx.Rollback()
        return err
    }
    
    return tx.Commit().Error
}
```

## Database Query Patterns

```go
// Find by ID
var model Model
r.db.WithContext(ctx).First(&model, "id = ?", id)

// Find with multiple conditions
r.db.WithContext(ctx).Where("status = ? AND user_id = ?", status, userID).Find(&models)

// Paginated query
offset := (page - 1) * limit
r.db.WithContext(ctx).Offset(offset).Limit(limit).Find(&models)

// Count
var total int64
r.db.WithContext(ctx).Model(&Model{}).Count(&total)

// Update
r.db.WithContext(ctx).Model(&Model{}).Where("id = ?", id).Updates(updates)

// Delete
r.db.WithContext(ctx).Delete(&Model{}, "id = ?", id)
```

## Testing Examples

### Unit Test

```go
func Test_Service_MethodName_Success(t *testing.T) {
    // Arrange
    mockRepo := new(MockRepository)
    mockRepo.On("FindByID", ctx, id).Return(&data, nil)
    service := NewService(mockRepo)
    
    // Act
    result, err := service.MethodName(ctx, id)
    
    // Assert
    assert.NoError(t, err)
    assert.NotNil(t, result)
    mockRepo.AssertExpectations(t)
}
```

### Integration Test

```go
func Test_Repository_Create_Success(t *testing.T) {
    // Setup database
    db := setupTestDB()
    repo := NewRepository(db)
    
    // Test
    model := &Model{ID: "123", Name: "Test"}
    err := repo.Create(context.Background(), model)
    
    // Verify
    assert.NoError(t, err)
    var found Model
    db.First(&found, "id = ?", "123")
    assert.Equal(t, "Test", found.Name)
}
```

## Common Mistakes to Avoid

1. Not propagating context
2. Not using transactions for multi-step ops
3. Hardcoding values instead of using config
4. Missing input validation
5. Not handling errors properly
6. Insufficient logging
7. No pagination on large queries
8. Not testing error cases
9. Using plain text passwords
10. Missing database indexes

## Performance Checklist

- [ ] Indexes on frequently queried columns
- [ ] Pagination for large datasets
- [ ] Caching implemented for expensive operations
- [ ] Connection pooling configured
- [ ] Query optimization applied
- [ ] N+1 query problem avoided
- [ ] Batch operations used where applicable
- [ ] Response times monitored
- [ ] Database load tested

## Security Checklist

- [ ] Input validation implemented
- [ ] SQL injection prevented (using GORM)
- [ ] Authentication required for protected endpoints
- [ ] Authorization checked
- [ ] Sensitive data encrypted
- [ ] No hardcoded secrets
- [ ] Rate limiting applied
- [ ] Audit logging for critical ops
- [ ] Error messages don't leak info
- [ ] CORS configured properly

## Deployment Checklist

- [ ] All tests passing
- [ ] Code review approved
- [ ] Documentation updated
- [ ] Database migrations created
- [ ] Configuration validated
- [ ] Logging configured
- [ ] Error handling verified
- [ ] Security checks done
- [ ] Performance benchmarked
- [ ] Rollback plan prepared

## File Locations

```
Module Implementation:
- Source: internal/modules/{moduleName}/
- Models: internal/models/
- Middleware: internal/middleware/
- Utils: internal/utils/

Documentation:
- Overview: docs/MODULES-OVERVIEW.md
- Index: docs/MODULE-INDEX.md
- Module Guides: docs/modules/{MODULE-NAME}.md
- Database: docs/database/

Tests:
- Unit Tests: internal/modules/{moduleName}/tests/
- Integration Tests: tests/integration/
```

## Database Migration

Create migration file:
```bash
migrate create -ext sql -dir migrations -seq add_column_name
```

Migration content:
```sql
-- Up
ALTER TABLE table_name ADD COLUMN new_column TYPE;

-- Down
ALTER TABLE table_name DROP COLUMN new_column;
```

## Routes Registration Pattern

```go
func RegisterRoutes(router *gin.Engine, handler *Handler) {
    group := router.Group("/path")
    {
        group.GET("", handler.GetAll)
        group.POST("", handler.Create)
        group.GET("/:id", handler.GetByID)
        group.PUT("/:id", handler.Update)
        group.DELETE("/:id", handler.Delete)
    }
}
```

## Module Initialization

```go
// In main.go
authRepo := auth.NewRepository(db)
authService := auth.NewService(authRepo)
authHandler := auth.NewHandler(authService)
auth.RegisterRoutes(router, authHandler)
```

## Critical Data to Never Log

- Passwords
- API keys and tokens
- Credit card numbers
- Bank account numbers
- SSN/ID numbers
- Personal health information
- Private messages
- Location (in some cases)

## Response Format

```json
{
    "success": true,
    "data": { /* actual data */ },
    "message": "Operation successful",
    "code": 200,
    "timestamp": "2024-02-20T10:30:00Z"
}
```

Error response:
```json
{
    "success": false,
    "data": null,
    "message": "Error message",
    "code": 400,
    "timestamp": "2024-02-20T10:30:00Z"
}
```

## Configuration Pattern

```yaml
server:
  port: 8080
  timeout: 30s
database:
  host: localhost
  port: 5432
  name: supr_backend
auth:
  jwt_secret: "your-secret"
  token_expiry: "15m"
```

## WebSocket Pattern (if applicable)

```go
func (h *Handler) WebSocketHandler(c *gin.Context) {
    conn, err := upgrader.Upgrade(c.ResponseWriter, c.Request, nil)
    if err != nil {
        return
    }
    defer conn.Close()
    
    for {
        var msg Message
        if err := conn.ReadJSON(&msg); err != nil {
            return
        }
        
        result := h.service.ProcessMessage(msg)
        conn.WriteJSON(result)
    }
}
```

## Useful Commands

```bash
# Run tests
go test ./...

# Run specific test
go test -run TestName ./...

# Coverage
go test -cover ./...

# Build
go build -o bin/app

# Run
./bin/app

# Format code
go fmt ./...

# Lint
golangci-lint run ./...

# Benchmark
go test -bench=. -benchmem ./...
```

## Documentation Links

- Module Overview: See MODULES-OVERVIEW.md
- Module Index: See MODULE-INDEX.md
- Specific Module: See docs/modules/{MODULE-NAME}.md
- Database Schema: See specific module docs

## Getting Help

1. Check the specific module documentation
2. Review code examples in the module docs
3. Look at existing implementations
4. Check the common pitfalls section
5. Ask for code review
6. Check the testing section for patterns

## Useful References

- Go Best Practices: https://golang.org/doc/effective_go
- Gin Documentation: https://github.com/gin-gonic/gin
- GORM Documentation: https://gorm.io/docs/
- RESTful API Design: https://restfulapi.net/
- Security: https://owasp.org/Top10/

