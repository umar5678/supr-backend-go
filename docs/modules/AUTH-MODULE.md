# Auth Module Development Guide

## Overview

The Auth Module handles all authentication and authorization operations including user signup, login, token management, and session handling. It supports multiple authentication methods and manages JWT tokens.

## Module Structure

```
auth/
├── handler.go         # HTTP request handlers
├── service.go         # Authentication business logic
├── repository.go      # Database operations
├── routes.go          # Route definitions
└── dto/
    ├── requests.go    # Request payloads
    └── responses.go   # Response structures
```

## Key Responsibilities

1. User Registration - Phone-based signup with validation
2. User Authentication - Credential verification and login
3. Token Management - JWT generation, validation, and refresh
4. OTP Handling - One-time password generation and verification
5. Session Management - User session tracking and logout
6. Password Management - Password hashing and recovery

## Architecture

### Handler Layer (handler.go)

Manages HTTP endpoints for authentication operations.

Key methods:

```
PhoneSignup(c *gin.Context)              // POST /auth/phone/signup
PhoneLogin(c *gin.Context)               // POST /auth/phone/login
RefreshToken(c *gin.Context)             // POST /auth/refresh
Logout(c *gin.Context)                   // POST /auth/logout
VerifyOTP(c *gin.Context)                // POST /auth/verify-otp
RequestPasswordReset(c *gin.Context)     // POST /auth/password-reset/request
ResetPassword(c *gin.Context)            // POST /auth/password-reset/confirm
```

Request flow:
1. Extract credentials from request body
2. Validate basic format and required fields
3. Call service method
4. Return token in response

### Service Layer (service.go)

Contains authentication business logic.

Key interface methods:

```
PhoneSignup(ctx context.Context, req PhoneSignupRequest) (*AuthResponse, error)
PhoneLogin(ctx context.Context, req PhoneLoginRequest) (*AuthResponse, error)
RefreshToken(ctx context.Context, refreshToken string) (*AuthResponse, error)
Logout(ctx context.Context, userID string) error
VerifyOTP(ctx context.Context, phone, otp string) (*models.User, error)
RequestPasswordReset(ctx context.Context, email string) error
ResetPassword(ctx context.Context, resetToken, newPassword string) error
ValidateToken(ctx context.Context, token string) (*models.User, error)
```

Logic flow:
1. Validate input parameters
2. Check for existing users/conflicts
3. Hash passwords using bcrypt
4. Generate JWT tokens
5. Generate OTP if needed
6. Store session information
7. Return auth response with tokens

### Repository Layer (repository.go)

Handles database operations for authentication.

Key interface methods:

```
CreateUser(ctx context.Context, user *models.User) error
FindUserByPhone(ctx context.Context, phone string) (*models.User, error)
FindUserByEmail(ctx context.Context, email string) (*models.User, error)
UpdatePassword(ctx context.Context, userID, hashedPassword string) error
CreateOTP(ctx context.Context, phone, otp string, expiresAt time.Time) error
FindOTP(ctx context.Context, phone string) (*models.OTP, error)
InvalidateOTP(ctx context.Context, phone string) error
CreateSession(ctx context.Context, userID string, token string) error
InvalidateSession(ctx context.Context, userID string) error
FindSession(ctx context.Context, userID string) (*models.Session, error)
```

Database operations:
- Use transactions for critical operations
- Implement proper indexing on frequently queried fields
- Store password hashes securely
- Expire OTPs automatically

## Data Transfer Objects

### PhoneSignupRequest

```go
type PhoneSignupRequest struct {
    Phone       string `json:"phone" binding:"required,phone"`
    Name        string `json:"name" binding:"required"`
    Email       string `json:"email" binding:"required,email"`
    UserType    string `json:"user_type" binding:"required,oneof=rider driver service_provider"`
    Password    string `json:"password" binding:"required,min=8"`
}
```

### PhoneLoginRequest

```go
type PhoneLoginRequest struct {
    Phone    string `json:"phone" binding:"required,phone"`
    Password string `json:"password" binding:"required"`
}
```

### AuthResponse

