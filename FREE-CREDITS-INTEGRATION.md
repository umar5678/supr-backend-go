# Free Credits Integration for Ride Creation

## Problem
When users had free ride credits, the system was still rejecting ride creation with "Insufficient wallet balance" error, even though the free credits could cover the entire fare or part of it.

## Root Cause
The `CreateRide` method was trying to hold the full fare amount in the wallet, without first checking if the user had free ride credits that could offset the amount needed.

## Solution
Implemented intelligent fare hold logic that:
1. Checks available free ride credits
2. Calculates the amount to hold after applying free credits
3. Only holds funds if there's a remaining balance after credits
4. Allows rides to proceed if free credits cover the entire fare

## Changes Made

### 1. Updated WalletResponse DTO
**File:** `internal/modules/wallet/dto/response.go`

Added `FreeRideCredits` field to expose free credits to the API:
```go
type WalletResponse struct {
    ID               string                `json:"id"`
    UserID           string                `json:"userId"`
    WalletType       models.WalletType     `json:"walletType"`
    Balance          float64               `json:"balance"`
    HeldBalance      float64               `json:"heldBalance"`
    AvailableBalance float64               `json:"availableBalance"`
    FreeRideCredits  float64               `json:"freeRideCredits"`  // ← NEW
    Currency         string                `json:"currency"`
    IsActive         bool                  `json:"isActive"`
    CreatedAt        time.Time             `json:"createdAt"`
    UpdatedAt        time.Time             `json:"updatedAt"`
    User             *authdto.UserResponse `json:"user,omitempty"`
}
```

### 2. Updated ToWalletResponse Converter
**File:** `internal/modules/wallet/dto/response.go`

Updated the converter to include FreeRideCredits:
```go
func ToWalletResponse(wallet *models.Wallet) *WalletResponse {
    resp := &WalletResponse{
        // ... existing fields
        FreeRideCredits:  wallet.FreeRideCredits,  // ← NEW
        // ... rest of fields
    }
    return resp
}
```

### 3. Enhanced CreateRide Method
**File:** `internal/modules/rides/service.go` (lines 223-262)

Added free credits logic before holding funds:

```go
// ✅ Calculate amount to hold (after applying free credits)
walletInfo, err := s.walletService.GetWallet(ctx, riderID)
if err != nil {
    logger.Warn("failed to get wallet info", "error", err, "riderID", riderID)
}

amountToHold := finalAmount
if walletInfo != nil && walletInfo.FreeRideCredits > 0 {
    // Free ride credits can cover part or all of the fare
    if walletInfo.FreeRideCredits >= finalAmount {
        amountToHold = 0 // Free credits cover everything
        logger.Info("ride fully covered by free credits", 
            "riderID", riderID, 
            "fareAmount", finalAmount, 
            "freeCredits", walletInfo.FreeRideCredits)
    } else {
        // Free credits cover partial amount, hold the remaining
        amountToHold = finalAmount - walletInfo.FreeRideCredits
        logger.Info("ride partially covered by free credits", 
            "riderID", riderID, 
            "fareAmount", finalAmount, 
            "freeCredits", walletInfo.FreeRideCredits, 
            "amountToHold", amountToHold)
    }
}

// 2. Hold funds with ReferenceID (only if amount > 0)
var holdID *string
if amountToHold > 0 {
    holdReq := walletdto.HoldFundsRequest{
        Amount:        amountToHold,
        ReferenceType: "ride",
        ReferenceID:   rideID,
        HoldDuration:  1800, // 30 minutes
    }

    holdResp, err := s.walletService.HoldFunds(ctx, riderID, holdReq)
    if err != nil {
        return nil, response.BadRequest("Insufficient wallet balance. Please add funds.")
    }
    holdID = &holdResp.ID
    logger.Info("funds held for ride", 
        "rideID", rideID, 
        "holdID", holdResp.ID, 
        "amount", amountToHold)
} else {
    logger.Info("no funds to hold - free credits cover entire fare", "rideID", rideID)
}
```

