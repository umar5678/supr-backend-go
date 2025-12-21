-- Add tip column to laundry_orders table for delivery person gratuity

ALTER TABLE IF EXISTS laundry_orders ADD COLUMN tip DECIMAL(10,2);

-- Create index for filtering orders with tips
CREATE INDEX IF NOT EXISTS idx_laundry_orders_has_tip ON laundry_orders((tip IS NOT NULL));
