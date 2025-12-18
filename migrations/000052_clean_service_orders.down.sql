-- Rollback
DROP TRIGGER IF EXISTS trigger_generate_order_number ON service_orders;
DROP TRIGGER IF EXISTS trigger_service_orders_updated_at ON service_orders;
DROP FUNCTION IF EXISTS generate_order_number();
DROP FUNCTION IF EXISTS update_service_orders_updated_at();
DROP TABLE IF EXISTS service_orders CASCADE;
