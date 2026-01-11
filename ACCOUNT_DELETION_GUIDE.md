# Account Deletion Guide - Phone: +923701653058

## Overview
This guide explains how to safely delete a user account with phone number `+923701653058` and all associated data.

## What Gets Deleted
When you delete a user, the following data is automatically deleted due to CASCADE constraints:

### Direct User Data
- ✅ User account (users table)
- ✅ User profile photo
- ✅ Emergency contact info
- ✅ KYC records (user_kyc table)
- ✅ Saved locations (saved_locations table)

### Wallet Data
- ✅ All wallets (rider and driver types)
- ✅ All wallet transactions
- ✅ All wallet holds
- ✅ Free ride credits

### Ride Data (as Rider)
- ✅ All ride requests created by user
- ✅ All rides where user was the rider
- ✅ Associated ratings and reviews
- ✅ Wait time charges
- ✅ Promo code usage

### Ride Data (as Driver)
- ✅ Driver profile
- ✅ All rides where user was the driver
- ✅ Vehicle records
- ✅ Driver locations history
- ✅ All ride requests sent to this driver

### Service Data
- ✅ Service provider profile (if exists)
- ✅ Service orders (if customer)
- ✅ Provider service categories
- ✅ Qualified services

### Safety Data
- ✅ SOS alerts
- ✅ Fraud patterns
- ✅ Fraud patterns where they were related user

## Deletion Methods

### Method 1: View What Will Be Deleted (SAFE - Read Only)
```bash
psql -h localhost -U go_backend_admin -d go_backend -f DELETE_ACCOUNT_SAFE.sql
```
This will show you counts of what will be deleted without actually deleting anything.

### Method 2: Simple Delete (PRODUCTION)
```bash
psql -h localhost -U go_backend_admin -d go_backend -f DELETE_ACCOUNT_PRODUCTION.sql
```
This will:
1. Find the user by phone
2. Delete the user (all related data deleted via CASCADE)
3. Show confirmation message

### Method 3: Manual Delete via psql
```bash
psql -h localhost -U go_backend_admin -d go_backend

-- View user first
SELECT id, name, email, phone FROM users WHERE phone = '+923701653058';

-- Delete the user
DELETE FROM users WHERE phone = '+923701653058';

-- Verify deletion
SELECT COUNT(*) FROM users WHERE phone = '+923701653058';
```

## Database Connection String
```
postgres://go_backend_admin:goPass_Secure123!@localhost:5432/go_backend?sslmode=disable
```

## Safety Precautions

### ⚠️ BEFORE DELETING:
1. **BACKUP YOUR DATABASE** - This action cannot be undone!
   ```bash
   pg_dump -h localhost -U go_backend_admin go_backend > backup_$(date +%Y%m%d_%H%M%S).sql
   ```

2. **Verify the correct phone number** - Check user details:
   ```sql
   SELECT id, name, email, phone, role, created_at FROM users WHERE phone = '+923701653058';
   ```

3. **Check for active rides** - Make sure user has no ongoing rides:
   ```sql
   SELECT * FROM rides 
   WHERE (rider_id = (SELECT id FROM users WHERE phone = '+923701653058') OR 
          driver_id = (SELECT id FROM users WHERE phone = '+923701653058'))
   AND status IN ('requested', 'accepted', 'started', 'arriving');
   ```

### ✅ AFTER DELETING:
1. Verify the user is gone:
   ```sql
   SELECT COUNT(*) FROM users WHERE phone = '+923701653058';
   ```

2. Check the app - The user should not be able to log in

3. Monitor logs for any orphaned data issues

## Rollback Option

If you need to undo the deletion (before closing the transaction):

```sql
ROLLBACK;
```

Or restore from backup if already committed:
```bash
# List backups
ls -la backup_*.sql

# Restore specific backup
psql -h localhost -U go_backend_admin -d go_backend < backup_20260111_120000.sql
```

## Query Information

| Script | Purpose | Usage |
|--------|---------|-------|
| DELETE_ACCOUNT_QUERY.sql | Quick delete | Simple, direct deletion |
| DELETE_ACCOUNT_SAFE.sql | Preview deletion | See what will be deleted first |
| DELETE_ACCOUNT_PRODUCTION.sql | Safe production delete | With error handling and logging |

## Related Data That References This User

If you need to check what other data references this user:

```sql
-- Find all references
SELECT * FROM wallets WHERE user_id = (SELECT id FROM users WHERE phone = '+923701653058');
SELECT * FROM rides WHERE rider_id = (SELECT id FROM users WHERE phone = '+923701653058');
SELECT * FROM service_orders WHERE customer_id = (SELECT id FROM users WHERE phone = '+923701653058');
SELECT * FROM fraud_patterns WHERE user_id = (SELECT id FROM users WHERE phone = '+923701653058');
SELECT * FROM sos_alerts WHERE user_id = (SELECT id FROM users WHERE phone = '+923701653058');
```

## Support

If you encounter any issues during deletion:
1. Check the error message
2. Ensure your backup is safe
3. Review the constraints in the schema
4. Contact database administrator

---
**Created**: 2026-01-11
**Phone**: +923701653058
**Action**: Delete account and cascade data
