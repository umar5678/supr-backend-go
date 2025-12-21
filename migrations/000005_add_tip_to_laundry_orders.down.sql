-- Revert: Remove tip column from laundry_orders table

DROP INDEX IF EXISTS idx_laundry_orders_has_tip CASCADE;

ALTER TABLE IF EXISTS laundry_orders DROP COLUMN IF EXISTS tip;
