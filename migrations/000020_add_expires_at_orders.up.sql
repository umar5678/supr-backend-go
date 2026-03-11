-- Add expires_at column to laundry_orders table
ALTER TABLE laundry_orders
ADD COLUMN IF NOT EXISTS expires_at TIMESTAMP WITH TIME ZONE;

-- Create index for efficient expiration queries
CREATE INDEX IF NOT EXISTS idx_laundry_orders_expires_at ON laundry_orders(expires_at)
WHERE expires_at IS NOT NULL AND status IN ('pending', 'searching_provider');

-- Create index for efficient status + expiry queries
CREATE INDEX IF NOT EXISTS idx_laundry_orders_status_expires ON laundry_orders(status, expires_at);
