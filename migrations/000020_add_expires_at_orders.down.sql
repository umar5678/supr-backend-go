-- Remove expires_at column from laundry_orders table
DROP INDEX IF EXISTS idx_laundry_orders_expires_at;
DROP INDEX IF EXISTS idx_laundry_orders_status_expires;

ALTER TABLE laundry_orders
DROP COLUMN IF EXISTS expires_at;
