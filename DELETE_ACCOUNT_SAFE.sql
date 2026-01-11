-- SAFE DELETE SCRIPT for account: +923701653058
-- This script shows what will be deleted before actually deleting

-- ============================================
-- STEP 1: Find the user
-- ============================================
SELECT 'FINDING USER' as step;
SELECT id as user_id, name, email, phone, role FROM users WHERE phone = '+923701653058';

-- ============================================
-- STEP 2: Preview what will be deleted
-- ============================================

SELECT 'WALLETS' as item_type, COUNT(*) as count FROM wallets 
WHERE user_id = (SELECT id FROM users WHERE phone = '+923701653058');

SELECT 'WALLET_TRANSACTIONS' as item_type, COUNT(*) as count FROM wallet_transactions 
WHERE wallet_id IN (SELECT id FROM wallets WHERE user_id = (SELECT id FROM users WHERE phone = '+923701653058'));

SELECT 'WALLET_HOLDS' as item_type, COUNT(*) as count FROM wallet_holds 
WHERE wallet_id IN (SELECT id FROM wallets WHERE user_id = (SELECT id FROM users WHERE phone = '+923701653058'));

SELECT 'RIDES_AS_RIDER' as item_type, COUNT(*) as count FROM rides 
WHERE rider_id = (SELECT id FROM users WHERE phone = '+923701653058');

SELECT 'RIDES_AS_DRIVER' as item_type, COUNT(*) as count FROM rides 
WHERE driver_id = (SELECT id FROM users WHERE phone = '+923701653058');

SELECT 'RIDE_REQUESTS' as item_type, COUNT(*) as count FROM ride_requests 
WHERE driver_id IN (SELECT id FROM driver_profiles WHERE user_id = (SELECT id FROM users WHERE phone = '+923701653058'));

SELECT 'SERVICE_ORDERS_AS_CUSTOMER' as item_type, COUNT(*) as count FROM service_orders 
WHERE customer_id = (SELECT id FROM users WHERE phone = '+923701653058');

SELECT 'PROMO_CODE_USAGE' as item_type, COUNT(*) as count FROM promo_code_usage 
WHERE user_id = (SELECT id FROM users WHERE phone = '+923701653058');

SELECT 'SOS_ALERTS' as item_type, COUNT(*) as count FROM sos_alerts 
WHERE user_id = (SELECT id FROM users WHERE phone = '+923701653058');

-- ============================================
-- STEP 3: EXECUTE DELETION (uncomment when ready)
-- ============================================
-- WARNING: This will permanently delete the account and all related data!
-- Backup your database before running this!

-- BEGIN TRANSACTION;

-- DELETE FROM users WHERE phone = '+923701653058';

-- COMMIT;

-- To ROLLBACK if something goes wrong:
-- ROLLBACK;
