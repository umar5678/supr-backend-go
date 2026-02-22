# Backend Modules Overview

This document provides a comprehensive guide to the architecture and functionality of all modules in the Supr backend system. Each module follows the standard MVC (Model-View-Controller) pattern with a layered architecture.

## Architecture Pattern

All modules follow a consistent structure:

- **handler.go** - HTTP request handlers and API endpoints
- **service.go** - Business logic and service operations
- **repository.go** - Data access and database operations
- **dto/** - Data Transfer Objects for request/response payloads
- **routes.go** - Route definitions and configuration

## Core Principles

1. Dependency Injection - Dependencies are injected through constructors
2. Interface-based Design - Services and repositories use interfaces for abstraction
3. Context Propagation - All database operations use context for cancellation
4. Error Handling - Standardized error responses using utility functions
5. Logging - Structured logging for debugging and monitoring

## Module List

1. [Admin Module](#admin-module) - System administration and user management
2. [Auth Module](#auth-module) - Authentication and authorization
3. [Batching Module](#batching-module) - Batch operations processing
4. [Drivers Module](#drivers-module) - Driver profile and management
5. [Fraud Module](#fraud-module) - Fraud detection and prevention
6. [Home Services Module](#home-services-module) - Home service operations
7. [Laundry Module](#laundry-module) - Laundry service management
8. [Messages Module](#messages-module) - In-app messaging and notifications
9. [Pricing Module](#pricing-module) - Fare calculation and pricing logic
10. [Profile Module](#profile-module) - User profile management
11. [Promotions Module](#promotions-module) - Promotional codes and discounts
12. [Ratings Module](#ratings-module) - User ratings and reviews
13. [Ride PIN Module](#ride-pin-module) - Ride PIN verification
14. [Riders Module](#riders-module) - Rider profile and management
15. [Rides Module](#rides-module) - Ride request and management
16. [Service Providers Module](#service-providers-module) - Service provider management
17. [SOS Module](#sos-module) - Emergency and safety features
18. [Tracking Module](#tracking-module) - Real-time location tracking
19. [Vehicles Module](#vehicles-module) - Vehicle management
20. [Wallet Module](#wallet-module) - Wallet and payment operations

---

## Admin Module

**Path:** `internal/modules/admin/`

### Overview

The Admin Module provides administrative functionality for system management, including user listing, approval of service providers, user suspension, and dashboard statistics.

### Key Responsibilities

- User management and filtering
- Service provider approval workflows
- User account suspension and status management
- Dashboard statistics generation
- Multi-role administration

### Core Components

#### Handler
- `ListUsers()` - Retrieve paginated list of users with filtering by role and status
- `ApproveServiceProvider()` - Approve pending service provider applications
- `SuspendUser()` - Suspend user accounts with reason logging
- `UpdateUserStatus()` - Modify user account status
- `GetDashboardStats()` - Generate administrative dashboard statistics

#### Service Layer
- User listing with pagination and filtering
- Service provider approval with status updates
- User suspension and status management
- Dashboard statistics aggregation
- Cross-module coordination with service provider module

#### Repository Layer
- Database queries for user data
- User status updates
- User deletion capabilities
- Statistics aggregation by role and status

### DTOs

- `ListUsersResponse` - User list with pagination metadata
- `SuspendUserRequest` - Suspension reason and details

### Security

- Requires BearerAuth token with admin privileges
- All operations are role-based

### Database Operations

- Queries users table with dynamic filtering
- Updates user status with context propagation
- Aggregates statistics by role and status

---

## Auth Module

**Path:** `internal/modules/auth/`

### Overview

The Auth Module handles authentication and authorization for all user types (riders, drivers, service providers). It manages signup, login, token generation, and session management.

### Key Responsibilities

- User signup and account creation
- User login and credential verification
- JWT token generation and validation
- Password management
- Multi-factor authentication
- Token refresh operations

### Core Components

#### Handler
- `PhoneSignup()` - Register new users via phone number
- `PhoneLogin()` - Authenticate users with phone credentials
- `RefreshToken()` - Generate new access tokens
- `Logout()` - Invalidate user sessions
- `VerifyOTP()` - Validate one-time passwords
- `RequestPasswordReset()` - Initiate password recovery

#### Service Layer
- Phone-based signup validation
- Credential verification
- JWT token creation and claims management
- OTP generation and verification
- Password hashing and validation
- Session management

#### Repository Layer
- User creation and retrieval
- Credential storage and verification
- Token blacklisting
- OTP storage and validation

### DTOs

- `PhoneSignupRequest` - User registration data
- `PhoneLoginRequest` - Login credentials
- `AuthResponse` - Token and user information
- `RefreshTokenRequest` - Token refresh payload

### Security Features

- Password hashing with bcrypt
- JWT token with expiration
- OTP-based verification
- Token blacklisting for logout
- Rate limiting on login attempts

### Token Claims

- User ID
- User role
- Email
- Phone number
- Token expiration time

---

## Batching Module

**Path:** `internal/modules/batching/`

### Overview

The Batching Module handles batch processing operations such as bulk user operations, batch updates, and scheduled batch jobs.

### Key Responsibilities

- Batch job creation and management
- Bulk user data processing
- Scheduled batch execution
- Job status tracking
- Error handling in batch operations

### Core Components

#### Handler
- `CreateBatch()` - Initialize new batch operations
- `GetBatchStatus()` - Retrieve batch job status
- `ListBatches()` - List all batch operations
- `CancelBatch()` - Cancel running batch jobs

#### Service Layer
- Batch job validation and creation
- Job status management
- Error recovery and retry logic
- Progress tracking

#### Repository Layer
- Batch job persistence
- Status updates
- Job history retrieval

### Use Cases

- Bulk user imports
- Scheduled promotional campaigns
- Daily report generation
- Data cleanup operations

---

## Drivers Module

**Path:** `internal/modules/drivers/`

### Overview

The Drivers Module manages driver-specific functionality including profile management, document verification, and earnings tracking.

### Key Responsibilities

- Driver profile creation and updates
- Driver document verification (license, insurance, etc.)
- Availability and status management
- Earnings and payment tracking
- Driver rating and performance metrics
- Driver restriction management

### Core Components

#### Handler
- `GetProfile()` - Retrieve driver profile
- `UpdateProfile()` - Modify driver information
- `UploadDocuments()` - Submit verification documents
- `GetEarnings()` - Retrieve earnings information
- `GetRating()` - Fetch driver rating

#### Service Layer
- Profile validation and updates
- Document verification workflow
- Earnings calculations
- Status management
- Performance tracking

#### Repository Layer
- Driver profile persistence
- Document storage and retrieval
- Earnings record management
- Rating updates

### Driver Status Types

- `active` - Available for rides
- `offline` - Not accepting rides
- `on_ride` - Currently on a ride
- `suspended` - Account restricted
- `pending_verification` - Awaiting document verification

### Document Requirements

- Driver's license
- Insurance certificate
- Vehicle registration
- Background check clearance
- Bank account details for payments

---

## Fraud Module

**Path:** `internal/modules/fraud/`

### Overview

The Fraud Module implements fraud detection, prevention mechanisms, and suspicious activity monitoring.

### Key Responsibilities

- Fraud detection algorithm execution
- Suspicious activity flagging
- User behavior analysis
- Payment fraud detection
- Account security monitoring
- Risk scoring

### Core Components

#### Handler
- `ReportSuspiciousActivity()` - Flag suspicious transactions
- `GetRiskScore()` - Calculate user risk score
- `ReviewFraudAlert()` - Admin review of fraud alerts

#### Service Layer
- Pattern-based fraud detection
- Risk assessment calculations
- Activity logging and analysis
- Machine learning integration

#### Repository Layer
- Fraud incident tracking
- Activity log persistence
- Risk score storage

### Fraud Detection Patterns

- Unusual location changes (impossible travel)
- Multiple failed payment attempts
- Abnormal spending patterns
- Account access anomalies
- Device fingerprint changes

---

## Home Services Module

**Path:** `internal/modules/homeservices/`

### Overview

The Home Services Module manages home-based service offerings and operations separate from ride-sharing.

### Key Responsibilities

- Home service request management
- Service provider assignment
- Service completion tracking
- Home service pricing
- Customer feedback on services

### Core Components

#### Handler
- `RequestService()` - Create home service request
- `GetServiceDetails()` - Retrieve service information
- `CompleteService()` - Mark service as completed

#### Service Layer
- Service request validation
- Provider matching and assignment
- Service status tracking
- Rating and feedback processing

#### Repository Layer
- Service request persistence
- Provider assignment records
- Service completion data

---

## Laundry Module

**Path:** `internal/modules/laundry/`

### Overview

The Laundry Module handles laundry service-specific operations including order management, pricing, and service tracking.

### Key Responsibilities

- Laundry order creation and management
- Weight-based pricing calculations
- Service provider (laundryman) assignment
- Order status tracking
- Tip management
- User-provider relationship tracking

### Core Components

#### Handler
- `CreateOrder()` - Submit new laundry order
- `GetOrder()` - Retrieve order details
- `ListOrders()` - Paginated order list
- `UpdateOrderStatus()` - Modify order status
- `AddTip()` - Add tip to completed order
- `RateOrder()` - Submit service rating

#### Service Layer
- Order validation and creation
- Weight-based price calculation
- Provider assignment logic
- Order lifecycle management
- Tip processing
- Earnings calculations

#### Repository Layer
- Order persistence
- Status tracking
- User-provider relationship management
- Tip recording

### Order Status Flow

1. `pending` - Awaiting provider assignment
2. `assigned` - Provider accepted the order
3. `in_progress` - Service being provided
4. `completed` - Service finished
5. `cancelled` - Order cancelled

### Key Features

- Weight-based dynamic pricing
- Optional tip system
- Service provider ratings
- Order history tracking
- User and provider relationship management

---

## Messages Module

**Path:** `internal/modules/messages/`

### Overview

The Messages Module manages in-app messaging between users, system notifications, and communication channels.

### Key Responsibilities

- Direct messaging between users
- System notifications
- Notification delivery and tracking
- Message history management
- Push notification integration

### Core Components

#### Handler
- `SendMessage()` - Send direct message
- `GetMessages()` - Retrieve message history
- `MarkAsRead()` - Mark messages as read
- `SendNotification()` - Send system notification

#### Service Layer
- Message validation and routing
- Notification creation and delivery
- Message encryption (if applicable)
- Read status management

#### Repository Layer
- Message storage
- Notification records
- Read status tracking

### Message Types

- Direct messages (user-to-user)
- System notifications (platform-to-user)
- Ride notifications
- Order updates
- Promotional messages

---

## Pricing Module

**Path:** `internal/modules/pricing/`

### Overview

The Pricing Module calculates fares, surge pricing, and manages dynamic pricing rules for ride services.

### Key Responsibilities

- Fare estimation and calculation
- Surge pricing multiplier determination
- Base fare and distance/time rate management
- Pricing rule management
- Cancellation fee calculations

### Core Components

#### Handler
- `GetFareEstimate()` - Calculate estimated fare
- `GetSurgeMultiplier()` - Get current surge multiplier
- `GetPricingRules()` - Retrieve active pricing rules
- `UpdatePricingRules()` - Modify pricing configurations

#### Service Layer
- Fare calculation algorithm
- Surge multiplier computation (demand/supply-based)
- Pricing rule application
- Distance and time-based pricing
- Special pricing for promotions

#### Repository Layer
- Pricing rule persistence
- Historical pricing data
- Surge pricing rule storage

### Fare Calculation

Base Fare + (Distance * Distance Rate) + (Time * Time Rate) * Surge Multiplier

### Surge Pricing

Calculated based on:
- Current demand vs available drivers
- Time of day
- Weather conditions
- Special events

---

## Profile Module

**Path:** `internal/modules/profile/`

### Overview

The Profile Module manages user profile information across all user types (riders, drivers, service providers).

### Key Responsibilities

- User profile creation and management
- Profile information updates
- Avatar and media management
- Preference management
- Privacy settings

### Core Components

#### Handler
- `GetProfile()` - Retrieve user profile
- `UpdateProfile()` - Modify profile information
- `UploadAvatar()` - Update profile picture
- `GetPreferences()` - Retrieve user preferences
- `UpdatePreferences()` - Modify user preferences

#### Service Layer
- Profile data validation
- Image processing and storage
- Preference management
- Profile completeness tracking

#### Repository Layer
- Profile data persistence
- Media storage references
- Preference storage

### Profile Types

- Rider profile
- Driver profile
- Service provider profile
- Admin profile

---

## Promotions Module

**Path:** `internal/modules/promotions/`

### Overview

The Promotions Module manages promotional codes, discounts, and marketing campaigns.

### Key Responsibilities

- Promotional code creation and validation
- Discount application
- Campaign management
- Usage tracking
- Redemption processing

### Core Components

#### Handler
- `CreatePromotion()` - Create new promo code
- `ValidateCode()` - Check promo code validity
- `ApplyPromotion()` - Apply discount to transaction
- `GetActivePromotions()` - List current promotions

#### Service Layer
- Promo code validation and application
- Discount calculation
- Campaign eligibility checking
- Usage limit enforcement
- Expiration handling

#### Repository Layer
- Promotion storage
- Usage tracking
- Redemption records

### Promotion Types

- Percentage-based discounts
- Fixed amount discounts
- Free ride credits
- Seasonal promotions
- User-specific offers

---

## Ratings Module

**Path:** `internal/modules/ratings/`

### Overview

The Ratings Module manages user ratings and reviews for rides, services, and providers.

### Key Responsibilities

- Rating submission and storage
- Review content management
- Rating aggregation and statistics
- Fraud detection in ratings
- Display and retrieval of ratings

### Core Components

#### Handler
- `SubmitRating()` - Submit rating and review
- `GetRating()` - Retrieve rating details
- `GetAverageRating()` - Get provider's average rating
- `FlagInappropriate()` - Report inappropriate ratings

#### Service Layer
- Rating validation
- Aggregate calculation (average, count)
- Review moderation logic
- Fraud detection in ratings

#### Repository Layer
- Rating storage
- Review persistence
- Rating history

### Rating Scale

- 1-5 star rating system
- Optional written review
- Photo attachments (optional)
- Categories (cleanliness, behavior, etc.)

---

## Ride PIN Module

**Path:** `internal/modules/ridepin/`

### Overview

The Ride PIN Module manages PIN-based verification for ride security and authentication.

### Key Responsibilities

- PIN generation and validation
- OTP-based verification
- Ride verification before pickup/completion
- PIN expiration management

### Core Components

#### Handler
- `GeneratePin()` - Create new ride PIN
- `VerifyPin()` - Validate entered PIN
- `ResendPin()` - Resend PIN to user

#### Service Layer
- PIN generation and encryption
- Verification logic
- Expiration handling
- Attempt tracking

#### Repository Layer
- PIN storage and retrieval
- Verification attempts logging

### PIN Security

- 4-6 digit random PIN
- Time-limited validity (5-15 minutes)
- Limited verification attempts
- Attempt tracking for fraud detection

---

## Riders Module

**Path:** `internal/modules/riders/`

### Overview

The Riders Module manages rider-specific functionality and user experience features.

### Key Responsibilities

- Rider profile management
- Saved addresses (home, work, favorites)
- Preferred drivers and ratings
- Rider preferences and settings
- Emergency contacts

### Core Components

#### Handler
- `GetProfile()` - Get rider profile
- `UpdateProfile()` - Modify rider details
- `SaveAddress()` - Store favorite location
- `GetSavedAddresses()` - List saved locations
- `GetRiderHistory()` - Retrieve ride history

#### Service Layer
- Profile validation
- Address management
- Preference handling
- History retrieval

#### Repository Layer
- Rider profile storage
- Saved addresses persistence
- Emergency contact management

### Key Features

- Multiple saved addresses
- Ride history with filters
- Favorite drivers tracking
- Emergency contact list
- Accessibility preferences

---

## Rides Module

**Path:** `internal/modules/rides/`

### Overview

The Rides Module is the core of the ride-sharing functionality, managing ride requests, driver assignments, ride tracking, and completion.

### Key Responsibilities

- Ride request creation and management
- Driver matching and assignment
- Real-time ride tracking
- Ride status management
- Ride completion and settlement
- WebSocket communication for live updates

### Core Components

#### Handler
- `CreateRide()` - Submit new ride request
- `GetRide()` - Retrieve ride details
- `ListRides()` - List user's rides
- `CancelRide()` - Cancel pending ride
- `CompleteRide()` - Mark ride as completed
- `UpdateRideStatus()` - Modify ride status
- `GetRideTracking()` - Real-time ride tracking
- `WebSocketHandler()` - WebSocket connection for live updates

#### Service Layer
- Ride validation and creation
- Driver matching algorithm
- Ride status transitions
- Distance and duration calculations
- Fare finalization
- ETA calculations
- Real-time location handling

#### Repository Layer
- Ride persistence
- Status updates
- Location data storage
- Ride history retrieval

### Ride Status Flow

1. `requested` - Awaiting driver acceptance
2. `accepted` - Driver assigned
3. `arrived` - Driver at pickup location
4. `started` - Ride in progress
5. `completed` - Ride finished
6. `cancelled` - Ride cancelled

### Key Features

- Real-time GPS tracking
- ETA estimation
- Dynamic fare calculation
- WebSocket-based live updates
- Driver assignment algorithm
- Ride history tracking

### WebSocket Events

- `ride_created` - New ride request
- `ride_accepted` - Driver accepted
- `driver_location_updated` - GPS update
- `ride_started` - Ride started
- `ride_completed` - Ride finished
- `ride_cancelled` - Ride cancelled

---

## Service Providers Module

**Path:** `internal/modules/serviceproviders/`

### Overview

The Service Providers Module manages service provider accounts and operations including profile management, service listings, and earnings.

### Key Responsibilities

- Service provider profile management
- Service listing and management
- Provider verification and approval
- Service availability management
- Earnings and payment tracking
- Provider performance metrics

### Core Components

#### Handler
- `GetProfile()` - Retrieve provider profile
- `UpdateProfile()` - Modify provider information
- `ListServices()` - Get available services
- `UpdateServiceAvailability()` - Change service status
- `GetEarnings()` - Retrieve earnings data

#### Service Layer
- Profile validation and updates
- Service management
- Approval workflow coordination
- Earnings calculation
- Performance tracking

#### Repository Layer
- Provider profile persistence
- Service listings storage
- Earnings records
- Availability management

### Service Provider Types

- Laundryman
- Home service provider
- Delivery partner
- Other service providers

### Verification Requirements

- Identity verification
- Service qualifications
- Background check
- Insurance documentation
- Banking information

---

## SOS Module

**Path:** `internal/modules/sos/`

### Overview

The SOS Module implements emergency and safety features including emergency contacts, panic buttons, and incident reporting.

### Key Responsibilities

- Emergency contact management
- Panic button/SOS triggering
- Emergency alert distribution
- Incident reporting
- Safety feature activation
- Emergency response tracking

### Core Components

#### Handler
- `AddEmergencyContact()` - Add contact to emergency list
- `GetEmergencyContacts()` - List emergency contacts
- `TriggerSOS()` - Activate emergency alert
- `ReportIncident()` - File safety incident
- `GetSOSStatus()` - Check SOS status

#### Service Layer
- Contact validation
- Emergency alert creation and distribution
- Incident logging
- Response coordination
- Location sharing for emergencies

#### Repository Layer
- Emergency contact storage
- SOS incident records
- Alert distribution logs

### Safety Features

- Quick access emergency contacts
- Real-time location sharing during emergency
- Automated alert distribution
- Police/Support services integration
- Incident documentation

---

## Tracking Module

**Path:** `internal/modules/tracking/`

### Overview

The Tracking Module provides real-time location tracking for rides and services using WebSocket connections and GPS data.

### Key Responsibilities

- Real-time location updates
- Location history storage
- Distance and duration calculations
- ETA estimation
- Geofencing
- Route optimization

### Core Components

#### Handler
- `UpdateLocation()` - Receive GPS update
- `GetCurrentLocation()` - Get current position
- `GetLocationHistory()` - Retrieve location history
- `EstimateArrival()` - Calculate ETA

#### Service Layer
- Location validation
- Distance calculations
- ETA computation
- Route optimization
- Geofence checking

#### Repository Layer
- Location history persistence
- Tracking data storage
- ETA cache management

### Location Tracking

- Real-time GPS updates via WebSocket
- Location history for all rides
- Driver location sharing with rider
- Pickup and dropoff location validation

### Geofencing

- Ride boundary definition
- Zone-based pricing
- Location alerts

---

## Vehicles Module

**Path:** `internal/modules/vehicles/`

### Overview

The Vehicles Module manages vehicle information, registration, and driver-vehicle relationships.

### Key Responsibilities

- Vehicle registration and management
- Driver-vehicle association
- Vehicle documentation (registration, insurance)
- Vehicle type management
- Vehicle status and availability

### Core Components

#### Handler
- `RegisterVehicle()` - Register new vehicle
- `GetVehicle()` - Retrieve vehicle details
- `UpdateVehicle()` - Modify vehicle information
- `ListDriverVehicles()` - Get driver's vehicles
- `VerifyDocuments()` - Verify vehicle documents

#### Service Layer
- Vehicle validation
- Document verification
- Driver-vehicle association
- Vehicle status management
- Availability tracking

#### Repository Layer
- Vehicle data persistence
- Document storage references
- Driver-vehicle relationships

### Vehicle Information

- License plate
- VIN (Vehicle Identification Number)
- Make and model
- Year of manufacture
- Registration number
- Insurance details
- Seating capacity
- Vehicle type/category

### Vehicle Types

- Sedan
- SUV
- Auto-rickshaw
- Bike
- Van

---

## Wallet Module

**Path:** `internal/modules/wallet/`

### Overview

The Wallet Module manages user wallets, payment transactions, balance management, and financial operations.

### Key Responsibilities

- Wallet creation and management
- Balance tracking
- Transaction processing
- Funds addition and withdrawal
- Payment processing
- Transaction history

### Core Components

#### Handler
- `GetWallet()` - Retrieve wallet details
- `GetBalance()` - Get current balance
- `AddFunds()` - Add money to wallet
- `WithdrawFunds()` - Withdraw money
- `GetTransactionHistory()` - List transactions
- `ProcessPayment()` - Execute payment

#### Service Layer
- Wallet initialization
- Balance management
- Transaction validation
- Payment processing
- Refund handling
- Settlement calculations

#### Repository Layer
- Wallet data persistence
- Transaction records
- Balance history

### Wallet Types

- Rider wallet (prepaid balance)
- Driver wallet (cash in hand, earnings)
- Service provider wallet (earnings)

### Transaction Types

- Top-up (adding funds)
- Ride payment
- Service payment
- Refund
- Withdrawal
- Tip
- Bonus/Promotion

### Key Features

- Instant balance updates
- Transaction history with filters
- Wallet-to-wallet transfers
- Payment method integration
- Automatic settlements
- Cash-based driver wallets

---

## Cross-Module Communication

Modules interact through dependency injection and shared interfaces:

- Admin module coordinates with service providers for approvals
- Rides module uses Pricing module for fare calculations
- Rides module uses Tracking module for location updates
- Wallet module handles payments from multiple modules
- Messages module sends notifications from all other modules
- Ratings module receives inputs from Rides and other service modules

## Error Handling Standards

All modules use standardized error responses:

```
response.BadRequest() - 400 Bad Request
response.NotFoundError() - 404 Not Found
response.InternalServerError() - 500 Internal Server Error
response.UnauthorizedError() - 401 Unauthorized
response.ForbiddenError() - 403 Forbidden
```

## Testing Strategy

Each module should include:

- Unit tests for business logic (service layer)
- Integration tests for database operations (repository layer)
- End-to-end tests for API endpoints (handler layer)
- Mock implementations for dependencies

## Development Guidelines

1. Always use interfaces for abstraction
2. Implement context propagation for database operations
3. Use proper error handling and logging
4. Add Swagger/OpenAPI documentation
5. Follow Go naming conventions
6. Write comprehensive comments for exported functions
7. Validate inputs at handler level
8. Keep business logic in service layer
9. Use repository pattern for data access
10. Implement proper transaction management
