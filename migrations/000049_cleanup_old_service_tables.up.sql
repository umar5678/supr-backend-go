-- Cleanup: Drop old service-related tables that were created by previous messy migrations
-- This migration cleans up before applying the new clean schema

-- Drop any leftover functions/triggers FIRST (before dropping tables that depend on them)
DROP FUNCTION IF EXISTS generate_order_number() CASCADE;
DROP FUNCTION IF EXISTS update_service_orders_updated_at() CASCADE;

-- Now drop tables
DROP TABLE IF EXISTS provider_qualified_services CASCADE;
DROP TABLE IF EXISTS provider_service_categories CASCADE;
DROP TABLE IF EXISTS service_orders CASCADE;
DROP TABLE IF EXISTS order_status_history CASCADE;
DROP TABLE IF EXISTS order_status_histories CASCADE;
DROP TABLE IF EXISTS addons CASCADE;
DROP TABLE IF EXISTS services CASCADE;
DROP TABLE IF EXISTS service_provider_profiles CASCADE;
DROP TABLE IF EXISTS service_categories CASCADE;
DROP TABLE IF EXISTS service_providers CASCADE;
