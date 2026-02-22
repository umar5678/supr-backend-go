# Laundry Module Development Guide

## Overview

The Laundry Module handles laundry service-specific operations including order management, weight-based pricing, service provider assignment, order status tracking, and tip management.

## Module Structure

```
laundry/
├── handler.go         # HTTP request handlers
├── service.go         # Business logic
├── repository.go      # Database operations
├── routes.go          # Route definitions
└── dto/
    ├── requests.go    # Request payloads
    └── responses.go   # Response structures
```

## Key Responsibilities

1. Order Management - Create and manage laundry orders
2. Weight-based Pricing - Calculate fees based on weight
3. Service Provider Assignment - Match orders to laundrymen
4. Status Tracking - Monitor order lifecycle
5. Tip Management - Handle tip additions
6. Rating and Feedback - Collect service feedback

## Architecture

### Handler Layer (handler.go)

Key methods:

```
CreateOrder(c *gin.Context)             // POST /laundry/orders
GetOrder(c *gin.Context)                // GET /laundry/orders/{id}
ListOrders(c *gin.Context)              // GET /laundry/orders
UpdateOrderStatus(c *gin.Context)       // PUT /laundry/orders/{id}/status
AddTip(c *gin.Context)                  // POST /laundry/orders/{id}/tip
RateOrder(c *gin.Context)               // POST /laundry/orders/{id}/rate
CancelOrder(c *gin.Context)             // POST /laundry/orders/{id}/cancel
TrackOrder(c *gin.Context)              // GET /laundry/orders/{id}/track
```

### Service Layer (service.go)

Key interface methods:

```
CreateOrder(ctx context.Context, userID string, req CreateOrderRequest) (*LaundryOrderResponse, error)
GetOrder(ctx context.Context, userID, orderID string) (*LaundryOrderResponse, error)
ListOrders(ctx context.Context, userID string, filters map[string]interface{}) ([]*LaundryOrderResponse, error)
UpdateOrderStatus(ctx context.Context, orderID string, status OrderStatus) error
AddTip(ctx context.Context, orderID string, tipAmount float64) error
RateOrder(ctx context.Context, orderID string, rating *OrderRating) error
CancelOrder(ctx context.Context, orderID string) error
AssignProvider(ctx context.Context, orderID, providerID string) error
CalculatePrice(ctx context.Context, weight float64) (float64, error)
```

Logic flow:
1. Validate order details
2. Calculate price based on weight
3. Create order record
4. Assign available laundryman
5. Send notifications
6. Update status transitions
7. Process payments

### Repository Layer (repository.go)

Key interface methods:

```
CreateOrder(ctx context.Context, order *models.LaundryOrder) error
FindOrderByID(ctx context.Context, orderID string) (*models.LaundryOrder, error)
ListOrders(ctx context.Context, userID string, filters map[string]interface{}) ([]*models.LaundryOrder, error)
UpdateOrderStatus(ctx context.Context, orderID string, status OrderStatus) error
AddTip(ctx context.Context, orderID string, amount float64) error
SaveRating(ctx context.Context, rating *models.LaundryOrderRating) error
CancelOrder(ctx context.Context, orderID string) error
```

## Data Transfer Objects

### CreateOrderRequest

```go
type CreateOrderRequest struct {
    Weight              float64 `json:"weight" binding:"required,gt=0"`
    ClothType           string  `json:"cloth_type" binding:"required"` // regular, delicate, wool
    PickupAddress       Address `json:"pickup_address" binding:"required"`
    PickupTime          time.Time `json:"pickup_time" binding:"required"`
    DropoffTime         time.Time `json:"dropoff_time,omitempty"`
    SpecialInstructions string  `json:"special_instructions,omitempty"`
    PaymentMethod       string  `json:"payment_method" binding:"required"`
    PreferredProvider   string  `json:"preferred_provider,omitempty"`
}
```

### LaundryOrderResponse

```go
type LaundryOrderResponse struct {
    ID                  string           `json:"id"`
    UserID              string           `json:"user_id"`
    ProviderID          string           `json:"provider_id,omitempty"`
    Weight              float64          `json:"weight"`
    ClothType           string           `json:"cloth_type"`
    Status              string           `json:"status"`
    EstimatedPrice      float64          `json:"estimated_price"`
    ActualPrice         float64          `json:"actual_price,omitempty"`
    Tip                 float64          `json:"tip,omitempty"`
    TotalAmount         float64          `json:"total_amount"`
    PickupAddress       Address          `json:"pickup_address"`
    PickupTime          time.Time        `json:"pickup_time"`
    DropoffTime         time.Time        `json:"dropoff_time,omitempty"`
    CompletedTime       *time.Time       `json:"completed_time,omitempty"`
    Rating              *OrderRating     `json:"rating,omitempty"`
    ProviderInfo        *ProviderInfo    `json:"provider_info,omitempty"`
    SpecialInstructions string          `json:"special_instructions,omitempty"`
    CreatedAt           time.Time        `json:"created_at"`
    UpdatedAt           time.Time        `json:"updated_at"`
}
```

### OrderRating

