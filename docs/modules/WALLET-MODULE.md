# Wallet Module Development Guide

## Overview

The Wallet Module manages user wallets, payment transactions, balance tracking, and financial operations. It handles fund additions, withdrawals, ride payments, and transaction history.

## Module Structure

```
wallet/
├── handler.go         # HTTP request handlers
├── service.go         # Business logic
├── repository.go      # Database operations
├── routes.go          # Route definitions
└── dto/
    ├── requests.go    # Request payloads
    └── responses.go   # Response structures
```

## Key Responsibilities

1. Wallet Management - Create and manage user wallets
2. Balance Tracking - Maintain accurate balance records
3. Transaction Processing - Handle financial transactions
4. Funds Management - Add and withdraw funds
5. Payment Processing - Process ride and service payments
6. Transaction History - Maintain transaction records
7. Refund Handling - Process and track refunds

## Architecture

### Handler Layer (handler.go)

Manages HTTP endpoints for wallet operations.

Key methods:

```
GetWallet(c *gin.Context)               // GET /wallet
GetBalance(c *gin.Context)              // GET /wallet/balance
AddFunds(c *gin.Context)                // POST /wallet/add-funds
WithdrawFunds(c *gin.Context)           // POST /wallet/withdraw
GetTransactionHistory(c *gin.Context)   // GET /wallet/transactions
ProcessPayment(c *gin.Context)          // POST /wallet/payment
RefundTransaction(c *gin.Context)       // POST /wallet/refund/{id}
GetWalletStatement(c *gin.Context)      // GET /wallet/statement
TransferFunds(c *gin.Context)           // POST /wallet/transfer
```

Request flow:
1. Extract parameters from URL/Query/Body
2. Validate request data
3. Ensure user owns the wallet
4. Call service method
5. Return response with updated balance

### Service Layer (service.go)

Contains wallet business logic.

Key interface methods:

```
GetWallet(ctx context.Context, userID string) (*WalletResponse, error)
GetBalance(ctx context.Context, userID string) (*WalletBalanceResponse, error)
AddFunds(ctx context.Context, userID string, req AddFundsRequest) (*TransactionResponse, error)
WithdrawFunds(ctx context.Context, userID string, req WithdrawFundsRequest) (*TransactionResponse, error)
GetTransactionHistory(ctx context.Context, userID string, filters map[string]interface{}) ([]*TransactionResponse, error)
ProcessPayment(ctx context.Context, userID, rideID string, amount float64) (*TransactionResponse, error)
RefundTransaction(ctx context.Context, transactionID string) (*TransactionResponse, error)
TransferFunds(ctx context.Context, fromUserID, toUserID string, amount float64) error
SettleEarnings(ctx context.Context, driverID string) error
```

Logic flow:
1. Validate wallet exists and is active
2. Verify sufficient balance for withdrawals/payments
3. Record transaction with timestamp
4. Update wallet balance atomically
5. Trigger payment gateway if needed
6. Create transaction record
7. Log all financial operations

### Repository Layer (repository.go)

Handles database operations for wallets.

Key interface methods:

```
CreateWallet(ctx context.Context, userID, userType string) error
FindWalletByUserID(ctx context.Context, userID string) (*models.Wallet, error)
UpdateBalance(ctx context.Context, walletID string, amount float64) error
CreateTransaction(ctx context.Context, transaction *models.Transaction) error
FindTransactionByID(ctx context.Context, transactionID string) (*models.Transaction, error)
ListTransactions(ctx context.Context, walletID string, filters map[string]interface{}) ([]*models.Transaction, error)
GetBalance(ctx context.Context, userID string) (float64, error)
UpdateTransactionStatus(ctx context.Context, transactionID, status string) error
```

Database operations:
- Use transactions for balance updates
- Implement proper locking for concurrent operations
- Store complete transaction audit trail
- Calculate running balance

## Data Transfer Objects

### WalletResponse

