-- Drop order_rejections table
DROP INDEX IF EXISTS idx_order_rejections_timestamp;
DROP INDEX IF EXISTS idx_order_rejections_provider;
DROP INDEX IF EXISTS idx_order_provider_rejection;
DROP TABLE IF EXISTS order_rejections;
