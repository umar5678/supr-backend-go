# Module Documentation Index

This document provides an organized index of all module documentation and guides.

## Quick Navigation

### Core Modules (Essential for Platform Operations)

1. [Auth Module](./modules/AUTH-MODULE.md)
   - User authentication and authorization
   - JWT token management
   - Phone-based signup/login
   - OTP verification
   - Password management

2. [Rides Module](./modules/RIDES-MODULE.md)
   - Ride request creation and management
   - Driver assignment and matching
   - Real-time location tracking
   - Ride status management
   - WebSocket communication

3. [Pricing Module](./modules/PRICING-MODULE.md)
   - Fare estimation and calculation
   - Surge pricing algorithms
   - Distance and time-based rates
   - Promotion code handling
   - Pricing rule management

4. [Wallet Module](./modules/WALLET-MODULE.md)
   - Wallet management and balance tracking
   - Payment processing
   - Transaction history
   - Fund management (add/withdraw)
   - Settlement and reconciliation

### User Management Modules

5. [Drivers Module](./modules/DRIVERS-MODULE.md)
   - Driver profile management
   - Document verification
   - Availability and status tracking
   - Earnings and payment tracking
   - Driver restrictions and verification

6. [Riders Module](./modules/RIDERS-MODULE.md)
   - Rider profile management
   - Saved addresses
   - Ride preferences
   - Emergency contacts
   - Rider history

7. [Service Providers Module](./modules/SERVICEPROVIDERS-MODULE.md)
   - Service provider profiles
   - Service listing and management
   - Provider approval workflow
   - Earnings tracking
   - Performance metrics

### Administrative Modules

8. [Admin Module](./modules/ADMIN-MODULE.md)
   - User listing and management
   - Service provider approval
   - User suspension and status management
   - Dashboard statistics
   - Administrative operations

### Feature-Specific Modules

9. [Laundry Module](./modules/LAUNDRY-MODULE.md)
   - Laundry order management
   - Weight-based pricing
   - Service provider assignment
   - Tip management
   - Order tracking

10. [Home Services Module](./modules/HOMESERVICES-MODULE.md)
    - Home service requests
    - Service provider matching
    - Service completion tracking
    - Feedback and ratings

### Support and Utility Modules

11. [Ratings Module](./modules/RATINGS-MODULE.md)
    - Rating submission and storage
    - Review management
    - Rating aggregation
    - Fraud detection in ratings

12. [Messages Module](./modules/MESSAGES-MODULE.md)
    - Direct messaging
    - System notifications
    - Notification delivery tracking
    - Message history

13. [Promotions Module](./modules/PROMOTIONS-MODULE.md)
    - Promotional code management
    - Discount application
    - Campaign management
    - Usage tracking and redemption

14. [Tracking Module](./modules/TRACKING-MODULE.md)
    - Real-time location tracking
    - Location history storage
    - Distance and duration calculations
    - ETA estimation
    - Geofencing

15. [Vehicles Module](./modules/VEHICLES-MODULE.md)
    - Vehicle registration and management
    - Driver-vehicle association
    - Document verification
    - Vehicle type management
    - Availability tracking

16. [Ride PIN Module](./modules/RIDEPIN-MODULE.md)
    - PIN generation and validation
    - OTP-based verification
    - Ride verification
    - Security features

17. [SOS Module](./modules/SOS-MODULE.md)
    - Emergency contact management
    - Panic button/SOS triggering
    - Emergency alert distribution
    - Incident reporting
    - Safety features

18. [Fraud Module](./modules/FRAUD-MODULE.md)
    - Fraud detection algorithms
    - Suspicious activity monitoring
    - Risk scoring
    - Payment fraud detection
    - Account security

19. [Batching Module](./modules/BATCHING-MODULE.md)
    - Batch job management
    - Bulk operations
    - Scheduled execution
    - Progress tracking

20. [Profile Module](./modules/PROFILE-MODULE.md)
    - User profile management
    - Avatar and media management
    - Preference management
    - Privacy settings

## Documentation Overview

### Main Documentation Files

- [MODULES-OVERVIEW.md](./MODULES-OVERVIEW.md) - High-level overview of all modules
- [MODULE-INDEX.md](./MODULE-INDEX.md) - This file - Index and quick reference

### Individual Module Guides

Each module has a dedicated development guide located in the `modules/` directory:

