-- Revert laundry_pickups and laundry_deliveries provider_id to NOT NULL

-- Revert laundry_pickups
ALTER TABLE IF EXISTS laundry_pickups DROP CONSTRAINT IF EXISTS laundry_pickups_provider_id_fkey CASCADE;
ALTER TABLE IF EXISTS laundry_pickups ALTER COLUMN provider_id SET NOT NULL;
ALTER TABLE IF EXISTS laundry_pickups ADD CONSTRAINT laundry_pickups_provider_id_fkey FOREIGN KEY (provider_id) REFERENCES service_provider_profiles(id) ON DELETE CASCADE;

-- Revert laundry_deliveries
ALTER TABLE IF EXISTS laundry_deliveries DROP CONSTRAINT IF EXISTS laundry_deliveries_provider_id_fkey CASCADE;
ALTER TABLE IF EXISTS laundry_deliveries ALTER COLUMN provider_id SET NOT NULL;
ALTER TABLE IF EXISTS laundry_deliveries ADD CONSTRAINT laundry_deliveries_provider_id_fkey FOREIGN KEY (provider_id) REFERENCES service_provider_profiles(id) ON DELETE CASCADE;
