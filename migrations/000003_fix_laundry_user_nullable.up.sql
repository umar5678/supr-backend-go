-- Fix laundry_orders user_id to be nullable and remove foreign key constraint
-- This allows orders to be created without requiring the user to exist in the database

ALTER TABLE IF EXISTS laundry_orders DROP CONSTRAINT IF EXISTS laundry_orders_user_id_fkey CASCADE;

ALTER TABLE IF EXISTS laundry_orders ALTER COLUMN user_id DROP NOT NULL;

-- Note: We do NOT re-add the foreign key constraint
-- This allows testing and edge cases where user_id is stored without user validation