```go
type AuthResponse struct {
    User         *UserInfo `json:"user"`
    AccessToken  string    `json:"access_token"`
    RefreshToken string    `json:"refresh_token"`
    ExpiresIn    int       `json:"expires_in"` // Seconds
    TokenType    string    `json:"token_type"` // "Bearer"
}

type UserInfo struct {
    ID       string `json:"id"`
    Email    string `json:"email"`
    Phone    string `json:"phone"`
    Name     string `json:"name"`
    UserType string `json:"user_type"`
}
```

### VerifyOTPRequest

```go
type VerifyOTPRequest struct {
    Phone string `json:"phone" binding:"required"`
    OTP   string `json:"otp" binding:"required,len=6"`
}
```

### RefreshTokenRequest

```go
type RefreshTokenRequest struct {
    RefreshToken string `json:"refresh_token" binding:"required"`
}
```

## Typical Use Cases

### 1. User Signup

Request:
```
POST /auth/phone/signup
{
    "phone": "+1234567890",
    "name": "John Doe",
    "email": "john@example.com",
    "user_type": "rider",
    "password": "securePassword123"
}
```

Flow:
1. Validate phone format and uniqueness
2. Validate email format and uniqueness
3. Validate password strength
4. Hash password using bcrypt
5. Create user record in database
6. Generate JWT access and refresh tokens
7. Create session record
8. Return auth response with tokens

### 2. User Login

Request:
```
POST /auth/phone/login
{
    "phone": "+1234567890",
    "password": "securePassword123"
}
```

Flow:
1. Validate phone format
2. Find user by phone
3. Verify password hash matches
4. Generate new JWT tokens
5. Update/create session record
6. Return auth response with new tokens

### 3. Token Refresh

Request:
```
POST /auth/refresh
{
    "refresh_token": "eyJhbGciOiJIUzI1NiIs..."
}
```

Flow:
1. Validate refresh token signature and expiration
2. Extract user ID from refresh token claims
3. Find user in database
4. Generate new access token
5. Keep refresh token unchanged (or regenerate)
6. Return new tokens

### 4. OTP Verification

Request:
```
POST /auth/verify-otp
{
    "phone": "+1234567890",
    "otp": "123456"
}
```

Flow:
1. Find OTP record for phone
2. Validate OTP hasn't expired
3. Compare provided OTP with stored OTP
4. Invalidate OTP after successful verification
5. Return user information

### 5. Password Reset

Request:
```
POST /auth/password-reset/request
{
    "email": "john@example.com"
}
```

Flow:
1. Find user by email
2. Generate reset token
3. Store reset token with expiration
4. Send reset link via email
5. Return success response

## JWT Token Structure

### Access Token Claims

```
{
    "user_id": "user-123",
    "email": "john@example.com",
    "phone": "+1234567890",
    "user_type": "rider",
    "exp": 1234567890,
    "iat": 1234567800,
    "iss": "supr-backend"
}
```

### Refresh Token Claims

```
{
    "user_id": "user-123",
    "exp": 1234567890,
    "iat": 1234567800,
    "iss": "supr-backend",
    "type": "refresh"
}
```

## Token Configuration

Typical configuration:

```yaml
auth:
  jwt:
    secret_key: "your-secret-key-here"
    access_token_expiry: "15m"    # 15 minutes
    refresh_token_expiry: "7d"    # 7 days
  otp:
    length: 6
    expiry: "5m"                  # 5 minutes
    max_attempts: 3
  password:
    min_length: 8
    require_uppercase: true
    require_numbers: true
    require_special_chars: false
```

## Security Considerations

1. Password Hashing
   - Use bcrypt with appropriate cost factor
   - Never store plain-text passwords
   - Salt passwords automatically in bcrypt

2. Token Security
   - Sign tokens with strong secret key
   - Set reasonable expiration times
   - Implement token blacklisting for logout
   - Validate token signature on each request

3. OTP Security
   - Generate cryptographically secure random OTPs
   - Limit OTP attempt count
   - Expire OTPs quickly (5 minutes)
   - Log failed OTP attempts

4. Input Validation
   - Validate phone format
   - Validate email format
   - Validate password strength
   - Check for injection attacks

5. Rate Limiting
   - Limit login attempts per IP/phone
   - Limit OTP verification attempts
   - Limit password reset requests

