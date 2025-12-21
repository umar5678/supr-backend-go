-- Fix laundry_pickups and laundry_deliveries provider_id to be nullable
-- This migration removes the NOT NULL constraint from provider_id in both tables

-- Fix laundry_pickups
ALTER TABLE IF EXISTS laundry_pickups DROP CONSTRAINT IF EXISTS laundry_pickups_provider_id_fkey CASCADE;
ALTER TABLE IF EXISTS laundry_pickups ALTER COLUMN provider_id DROP NOT NULL;
ALTER TABLE IF EXISTS laundry_pickups ADD CONSTRAINT laundry_pickups_provider_id_fkey FOREIGN KEY (provider_id) REFERENCES service_provider_profiles(id) ON DELETE CASCADE;

-- Fix laundry_deliveries
ALTER TABLE IF EXISTS laundry_deliveries DROP CONSTRAINT IF EXISTS laundry_deliveries_provider_id_fkey CASCADE;
ALTER TABLE IF EXISTS laundry_deliveries ALTER COLUMN provider_id DROP NOT NULL;
ALTER TABLE IF EXISTS laundry_deliveries ADD CONSTRAINT laundry_deliveries_provider_id_fkey FOREIGN KEY (provider_id) REFERENCES service_provider_profiles(id) ON DELETE CASCADE;