```
docs/modules/
├── ADMIN-MODULE.md
├── AUTH-MODULE.md
├── BATCHING-MODULE.md
├── DRIVERS-MODULE.md
├── FRAUD-MODULE.md
├── HOMESERVICES-MODULE.md
├── LAUNDRY-MODULE.md
├── MESSAGES-MODULE.md
├── PRICING-MODULE.md
├── PROFILE-MODULE.md
├── PROMOTIONS-MODULE.md
├── RATINGS-MODULE.md
├── RIDEPIN-MODULE.md
├── RIDERS-MODULE.md
├── RIDES-MODULE.md
├── SERVICEPROVIDERS-MODULE.md
├── SOS-MODULE.md
├── TRACKING-MODULE.md
├── VEHICLES-MODULE.md
└── WALLET-MODULE.md
```

## Module Relationships and Dependencies

### Dependency Map

```
Auth Module
    |
    +-- Profile Module
    +-- Drivers Module
    +-- Riders Module
    +-- Service Providers Module
    |
Rides Module
    |
    +-- Pricing Module (fare calculation)
    +-- Tracking Module (location data)
    +-- Wallet Module (payment processing)
    +-- Messages Module (notifications)
    +-- Ratings Module (post-ride reviews)
    +-- Drivers Module (driver info)
    +-- Riders Module (rider info)
    |
Wallet Module
    |
    +-- Rides Module (ride payments)
    +-- Drivers Module (earnings)
    +-- Service Providers Module (earnings)
    |
Laundry Module
    |
    +-- Pricing Module (weight-based pricing)
    +-- Service Providers Module (provider assignment)
    +-- Ratings Module (order ratings)
    +-- Messages Module (order notifications)
    |
Admin Module
    |
    +-- Service Providers Module (approval coordination)
    +-- Messages Module (notifications)
    |
Fraud Module
    |
    +-- All modules (monitoring)
    |
Messages Module
    |
    +-- All modules (notification distribution)
```

## Development Workflow

### 1. Understanding Module Architecture

Each module follows this standard pattern:

```
Module/
├── handler.go      - HTTP endpoints and request handling
├── service.go      - Business logic and orchestration
├── repository.go   - Database operations and data access
├── routes.go       - Route definitions
└── dto/            - Data Transfer Objects
    ├── requests.go
    └── responses.go
```

### 2. Request Flow

For any HTTP request:

```
Client Request
    |
    v
Handler (Validation & Parameter Extraction)
    |
    v
Service (Business Logic)
    |
    v
Repository (Database Operations)
    |
    v
Database/External Services
    |
    v
Response (via response utility)
    |
    v
Client Response
```

### 3. Adding New Features

Steps to add a new feature to a module:

1. Define DTOs (request/response structures)
2. Update repository interface with new methods
3. Implement repository methods with database logic
4. Update service interface with business logic
5. Implement service methods
6. Add handler endpoints
7. Register routes
8. Add Swagger documentation
9. Write comprehensive tests
10. Update module documentation

### 4. Module Initialization

When initializing a module:

```go
// Create repository with database connection
repo := moduleName.NewRepository(db)

// Create service with repository dependency
service := moduleName.NewService(repo, otherDeps...)

// Create handler with service dependency
handler := moduleName.NewHandler(service)

// Register routes
moduleName.RegisterRoutes(router, handler)
```

## Common Patterns Across Modules

### Error Handling

All modules use standardized error responses:

```go
response.BadRequest("message")
response.NotFoundError("resource name")
response.InternalServerError("message", err)
response.UnauthorizedError("message")
response.ForbiddenError("message")
response.ConflictError("message")
```

### Context Propagation

All database operations use context for proper cancellation:

```go
func (r *repository) SomeOperation(ctx context.Context, ...) {
    return r.db.WithContext(ctx)./* operation */
}
```

### Logging

Consistent logging across modules:

```go
logger.Info("operation description", "key1", value1, "key2", value2)
logger.Error("error description", "error", err)
logger.Debug("debug info", "details", details)
```

### Pagination

