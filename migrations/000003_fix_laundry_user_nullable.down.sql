-- Revert laundry_orders user_id to NOT NULL

ALTER TABLE IF EXISTS laundry_orders DROP CONSTRAINT IF EXISTS laundry_orders_user_id_fkey CASCADE;

ALTER TABLE IF EXISTS laundry_orders ALTER COLUMN user_id SET NOT NULL;

ALTER TABLE IF EXISTS laundry_orders ADD CONSTRAINT laundry_orders_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