### 4. Fixed Nil Pointer Handling
Updated hold release logic to safely handle cases where holdID is nil:

```go
// In error handling - only release hold if one was created
if holdID != nil {
    s.walletService.ReleaseHold(ctx, riderID, walletdto.ReleaseHoldRequest{HoldID: *holdID})
}

// In async ride finder cancellation
if holdID != nil {
    if err := s.walletService.ReleaseHold(bgCtx, riderID, walletdto.ReleaseHoldRequest{HoldID: *holdID}); err != nil {
        logger.Error("failed to release hold", "error", err, "rideID", rideID)
    }
}
```

## Test Scenarios

### Scenario 1: User with Sufficient Free Credits (Entire Fare Covered)
```
Fare: $10.00
Free Credits: $15.00
Amount to Hold: $0.00 ✅
Result: Ride created successfully, no wallet hold needed
```

### Scenario 2: User with Partial Free Credits
```
Fare: $10.00
Free Credits: $4.00
Amount to Hold: $6.00 ✅
Result: Only $6 is held from wallet, $4 from free credits
```

### Scenario 3: User with No Free Credits
```
Fare: $10.00
Free Credits: $0.00
Amount to Hold: $10.00 ✅
Result: Standard behavior - full amount held
```

### Scenario 4: User with Free Credits but Insufficient Wallet Balance
```
Fare: $10.00
Free Credits: $3.00
Amount to Hold: $7.00
Wallet Balance: $2.00
Result: ❌ Error - "Insufficient wallet balance"
```

## Logging Examples

### Fully Covered by Free Credits
```
{"level":"info","msg":"ride fully covered by free credits","riderID":"user-123","fareAmount":10.00,"freeCredits":15.00}
```

### Partially Covered
```
{"level":"info","msg":"ride partially covered by free credits","riderID":"user-123","fareAmount":10.00,"freeCredits":4.00,"amountToHold":6.00}
```

### Funds Held
```
{"level":"info","msg":"funds held for ride","rideID":"ride-456","holdID":"hold-789","amount":6.00}
```

### No Hold Needed
```
{"level":"info","msg":"no funds to hold - free credits cover entire fare","rideID":"ride-456"}
```

## API Changes

### GET /api/v1/wallet
The wallet endpoint now returns free credits:

**Response:**
```json
{
  "id": "wallet-123",
  "userId": "user-456",
  "walletType": "rider",
  "balance": 50.00,
  "heldBalance": 10.00,
  "availableBalance": 40.00,
  "freeRideCredits": 25.00,
  "currency": "USD",
  "isActive": true,
  "createdAt": "2024-01-15T10:00:00Z",
  "updatedAt": "2024-01-15T10:00:00Z"
}
```

## Billing Impact

When a ride is completed and payment is collected:

1. **If fully covered by free credits:**
   - Free credits deducted
   - No wallet charge
   - No transaction record needed (optional - can track for analytics)

2. **If partially covered:**
   - Free credits fully deducted
   - Remaining amount charged from wallet/held funds
   - Single transaction record for the balance

3. **If no free credits used:**
   - Full amount charged from held funds
   - Standard transaction record

## Benefits

✅ Users with promotional free credits can now book rides immediately  
✅ Seamless experience - no need to "use" credits explicitly  
✅ Reduces support tickets about "insufficient balance despite having credits"  
✅ Encourages free credit redemption  
✅ Accurate logging of credit usage for analytics  

## Future Enhancements

1. **Partial Credit Redemption Options**: Let users choose if they want to use credits or pay with wallet
2. **Credit Expiry Tracking**: Track when credits expire and warn users
3. **Credit Breakdown in Receipt**: Show how much was paid from credits vs wallet
4. **Analytics Dashboard**: Track credit usage patterns by promotion/user segment
5. **Automatic Credit Application**: Auto-apply highest-value credits first (already done via credits design)

## Build Status
✅ **Success** - All code compiles with no errors
