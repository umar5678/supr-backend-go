## Laundry Service Feature – Outline & Short Documentation

**Purpose**:  
The Laundry Service feature enables customers to request laundry pickup, processing, and delivery through designated laundry facilities. It connects customers with facility providers, manages order lifecycle from pickup through item processing to delivery, tracks individual items with QR codes, and handles service-specific pricing (express options, per-unit/per-hour pricing). This creates a convenient laundry-on-demand experience for the platform.

### Core Features

| Feature                        | Description                                                                                   | Who Uses It                  |
|-------------------------------|-----------------------------------------------------------------------------------------------|------------------------------|
| Service Catalog                | Laundry services with pricing models (per-unit/per-hour), express options, turnaround times. | Customers (discovery)        |
| Order Booking                  | Create orders with pickup date/time, location, service selection, express option.             | Customers                    |
| Facility Matching              | Geo-based search for nearest available facilities; distance calculation.                      | System (real-time)           |
| Pickup Management              | Initiate pickup, complete pickup with photo/notes, track pickup status.                       | Providers                    |
| Item Processing                | Add items with QR codes, track status through wash/dry/press/pack workflow.                   | Providers                    |
| Delivery Management            | Initiate delivery, complete delivery with recipient name and photo.                           | Providers                    |
| Issue Reporting & Resolution   | Report issues (missing items, damage, poor quality), resolve with refunds.                    | Customers & Providers        |
| Pricing Calculation            | Base price + express surcharge + quantity adjustments.                                        | System (real-time)           |
| Order Tracking                 | Real-time status updates from order creation through delivery.                                | Customers                    |

### Folder Structure (Modular Design)

```
internal/modules/laundry/
├── dto/
│   ├── request.go    → All request DTOs (CreateLaundryOrderRequest, CompletePickupRequest, etc.) with Validate()
│   └── response.go   → All response DTOs + converter functions (ToLaundryOrderResponse, etc.)
├── handler.go        → 14 HTTP handlers for customers/providers
├── routes.go         → /laundry group with role middleware
├── service.go        → Business logic (order creation, item tracking, issue resolution)
├── repository.go     → GORM database operations
├── docs.go           → Swagger API documentation
└── outline.md        → This file
```

### API Endpoints Summary

#### Public/Catalog Routes
| Method | Path                            | Auth? | Description                              |
|-------|---------------------------------|-------|------------------------------------------|
| GET   | `/api/v1/laundry/services`      | No    | List all laundry services                |

#### Customer Routes
| Method | Path                              | Auth? | Description                              |
|-------|-----------------------------------|-------|------------------------------------------|
| POST  | `/api/v1/laundry/orders`          | Yes   | Create new laundry order                 |
| GET   | `/api/v1/laundry/orders`          | Yes   | List customer's orders                   |
| GET   | `/api/v1/laundry/orders/{id}`     | Yes   | Get order details                        |
| POST  | `/api/v1/laundry/orders/{id}/issues` | Yes | Report issue with order                  |
| GET   | `/api/v1/laundry/orders/{id}/issues` | Yes | Get order issues                         |
| POST  | `/api/v1/laundry/issues/{id}/resolve` | Yes | Resolve issue (provider/admin)          |

#### Provider Routes
| Method | Path                                   | Description                              |
|-------|----------------------------------------|------------------------------------------|
| POST  | `/api/v1/laundry/orders/{id}/pickup/start` | Initiate pickup                   |
| POST  | `/api/v1/laundry/orders/{id}/pickup/complete` | Complete pickup             |
| GET   | `/api/v1/laundry/provider/pickups`  | List provider's pickup assignments       |
| POST  | `/api/v1/laundry/orders/{id}/items` | Add items to order                       |
| PUT   | `/api/v1/laundry/items/{qrCode}`    | Update item status during processing    |
| GET   | `/api/v1/laundry/orders/{id}/items` | Get order items                         |
| POST  | `/api/v1/laundry/orders/{id}/delivery/start` | Initiate delivery            |
| POST  | `/api/v1/laundry/orders/{id}/delivery/complete` | Complete delivery          |
| GET   | `/api/v1/laundry/provider/deliveries` | List provider's delivery assignments   |
| GET   | `/api/v1/laundry/provider/issues`  | Get issues assigned to provider          |