```go
type OrderRating struct {
    Stars      int    `json:"stars" binding:"required,min=1,max=5"`
    Review     string `json:"review,omitempty"`
    Punctuality int   `json:"punctuality,omitempty"` // 1-5
    Quality    int    `json:"quality,omitempty"` // 1-5
    Packaging  int    `json:"packaging,omitempty"` // 1-5
}
```

## Pricing Model

### Weight-Based Pricing

```
Base Fee: 5.00
Price per KG: 2.50
Cloth Type Multiplier:
  - Regular: 1.0x
  - Delicate: 1.5x
  - Wool: 1.8x

Formula:
EstimatedPrice = BaseFee + (Weight * PricePerKG * ClothTypeMultiplier)
```

### Example Calculations

```
10 KG Regular Clothes:
= 5.00 + (10 * 2.50 * 1.0)
= 5.00 + 25.00
= 30.00

5 KG Delicate Clothes:
= 5.00 + (5 * 2.50 * 1.5)
= 5.00 + 18.75
= 23.75
```

## Order Status Flow

```
PENDING
    |
    v
ASSIGNED (laundryman assigned)
    |
    v
PICKED_UP (clothes collected)
    |
    v
IN_PROGRESS (being washed)
    |
    v
READY (ready for delivery)
    |
    v
DELIVERED (completed)

Cancellation paths:
PENDING -> CANCELLED
ASSIGNED -> CANCELLED (with charges)
```

## Typical Use Cases

### 1. Create Laundry Order

Request:
```
POST /laundry/orders
{
    "weight": 10.5,
    "cloth_type": "regular",
    "pickup_address": {
        "latitude": 40.7128,
        "longitude": -74.0060,
        "address": "123 Main St, NYC"
    },
    "pickup_time": "2024-02-21T10:00:00Z",
    "special_instructions": "Handle with care",
    "payment_method": "wallet"
}
```

Flow:
1. Validate weight and cloth type
2. Calculate estimated price
3. Create order with status PENDING
4. Find available laundryman
5. Assign order
6. Send notification to user and provider
7. Return order details

### 2. Add Tip to Order

Request:
```
POST /laundry/orders/{orderID}/tip
{
    "tip_amount": 5.00
}
```

Flow:
1. Find order by ID
2. Verify order is completed
3. Add tip amount
4. Create transaction for tip
5. Notify provider of tip
6. Update order total

### 3. Rate Order

Request:
```
POST /laundry/orders/{orderID}/rate
{
    "stars": 5,
    "review": "Excellent service!",
    "punctuality": 5,
    "quality": 5,
    "packaging": 5
}
```

Flow:
1. Find order by ID
2. Verify user is order owner
3. Create rating record
4. Update provider average rating
5. Update order with rating
6. Return confirmation

### 4. Update Order Status

Request (Provider):
```
PUT /laundry/orders/{orderID}/status
{
    "status": "picked_up"
}
```

Flow:
1. Verify provider owns order
2. Validate status transition
3. Update order status
4. Send notification to user
5. Log status change

## Error Handling

Common error scenarios:

1. Insufficient Weight
   - Response: 400 Bad Request
   - Message: "Weight must be greater than 0"

2. No Available Providers
   - Response: 202 Accepted
   - Action: Queue order, notify when available

3. Invalid Cloth Type
   - Response: 400 Bad Request
   - Message: "Cloth type must be: regular, delicate, or wool"

4. Order Already Completed
   - Response: 409 Conflict
   - Message: "Cannot modify completed order"

## Testing Strategy

### Unit Tests

```go
Test_CalculatePrice_Regular()
Test_CalculatePrice_Delicate()
Test_CreateOrder_Success()
Test_AddTip_OnCompletedOrder()
Test_RateOrder_ValidRating()
Test_CancelOrder_Refund()
```

## Database Schema

### Laundry Orders Table

```sql
CREATE TABLE laundry_orders (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    provider_id VARCHAR(36),
    weight DECIMAL(10, 2),
    cloth_type VARCHAR(50),
    status VARCHAR(50),
    estimated_price DECIMAL(10, 2),
    actual_price DECIMAL(10, 2),
    tip DECIMAL(10, 2) DEFAULT 0,
    total_amount DECIMAL(10, 2),
    pickup_address JSON,
    pickup_time TIMESTAMP,
    dropoff_time TIMESTAMP,
    completed_time TIMESTAMP,
    special_instructions TEXT,
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id),
    FOREIGN KEY (provider_id) REFERENCES users(id)
);
```

### Laundry Order Ratings Table

```sql
CREATE TABLE laundry_order_ratings (
    id VARCHAR(36) PRIMARY KEY,
    order_id VARCHAR(36) UNIQUE NOT NULL,
    stars INT,
    review TEXT,
    punctuality INT,
    quality INT,
    packaging INT,
    created_at TIMESTAMP,
    FOREIGN KEY (order_id) REFERENCES laundry_orders(id)
);
```

## Integration Points

1. Pricing Module - For price calculations
2. Wallet Module - For payment processing
3. Ratings Module - For order ratings
4. Messages Module - For notifications
5. Service Providers Module - For laundryman info

## Related Documentation

- See MODULES-OVERVIEW.md for module architecture
- See PRICING-MODULE.md for pricing integration
- See WALLET-MODULE.md for payment handling
