-- =====================================================
-- DROP EXTENSIONS FIRST (before tables that depend on them)
-- Note: We don't actually drop extensions to preserve GIS data
-- If you need a complete cleanup, connect as postgres superuser
-- =====================================================

-- DROP EXTENSION IF EXISTS "postgis" CASCADE;
-- DROP EXTENSION IF EXISTS "uuid-ossp" CASCADE;

-- =====================================================
-- DROP TRIGGERS
-- =====================================================

DROP TRIGGER IF EXISTS update_service_orders_updated_at ON service_orders;
DROP TRIGGER IF EXISTS update_provider_service_categories_updated_at ON provider_service_categories;
DROP TRIGGER IF EXISTS update_service_provider_profiles_updated_at ON service_provider_profiles;
DROP TRIGGER IF EXISTS update_addons_updated_at ON addons;
DROP TRIGGER IF EXISTS update_services_updated_at ON services;
DROP TRIGGER IF EXISTS update_surge_pricing_zones_updated_at ON surge_pricing_zones;
DROP TRIGGER IF EXISTS update_ride_requests_updated_at ON ride_requests;
DROP TRIGGER IF EXISTS update_rides_updated_at ON rides;
DROP TRIGGER IF EXISTS update_rider_profiles_updated_at ON rider_profiles;
DROP TRIGGER IF EXISTS update_vehicles_updated_at ON vehicles;
DROP TRIGGER IF EXISTS update_driver_profiles_updated_at ON driver_profiles;
DROP TRIGGER IF EXISTS update_vehicle_types_updated_at ON vehicle_types;
DROP TRIGGER IF EXISTS update_wallets_updated_at ON wallets;
DROP TRIGGER IF EXISTS update_users_updated_at ON users;

DROP FUNCTION IF EXISTS update_updated_at_column();

-- =====================================================
-- DROP TABLES (reverse order of dependencies)
-- =====================================================

DROP TABLE IF EXISTS schema_migrations CASCADE;

-- Home Services
DROP TABLE IF EXISTS order_status_history CASCADE;
DROP TABLE IF EXISTS service_orders CASCADE;
DROP TABLE IF EXISTS provider_qualified_services CASCADE;
DROP TABLE IF EXISTS provider_service_categories CASCADE;
DROP TABLE IF EXISTS service_provider_profiles CASCADE;
DROP TABLE IF EXISTS addons CASCADE;
DROP TABLE IF EXISTS services CASCADE;

-- Ride-Hailing
DROP TABLE IF EXISTS surge_pricing_zones CASCADE;
DROP TABLE IF EXISTS ride_requests CASCADE;
DROP TABLE IF EXISTS rides CASCADE;
DROP TABLE IF EXISTS rider_profiles CASCADE;
DROP TABLE IF EXISTS driver_locations CASCADE;
DROP TABLE IF EXISTS vehicles CASCADE;
DROP TABLE IF EXISTS driver_profiles CASCADE;
DROP TABLE IF EXISTS vehicle_types CASCADE;

-- Wallet
DROP TABLE IF EXISTS wallet_holds CASCADE;
DROP TABLE IF EXISTS wallet_transactions CASCADE;
DROP TABLE IF EXISTS wallets CASCADE;

-- Core
DROP TABLE IF EXISTS users CASCADE;

-- =====================================================
-- DROP ENUM TYPES
-- =====================================================

DROP TYPE IF EXISTS transaction_status;
DROP TYPE IF EXISTS transaction_type;
DROP TYPE IF EXISTS wallet_type;
DROP TYPE IF EXISTS user_status;
DROP TYPE IF EXISTS user_role;
-- DROP EXTENSION IF EXISTS "uuid-ossp";