```go
type WalletResponse struct {
    ID              string    `json:"id"`
    UserID          string    `json:"user_id"`
    UserType        string    `json:"user_type"` // rider, driver, provider
    CurrentBalance  float64   `json:"current_balance"`
    TotalEarnings   float64   `json:"total_earnings,omitempty"` // for drivers
    TotalSpent      float64   `json:"total_spent"`
    Status          string    `json:"status"`
    CreatedAt       time.Time `json:"created_at"`
    UpdatedAt       time.Time `json:"updated_at"`
}
```

### WalletBalanceResponse

```go
type WalletBalanceResponse struct {
    Balance         float64      `json:"balance"`
    Currency        string       `json:"currency"`
    LastUpdated     time.Time    `json:"last_updated"`
}
```

### AddFundsRequest

```go
type AddFundsRequest struct {
    Amount          float64 `json:"amount" binding:"required,gt=0"`
    PaymentMethod   string  `json:"payment_method" binding:"required"` // credit_card, debit_card, bank_transfer
    PaymentDetails  PaymentDetails `json:"payment_details"`
    Remarks         string  `json:"remarks,omitempty"`
}

type PaymentDetails struct {
    CardToken       string  `json:"card_token,omitempty"` // Encrypted token
    BankAccount     string  `json:"bank_account,omitempty"`
    TransactionRef  string  `json:"transaction_ref,omitempty"`
}
```

### TransactionResponse

```go
type TransactionResponse struct {
    ID              string    `json:"id"`
    WalletID        string    `json:"wallet_id"`
    Type            string    `json:"type"` // debit, credit, payment, refund, transfer
    Amount          float64   `json:"amount"`
    Description     string    `json:"description"`
    Status          string    `json:"status"` // pending, completed, failed, cancelled
    ReferenceID     string    `json:"reference_id,omitempty"` // ride_id, order_id, etc
    BalanceBefore   float64   `json:"balance_before"`
    BalanceAfter    float64   `json:"balance_after"`
    CreatedAt       time.Time `json:"created_at"`
    CompletedAt     *time.Time `json:"completed_at,omitempty"`
}
```

### WithdrawFundsRequest

```go
type WithdrawFundsRequest struct {
    Amount          float64 `json:"amount" binding:"required,gt=0"`
    BankAccount     string  `json:"bank_account" binding:"required"`
    AccountName     string  `json:"account_name"`
    Remarks         string  `json:"remarks,omitempty"`
}
```

## Wallet Types and Rules

### Rider Wallet

```
Purpose: Prepaid balance for ride payments
Top-up Methods: Credit card, debit card, bank transfer
Withdrawal: Not allowed
Expiry: No expiry for balance
Minimum Balance: 0
```

### Driver Wallet

```
Purpose: Earnings tracking and cash management
Top-up Methods: Not applicable (auto-funded by earnings)
Withdrawal: To bank account
Expiry: No expiry
Minimum Balance: 0
Cash Tracking: For cash-based drivers
Settlement Frequency: Daily or weekly
```

### Service Provider Wallet

```
Purpose: Earnings and payment management
Top-up Methods: Bank transfer
Withdrawal: To bank account
Expiry: No expiry
Settlement Frequency: Bi-weekly or monthly
```

## Transaction Types

```
CREDIT:
- Refund - Ride cancellation refund
- Bonus - Referral bonus, promotional credit
- Adjustment - Admin adjustment
- Transfer - Received from another user

DEBIT:
- Ride - Ride payment
- Service - Service payment
- Withdrawal - Cash withdrawal
- Transfer - Sent to another user
- Fee - Processing fee
- Cancellation - Cancellation charge

PAYMENT:
- Completed - Successfully charged
- Pending - Awaiting confirmation
- Failed - Payment failed
- Cancelled - User cancelled
```

## Typical Use Cases

### 1. Get Wallet Balance

Request:
```
GET /wallet/balance
```

Response:
```json
{
    "balance": 250.50,
    "currency": "USD",
    "last_updated": "2024-02-20T10:30:00Z"
}
```

Flow:
1. Extract user ID from JWT token
2. Find wallet by user ID
3. Return current balance

### 2. Add Funds to Wallet

