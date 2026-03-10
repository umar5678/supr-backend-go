-- =====================================================
-- SIMPLE FORCED DELETION - NO QUESTIONS ASKED
-- =====================================================
-- Execute this to delete all online drivers immediately
-- Backup your database first!

BEGIN;

-- Disable constraints
SET CONSTRAINTS ALL DEFERRED;

-- Step 1: Get all online driver user IDs
CREATE TEMP TABLE online_driver_ids AS
SELECT user_id FROM driver_profiles WHERE status = 'online';

-- Step 2: Delete everything related to online drivers in order
DELETE FROM ride_requests WHERE driver_id IN (SELECT id FROM driver_profiles WHERE status = 'online');
DELETE FROM driver_locations WHERE driver_id IN (SELECT id FROM driver_profiles WHERE status = 'online');
DELETE FROM vehicles WHERE driver_id IN (SELECT id FROM driver_profiles WHERE status = 'online');
DELETE FROM fraud_patterns WHERE driver_id IN (SELECT id FROM driver_profiles WHERE status = 'online');

-- Step 3: Delete rides where these users are either driver or rider
DELETE FROM rides WHERE driver_id IN (SELECT user_id FROM online_driver_ids) OR rider_id IN (SELECT user_id FROM online_driver_ids);

-- Step 4: Delete ride-related records
DELETE FROM sos_alerts WHERE ride_id IN (SELECT id FROM rides WHERE driver_id IN (SELECT user_id FROM online_driver_ids) OR rider_id IN (SELECT user_id FROM online_driver_ids));
DELETE FROM wait_time_charges WHERE ride_id IN (SELECT id FROM rides WHERE driver_id IN (SELECT user_id FROM online_driver_ids) OR rider_id IN (SELECT user_id FROM online_driver_ids));

-- Step 5: Delete wallet-related records for these users
DELETE FROM wallet_holds WHERE wallet_id IN (SELECT id FROM wallets WHERE user_id IN (SELECT user_id FROM online_driver_ids));
DELETE FROM wallet_transactions WHERE wallet_id IN (SELECT id FROM wallets WHERE user_id IN (SELECT user_id FROM online_driver_ids));
DELETE FROM wallets WHERE user_id IN (SELECT user_id FROM online_driver_ids);

-- Step 6: Delete user-related records
DELETE FROM user_kyc WHERE user_id IN (SELECT user_id FROM online_driver_ids);
DELETE FROM saved_locations WHERE user_id IN (SELECT user_id FROM online_driver_ids);
DELETE FROM rider_profiles WHERE user_id IN (SELECT user_id FROM online_driver_ids);
DELETE FROM promo_code_usage WHERE user_id IN (SELECT user_id FROM online_driver_ids);

-- Step 7: Delete users and driver profiles
DELETE FROM users WHERE id IN (SELECT user_id FROM online_driver_ids);
DELETE FROM driver_profiles WHERE status = 'online';

COMMIT;

-- Verify
SELECT COUNT(*) as remaining_online_drivers FROM driver_profiles WHERE status = 'online';


