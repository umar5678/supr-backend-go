-- Delete account and all related data for phone number: +923701653058

-- Step 0: Delete dependent records first (due to missing CASCADE constraints)
-- Delete service orders first since they reference the user
DELETE FROM order_status_history 
WHERE order_id IN (
    SELECT id FROM service_orders WHERE customer_id = (SELECT id FROM users WHERE phone = '+923701653058')
);

DELETE FROM service_orders 
WHERE customer_id = (SELECT id FROM users WHERE phone = '+923701653058');

-- Step 1: Find the user ID first (to see what will be deleted)
SELECT id, name, email, phone, role FROM users WHERE phone = '+923701653058';

-- Step 2: Delete the user and cascade delete all related data
-- This will delete:
-- - User account
-- - All wallets (rider and driver)
-- - All wallet transactions and holds
-- - All rides (as rider or driver)
-- - All ride requests
-- - Driver profile (if exists)
-- - Rider profile (if exists)
-- - Service provider profile (if exists)
-- - Any fraud patterns associated
-- - Any SOS alerts
-- - Any KYC records
-- - Any saved locations
-- - Any promo code usage
-- NOTE: CASCADE happens automatically due to ON DELETE CASCADE in foreign key constraints
DELETE FROM users 
WHERE phone = '+923701653058';

-- Verify deletion
SELECT COUNT(*) as remaining_users FROM users WHERE phone = '+923701653058';