Request:
```
POST /wallet/add-funds
{
    "amount": 100.00,
    "payment_method": "credit_card",
    "payment_details": {
        "card_token": "tok_visa_123456"
    }
}
```

Flow:
1. Validate amount (min: 1, max: 5000)
2. Create payment intent with payment gateway
3. Process payment and get transaction ID
4. Create pending transaction record
5. Wait for payment confirmation webhook
6. Update wallet balance once confirmed
7. Return transaction details

### 3. Process Ride Payment

Request:
```
POST /wallet/payment
{
    "ride_id": "ride-123",
    "amount": 25.50
}
```

Flow:
1. Find ride by ID
2. Verify ride is completed
3. Check wallet has sufficient balance
4. Deduct amount from balance (atomic)
5. Create debit transaction
6. Create credit transaction for driver/provider
7. Return transaction confirmation

### 4. Get Transaction History

Request:
```
GET /wallet/transactions?type=debit&status=completed&limit=20
```

Response:
```json
{
    "transactions": [
        {
            "id": "txn-123",
            "type": "debit",
            "amount": 25.50,
            "description": "Ride payment - Trip ID: ride-123",
            "status": "completed",
            "reference_id": "ride-123",
            "balance_before": 300.00,
            "balance_after": 274.50,
            "created_at": "2024-02-20T10:30:00Z"
        }
    ],
    "total": 1,
    "page": 1
}
```

Flow:
1. Find wallet by user ID
2. Query transactions with filters
3. Order by creation date descending
4. Apply pagination
5. Return transaction list

### 5. Refund Transaction

Request:
```
POST /wallet/refund/txn-123
{
    "reason": "Ride cancelled by driver"
}
```

Flow:
1. Find original transaction
2. Verify transaction can be refunded (status, time window)
3. Calculate refund amount
4. Check for duplicates
5. Create refund transaction
6. Update wallet balance
7. Notify user
8. Return refund details

### 6. Withdraw Funds

Request:
```
POST /wallet/withdraw
{
    "amount": 500.00,
    "bank_account": "1234567890",
    "account_name": "John Doe"
}
```

Flow:
1. Verify user is a driver or service provider
2. Validate amount (min: 100, max: current balance)
3. Verify bank account is registered
4. Create pending withdrawal transaction
5. Queue for batch settlement
6. Return withdrawal details
7. Process actual bank transfer in batch job

### 7. Driver Settlement

Automated daily settlement:

Flow:
1. Find all drivers with pending earnings
2. Get earnings from completed rides
3. Deduct platform commission (15%)
4. Deduct payment processing fees
5. Add balance to driver wallet
6. Create settlement transaction
7. Queue for bank transfer
8. Send notification to driver

## Payment Gateway Integration

The module integrates with payment gateways for external payments:

```
Card Payments:
- Stripe, Razorpay, or similar gateway
- Store encrypted card tokens
- Tokenized payments for subsequent transactions

Bank Transfers:
- NEFT, RTGS for India
- ACH for US
- SWIFT for international
- Batch processing daily/weekly

Payment Flow:
1. Create payment intent with gateway
2. Handle 3D secure if required
3. Webhook verification
4. Update transaction status
5. Handle failures and retries
```

## Error Handling

Common error scenarios:

1. Insufficient Balance
   - Response: 400 Bad Request
   - Message: "Insufficient balance. Available: 50, Required: 100"

2. Wallet Not Found
   - Response: 404 Not Found
   - Action: Create wallet for new users

3. Payment Failed
   - Response: 400 Bad Request
   - Message: "Payment declined by card issuer"
   - Action: Retry with different method

4. Transaction Already Refunded
   - Response: 409 Conflict
   - Message: "This transaction has already been refunded"

5. Concurrent Balance Updates
   - Response: 409 Conflict
   - Action: Use database locking for atomic updates

## Testing Strategy

### Unit Tests (Service Layer)

```go
Test_AddFunds_Success()
Test_AddFunds_InsufficientAmount()
Test_ProcessPayment_Success()
Test_ProcessPayment_InsufficientBalance()
Test_RefundTransaction_Success()
Test_WithdrawFunds_ToBank()
Test_TransferFunds_BetweenUsers()
Test_SettleEarnings_Daily()
```