### Data Models

#### Domain Models (in internal/models/laundry.go)
- **LaundryServiceCatalog**: Service definition with pricing (BasePrice, PricingUnit, TurnaroundHours, ExpressFee, ExpressHours)
- **LaundryFacility**: Facility with location, capacity, operational status
- **ServiceOrder**: Main order entity with status tracking
- **LaundryOrderItem**: Individual items in order with QR code and status
- **LaundryPickup**: Pickup operation with bag count and completion tracking
- **LaundryDelivery**: Delivery operation with recipient info and completion tracking
- **LaundryIssue**: Issue reports with resolution status and refund tracking

#### DTOs (in internal/modules/laundry/dto/)
**Request DTOs**:
- `CreateLaundryOrderRequest`: Order creation with pickup date/time, services, address, coordinates
- `CompletePickupRequest`: Pickup completion with bag count and optional photo
- `AddLaundryItemsRequest`: Add items with type, quantity, service, price
- `UpdateItemStatusRequest`: Update item status through processing workflow
- `CompleteDeliveryRequest`: Delivery completion with recipient name and photo
- `ReportIssueRequest`: Report issue with type and description
- `ResolveIssueRequest`: Resolve issue with resolution details and refund

**Response DTOs**:
- `LaundryOrderResponse`: Complete order with pickup, delivery, items, issues
- `LaundryPickupResponse`: Pickup details
- `LaundryDeliveryResponse`: Delivery details
- `LaundryOrderItemResponse`: Item details with status
- `LaundryIssueResponse`: Issue details with resolution
- `LaundryServiceResponse`: Service catalog entry
- `FacilityDistanceResponse`: Facility with distance information

### Key Design Decisions & Highlights

- **Modular DTO Layer**: Separated request/response DTOs in `dto/` folder with dedicated `Validate()` methods for input validation at handler layer.
- **Service Layer Logging**: Structured logging using project's logger utility with contextual information (customerID, facilityID, coordinates, etc.) for debugging and monitoring.
- **Response Utilities**: All handlers use project's response utility functions (response.Success, response.BadRequest, response.InternalServerError) for consistent error handling.
- **Multi-step Workflow**: Orders flow through distinct phases (creation → pickup → item processing → delivery → completion), each with validation and status tracking.
- **Item Tracking**: QR codes enable granular item tracking through wash/dry/press/pack workflow with status updates.
- **Facility Matching**: Geo-based distance calculation for finding nearest laundry facilities.
- **Issue Management**: Comprehensive issue reporting with type classification (missing_item, damage, poor_cleaning, late_delivery) and resolution with optional refunds.
- **Express Service**: Optional express processing with separate fee and shorter turnaround time.
- **Pricing Models**: Support for per-unit and per-hour pricing with express surcharges.
- **Security**: Role-based access control with customer/provider middleware; ownership verification on operations.
- **Scalability**: Database-driven status tracking; supports concurrent orders and items.

### Request Validation

All request DTOs include `Validate()` methods that perform:
- **Required field checks**: Ensures mandatory fields are present
- **Format validation**: Validates phone numbers, email formats, coordinates
- **Range validation**: Checks numeric bounds (quantity > 0, price > 0, dates in future)
- **Enum validation**: Validates status, issue type against allowed values
- **Cross-field validation**: Ensures related fields are consistent
- **Custom error messages**: Provides specific feedback for each validation failure

### Service Layer Enhancements

- **Structured Logging**: Every major operation (CreateOrder, CompletePickup, ReportIssue) logs key information using project's logger with error context
- **Error Handling**: Wrapped errors with context; validation errors surfaced to clients
- **Data Consistency**: Multi-step operations verify prerequisites before execution
- **Transaction Safety**: Database operations leverage GORM's transaction support

### Dependencies Used

- Gin + binding/validation
- GORM for database operations
- UUID for ID generation
- Project's response utilities for consistent API responses
- Project's logger for structured logging
- Time utilities for date/time handling
- Math utilities for distance calculation

### Integration Points

- **Wallet**: Potential integration for payment holds/captures on order completion
- **User Service**: References customer IDs for order ownership
- **Notification Service**: Potential integration for pickup/delivery notifications
- **Analytics**: Order data for laundry service metrics and reporting
