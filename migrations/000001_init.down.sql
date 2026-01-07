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
DROP TABLE IF EXISTS fraud_patterns CASCADE;

-- create_sos_alerts
DROP TABLE IF EXISTS sos_alerts;

-- Wallet
DROP TABLE IF EXISTS wallet_holds CASCADE;
DROP TABLE IF EXISTS wallet_transactions CASCADE;
DROP TABLE IF EXISTS wallets CASCADE;

-- Core
DROP TABLE IF EXISTS users CASCADE;
DROP TABLE IF EXISTS saved_locations;
DROP TABLE IF EXISTS user_kyc;
DROP TABLE IF EXISTS wait_time_charges;
DROP TABLE IF EXISTS price_capping_rules;
DROP TABLE IF EXISTS promo_code_usage;
DROP TABLE IF EXISTS promo_codes;

-- =====================================================
-- DROP ENUM TYPES
-- =====================================================

DROP TYPE IF EXISTS transaction_status;
DROP TYPE IF EXISTS transaction_type;
DROP TYPE IF EXISTS wallet_type;
DROP TYPE IF EXISTS user_status;
DROP TYPE IF EXISTS user_role;
-- DROP EXTENSION IF EXISTS "uuid-ossp";

-- =====================================================
-- Alterations
-- =====================================================

ALTER TABLE users DROP COLUMN IF EXISTS referral_code;
ALTER TABLE users DROP COLUMN IF EXISTS referred_by;
ALTER TABLE users DROP COLUMN IF EXISTS emergency_contact_name;
ALTER TABLE users DROP COLUMN IF EXISTS emergency_contact_phone;
ALTER TABLE users DROP COLUMN IF EXISTS ride_pin;
ALTER TABLE rides DROP COLUMN IF EXISTS destination_changed;
ALTER TABLE rides DROP COLUMN IF EXISTS destination_change_charge;
ALTER TABLE rides DROP COLUMN IF EXISTS wait_time_charge;
ALTER TABLE rides DROP COLUMN IF EXISTS promo_code_id;
ALTER TABLE rides DROP COLUMN IF EXISTS promo_discount;
ALTER TABLE wallets DROP COLUMN IF EXISTS free_ride_credits;
ALTER TABLE rides DROP COLUMN IF EXISTS rider_rating;
ALTER TABLE rides DROP COLUMN IF EXISTS driver_rating;
ALTER TABLE rides DROP COLUMN IF EXISTS rider_rating_comment;
ALTER TABLE rides DROP COLUMN IF EXISTS driver_rating_comment;
ALTER TABLE rides DROP COLUMN IF EXISTS rider_rated_at;
ALTER TABLE rides DROP COLUMN IF EXISTS driver_rated_at;


ALTER TABLE rides DROP CONSTRAINT IF EXISTS fk_rides_promo_code;
-- =====================================================
-- DropIndexes
-- =====================================================

DROP INDEX IF EXISTS idx_users_ride_pin;