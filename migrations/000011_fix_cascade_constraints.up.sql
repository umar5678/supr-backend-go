-- Fix foreign key constraints to enable CASCADE delete
-- This allows users and service providers to be deleted with their related data

-- Step 1: Drop the old constraints that don't have ON DELETE CASCADE
ALTER TABLE service_orders DROP CONSTRAINT IF EXISTS fk_service_orders_customer;
ALTER TABLE service_orders DROP CONSTRAINT IF EXISTS fk_service_orders_provider;

-- Step 2: Add new constraints with ON DELETE CASCADE
ALTER TABLE service_orders 
ADD CONSTRAINT fk_service_orders_customer 
FOREIGN KEY (customer_id) REFERENCES users(id) ON DELETE CASCADE;

ALTER TABLE service_orders 
ADD CONSTRAINT fk_service_orders_provider 
FOREIGN KEY (assigned_provider_id) REFERENCES service_provider_profiles(id) ON DELETE CASCADE;

-- Step 3: Check for other missing CASCADE constraints

-- Service orders also references order_status_history which might need CASCADE
ALTER TABLE order_status_history DROP CONSTRAINT IF EXISTS fk_order_status_history_order;
ALTER TABLE order_status_history DROP CONSTRAINT IF EXISTS fk_order_status_history_user;

ALTER TABLE order_status_history 
ADD CONSTRAINT fk_order_status_history_order 
FOREIGN KEY (order_id) REFERENCES service_orders(id) ON DELETE CASCADE;

ALTER TABLE order_status_history 
ADD CONSTRAINT fk_order_status_history_user 
FOREIGN KEY (changed_by) REFERENCES users(id) ON DELETE SET NULL;

-- Step 4: Fix ride foreign keys too (if missing CASCADE)
ALTER TABLE rides DROP CONSTRAINT IF EXISTS fk_rides_promo_code;
ALTER TABLE rides 
ADD CONSTRAINT fk_rides_promo_code 
FOREIGN KEY (promo_code_id) REFERENCES promo_codes(id) ON DELETE SET NULL;

-- Step 5: Fix user_kyc
ALTER TABLE user_kyc DROP CONSTRAINT IF EXISTS fk_user_kyc_user;
ALTER TABLE user_kyc 
ADD CONSTRAINT fk_user_kyc_user 
FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

-- Verify the constraints are now in place
SELECT constraint_name, table_name FROM information_schema.table_constraints 
WHERE table_name IN ('service_orders', 'order_status_history', 'user_kyc', 'rides')
AND constraint_type = 'FOREIGN KEY'
ORDER BY table_name, constraint_name;
