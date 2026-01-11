# Wallet Type Fix - Migration Guide

## Problem
Existing wallets in the database had `NULL` or empty `wallet_type` values, which caused errors when trying to:
1. Insert new records with proper type constraints
2. Query wallets by type using `FindWalletByUserID(userID, walletType)`

## Error
```
ERROR: invalid input value for enum wallet_type: "" (SQLSTATE 22P02)
```

## Root Cause
The wallet service was creating wallets without setting the `WalletType` field. When multiple users tried to access their wallet, the code would:
1. Try to find wallet by `user_id` AND `wallet_type`
2. Not find it (because wallet_type was NULL/empty)
3. Try to create a new wallet
4. Fail because of enum validation

## Solution Implemented

### 1. Fixed Code (internal/modules/wallet/service.go)
✅ All wallet creation now sets `WalletType`:
- `GetBalance()` - Sets `WalletType: models.WalletTypeRider`
- `HoldFunds()` - Sets `WalletType: models.WalletTypeRider`
- `CreditWallet()` - Sets `WalletType: models.WalletTypeRider`
- `CreditDriverWallet()` - Sets `WalletType: models.WalletTypeDriver`
- `DebitDriverWallet()` - Sets `WalletType: models.WalletTypeDriver`
- `RecordCashCollection()` - Sets `WalletType: models.WalletTypeDriver`

### 2. Fallback Query (internal/modules/wallet/repository.go)
✅ `FindWalletByUserID()` now has fallback logic:
- First tries: `WHERE user_id = ? AND wallet_type = ?`
- If not found, tries: `WHERE user_id = ?` (finds any wallet for that user)
- If found via fallback, automatically updates the wallet_type
- This handles legacy data gracefully

### 3. Data Cleanup Migration (000010_fix_wallet_type)
✅ Removes invalid wallet records:
```sql
-- Delete wallets with NULL wallet_type
DELETE FROM wallet_transactions WHERE wallet_id IN (SELECT id FROM wallets WHERE wallet_type IS NULL);
DELETE FROM wallet_holds WHERE wallet_id IN (SELECT id FROM wallets WHERE wallet_type IS NULL);
DELETE FROM wallets WHERE wallet_type IS NULL;

-- Ensure constraints
ALTER TABLE wallets ALTER COLUMN wallet_type SET NOT NULL;
ALTER TABLE wallets ALTER COLUMN wallet_type SET DEFAULT 'rider'::wallet_type;

-- Prevent duplicates
ALTER TABLE wallets ADD CONSTRAINT unique_user_wallet_type UNIQUE (user_id, wallet_type);

-- Faster queries
CREATE INDEX idx_wallets_user_type ON wallets(user_id, wallet_type);
```

## How to Run the Migration

```bash
migrate -path ./migrations -database 'postgres://go_backend_admin:goPass_Secure123!@localhost:5432/go_backend?sslmode=disable' up
```

## What Gets Deleted?
- All wallets with `wallet_type IS NULL` (broken records)
- All related transactions and holds for those wallets
- These are typically test data or records from before the fix

## After Migration
✅ All new wallet operations will have proper types
✅ Queries will be faster with the new unique constraint and index
✅ No more enum validation errors
✅ Fallback logic ensures backward compatibility with any legacy data

## Testing
```bash
# Try creating a ride (which creates a hold)
POST /api/v1/rides

# Try getting wallet status
GET /api/v1/drivers/wallet/status

# Try topping up wallet
POST /api/v1/drivers/wallet/topup
```

All should work without wallet_type enum errors.