### Integration Tests (Repository Layer)

```go
Test_CreateWallet()
Test_UpdateBalance_Atomic()
Test_CreateTransaction()
Test_ListTransactions_WithFilters()
```

### End-to-End Tests (Handler Layer)

```go
Test_AddFunds_FullFlow()
Test_RidePayment_FullFlow()
Test_Withdrawal_FullFlow()
```

## Database Schema

### Wallets Table

```sql
CREATE TABLE wallets (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) UNIQUE NOT NULL,
    user_type VARCHAR(50),
    balance DECIMAL(15, 2),
    total_earned DECIMAL(15, 2),
    total_spent DECIMAL(15, 2),
    status VARCHAR(50),
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id)
);
```

### Transactions Table

```sql
CREATE TABLE transactions (
    id VARCHAR(36) PRIMARY KEY,
    wallet_id VARCHAR(36) NOT NULL,
    type VARCHAR(50),
    amount DECIMAL(15, 2),
    description TEXT,
    status VARCHAR(50),
    reference_id VARCHAR(36),
    balance_before DECIMAL(15, 2),
    balance_after DECIMAL(15, 2),
    metadata JSON,
    created_at TIMESTAMP,
    completed_at TIMESTAMP,
    FOREIGN KEY (wallet_id) REFERENCES wallets(id),
    INDEX (wallet_id, created_at),
    INDEX (reference_id),
    INDEX (status)
);
```

### Bank Accounts Table

```sql
CREATE TABLE bank_accounts (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    account_number VARCHAR(100),
    account_holder VARCHAR(255),
    bank_name VARCHAR(255),
    ifsc_code VARCHAR(20),
    is_verified BOOLEAN,
    verified_at TIMESTAMP,
    created_at TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id)
);
```

## Performance Optimization

1. Index on user_id for quick balance lookup
2. Use database transactions for atomic balance updates
3. Cache wallet balance in Redis
4. Batch settlement processes
5. Use async processing for payment webhooks
6. Implement connection pooling

## Security Considerations

1. Encryption - Store sensitive payment data encrypted
2. PCI Compliance - Never store full card numbers
3. Tokenization - Use payment gateway tokens
4. Authentication - Verify user identity
5. Authorization - Only allow users to access own wallets
6. Audit Logging - Log all financial transactions
7. Fraud Detection - Monitor unusual patterns
8. Rate Limiting - Limit transaction frequency

## Integration Points

1. Rides Module - For ride payment deduction
2. Drivers Module - For driver earnings
3. Service Providers Module - For provider earnings
4. Payments Module - For external payment processing
5. Messages Module - For transaction notifications
6. Laundry Module - For service payments

## Configuration

Typical wallet configuration:

```yaml
wallet:
  minimum_topup: 1.00
  maximum_topup: 5000.00
  minimum_withdrawal: 100.00
  commission_rate: 0.15
  payment_methods:
    - credit_card
    - debit_card
    - bank_transfer
  settlement:
    frequency: daily
    batch_time: "02:00"
  refund_window: 7d
```

## Related Documentation

- See MODULES-OVERVIEW.md for module architecture
- See RIDES-MODULE.md for payment integration
- See DRIVERS-MODULE.md for earnings tracking
- See internal/utils/response for error handling

## Common Pitfalls

1. Race conditions in balance updates
2. Not using transactions for atomic operations
3. Storing sensitive payment data
4. Not implementing proper refund logic
5. Insufficient transaction logging
6. No idempotency for payment operations
7. Missing reconciliation with payment gateway
8. Not handling payment webhook failures

## Future Enhancements

1. Cryptocurrency wallet support
2. Subscription/auto-replenishment
3. Wallet lending/credit facility
4. Rewards and loyalty points
5. Multi-currency support
6. Instant settlement options
7. Peer-to-peer transfers
8. Bill payment integration
9. Donation feature
10. Insurance premium deduction
