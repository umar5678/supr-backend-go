-- Rollback CASCADE constraint fixes
-- Restore original constraints (without CASCADE)

-- Note: This is just for reference. In production, you probably don't want to rollback these changes
-- as CASCADE delete is safer and prevents orphaned records.

-- If you really need to revert:

ALTER TABLE service_orders DROP CONSTRAINT IF EXISTS fk_service_orders_customer;
ALTER TABLE service_orders DROP CONSTRAINT IF EXISTS fk_service_orders_provider;

ALTER TABLE service_orders 
ADD CONSTRAINT fk_service_orders_customer 
FOREIGN KEY (customer_id) REFERENCES users(id);

ALTER TABLE service_orders 
ADD CONSTRAINT fk_service_orders_provider 
FOREIGN KEY (assigned_provider_id) REFERENCES service_provider_profiles(id);

ALTER TABLE order_status_history DROP CONSTRAINT IF EXISTS fk_order_status_history_order;
ALTER TABLE order_status_history DROP CONSTRAINT IF EXISTS fk_order_status_history_user;

ALTER TABLE order_status_history 
ADD CONSTRAINT fk_order_status_history_order 
FOREIGN KEY (order_id) REFERENCES service_orders(id);

ALTER TABLE order_status_history 
ADD CONSTRAINT fk_order_status_history_user 
FOREIGN KEY (changed_by) REFERENCES users(id);