Standard pagination pattern:

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
```

## Testing Standards

Each module should include:

### Unit Tests
- Service layer business logic
- Individual method functionality
- Edge cases and error scenarios

### Integration Tests
- Repository database operations
- Multi-table transactions
- External service integrations

### End-to-End Tests
- Complete API endpoints
- Full request/response cycles
- Security and authorization checks

## Security Considerations

All modules must implement:

1. Authentication - Verify user identity via JWT
2. Authorization - Check user roles and permissions
3. Input Validation - Validate all incoming data
4. SQL Injection Prevention - Use parameterized queries (GORM)
5. Data Encryption - Encrypt sensitive data
6. Rate Limiting - Prevent abuse
7. Audit Logging - Log significant operations

## Performance Optimization Guidelines

1. Database Indexing
   - Index frequently queried columns
   - Composite indexes for common filters

2. Caching
   - Cache pricing rules
   - Cache user profiles
   - Cache surge multipliers

3. Query Optimization
   - Use pagination for large datasets
   - Select only needed columns
   - Join optimization

4. Async Processing
   - Use background jobs for non-critical operations
   - Batch settle transactions
   - Async notification sending

5. Connection Pooling
   - Configure appropriate pool size
   - Set connection timeout
   - Monitor pool usage

## Monitoring and Observability

Each module should emit:

1. Metrics
   - Operation success/failure rates
   - Response times
   - Database query times
   - Transaction amounts (wallet)

2. Logs
   - Structured logging with context
   - Error logs with stack traces
   - Audit logs for sensitive operations

3. Traces
   - Request trace IDs
   - Service-to-service calls
   - Database operation traces

## Quick Reference: Module Features

| Module | Purpose | Key Feature | Critical |
|--------|---------|-------------|----------|
| Auth | Authentication | JWT tokens | Yes |
| Rides | Ride management | Real-time tracking | Yes |
| Pricing | Fare calculation | Surge pricing | Yes |
| Wallet | Payment management | Balance tracking | Yes |
| Drivers | Driver management | Document verification | Yes |
| Riders | Rider management | Saved addresses | No |
| Admin | Administration | User management | No |
| Laundry | Laundry orders | Weight-based pricing | No |
| Messages | Notifications | Delivery tracking | No |
| Ratings | Reviews | Fraud detection | No |
| Tracking | Location tracking | ETA calculation | Yes |
| Vehicles | Vehicle management | Registration | No |
| Promotions | Discounts | Code redemption | No |
| SOS | Safety | Emergency contacts | Yes |
| Fraud | Security | Risk scoring | Yes |

## Getting Started as a Developer

1. Start with [MODULES-OVERVIEW.md](./MODULES-OVERVIEW.md)
2. Choose a module to work on
3. Read the specific module documentation
4. Review existing code in the module
5. Understand the architecture pattern
6. Follow established coding standards
7. Write tests for your changes
8. Update documentation
9. Ensure code review before merging

## Resources and References

### Internal Documentation
- See `internal/models` for data models
- See `internal/middleware` for authentication middleware
- See `internal/utils/response` for error handling
- See `internal/utils/logger` for logging utilities
- See `internal/websocket` for WebSocket implementation

### External Standards
- RESTful API best practices
- Go code style guide
- GORM documentation
- Gin framework documentation

## Contributing to Module Documentation

When adding or modifying a module:

1. Update this index
2. Update MODULES-OVERVIEW.md if needed
3. Create/update module-specific documentation
4. Include code examples
5. Document database schema
6. Include error scenarios
7. Add testing guidelines
8. Note integration points
9. Include performance considerations
10. Document configuration options

## FAQ

### Q: How do I add a new endpoint to a module?
A: Add handler method, update service interface, implement service logic, update repository if needed, register route, add Swagger docs, write tests.

### Q: How do I handle errors in my module?
A: Use the standardized response utility functions (response.BadRequest, etc.). Log errors with logger. Return appropriate HTTP status codes.

### Q: How do modules communicate with each other?
A: Through dependency injection. Modules are dependencies of other modules passed via constructor.

### Q: Where do I store sensitive data?
A: Encrypt before storage. Use secure environment variables for secrets. Never log sensitive data.

### Q: How do I test my module?
A: Write unit tests for service layer, integration tests for repository layer, end-to-end tests for handlers. Mock external dependencies.

## Contact and Support

For questions or clarifications about specific modules, refer to the dedicated module documentation or contact the development team.

## Change Log

- Created Module Documentation Index
- Documented 20 core modules
- Added architecture patterns and best practices
- Included quick reference guides
