-- Rollback Laundry Service Schema

-- Drop views first
DROP VIEW IF EXISTS v_laundry_services_with_products;

-- Drop triggers
DROP TRIGGER IF EXISTS trigger_update_laundry_issues_updated_at ON laundry_issues;
DROP TRIGGER IF EXISTS trigger_update_laundry_deliveries_updated_at ON laundry_deliveries;
DROP TRIGGER IF EXISTS trigger_update_laundry_pickups_updated_at ON laundry_pickups;
DROP TRIGGER IF EXISTS trigger_update_laundry_order_items_updated_at ON laundry_order_items;
DROP TRIGGER IF EXISTS trigger_update_laundry_orders_updated_at ON laundry_orders;
DROP TRIGGER IF EXISTS trigger_update_laundry_service_products_updated_at ON laundry_service_products;
DROP TRIGGER IF EXISTS trigger_update_laundry_service_catalog_updated_at ON laundry_service_catalog;

-- Drop tables in correct order (dependent tables first)
DROP TABLE IF EXISTS laundry_issues CASCADE;
DROP TABLE IF EXISTS laundry_deliveries CASCADE;
DROP TABLE IF EXISTS laundry_pickups CASCADE;
DROP TABLE IF EXISTS laundry_order_items CASCADE;
DROP TABLE IF EXISTS laundry_orders CASCADE;
DROP TABLE IF EXISTS laundry_service_products CASCADE;
DROP TABLE IF EXISTS laundry_service_catalog CASCADE;

-- Drop functions
DROP FUNCTION IF EXISTS update_laundry_table_updated_at();
DROP FUNCTION IF EXISTS update_laundry_service_catalog_updated_at();
DROP FUNCTION IF EXISTS get_nearest_facility(DECIMAL, DECIMAL, INTEGER);