## Error Handling

Common error scenarios:

1. Duplicate Phone/Email
   - Response: 409 Conflict
   - Message: "Phone number already registered"

2. Invalid Credentials
   - Response: 401 Unauthorized
   - Message: "Invalid phone or password"

3. User Not Found
   - Response: 404 Not Found
   - Message: "User not found"

4. Invalid OTP
   - Response: 400 Bad Request
   - Message: "Invalid or expired OTP"

5. Token Expired
   - Response: 401 Unauthorized
   - Message: "Token expired, please refresh"

6. Rate Limit Exceeded
   - Response: 429 Too Many Requests
   - Message: "Too many attempts, please try again later"

## Testing Strategy

### Unit Tests (Service Layer)

```go
Test_PhoneSignup_Success()
Test_PhoneSignup_DuplicatePhone()
Test_PhoneLogin_ValidCredentials()
Test_PhoneLogin_InvalidPassword()
Test_RefreshToken_ValidToken()
Test_RefreshToken_ExpiredToken()
Test_GenerateOTP()
Test_VerifyOTP_Success()
Test_VerifyOTP_MaxAttemptsExceeded()
```

### Integration Tests (Repository Layer)

```go
Test_CreateUser_Success()
Test_FindUserByPhone()
Test_UpdatePassword()
Test_CreateAndFind_OTP()
```

### End-to-End Tests (Handler Layer)

```go
Test_PhoneSignup_FullFlow()
Test_PhoneLogin_FullFlow()
Test_TokenRefresh_FullFlow()
```

## Database Schema

### Users Table

```sql
CREATE TABLE users (
    id VARCHAR(36) PRIMARY KEY,
    email VARCHAR(255) UNIQUE,
    phone VARCHAR(20) UNIQUE,
    password_hash VARCHAR(255),
    name VARCHAR(255),
    user_type VARCHAR(50),
    status VARCHAR(50),
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);
```

### OTP Table

```sql
CREATE TABLE otps (
    id VARCHAR(36) PRIMARY KEY,
    phone VARCHAR(20),
    otp VARCHAR(10),
    attempts INT DEFAULT 0,
    expires_at TIMESTAMP,
    created_at TIMESTAMP
);
```

### Sessions Table

```sql
CREATE TABLE sessions (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36),
    refresh_token TEXT,
    expires_at TIMESTAMP,
    created_at TIMESTAMP
);
```

### Password Reset Table

```sql
CREATE TABLE password_resets (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36),
    reset_token VARCHAR(255),
    expires_at TIMESTAMP,
    created_at TIMESTAMP
);
```

## Integration Points

1. Profile Module - For user profile creation after signup
2. Wallet Module - For wallet initialization
3. Notification Module - For OTP and reset email delivery
4. Audit Module - For login/logout logging

## Related Documentation

- See MODULES-OVERVIEW.md for module architecture overview
- See internal/models documentation for user data models
- See internal/middleware documentation for token validation
- See internal/utils/response for error handling patterns

## Common Pitfalls

1. Weak password hashing - Always use bcrypt with proper cost
2. Token signature not verified - Validate on every request
3. Exposure of secret keys - Use environment variables
4. No rate limiting - Implement to prevent brute force
5. Storing plain-text passwords - Never store unencrypted
6. Not handling token expiration - Refresh tokens proactively
7. Insufficient logging - Log all auth failures for security
8. SQL injection in queries - Use parameterized queries (GORM)

## Future Enhancements

1. Multi-factor authentication (SMS/Email)
2. Social login (Google, Facebook)
3. Biometric authentication
4. Device fingerprinting
5. Session management across devices
6. OAuth2/OpenID Connect support
7. Two-factor authentication
8. Passwordless authentication

## Performance Optimization

1. Cache user lookups with Redis
2. Use database connection pooling
3. Implement JWT caching
4. Batch OTP cleanup operations
5. Index on phone and email columns
6. Use prepared statements

## Maintenance

Regular tasks:
1. Rotate JWT secret keys periodically
2. Clean up expired OTPs and password resets
3. Audit login patterns for anomalies
4. Review and update password policies
5. Monitor failed login attempts
6. Update dependencies for security patches
