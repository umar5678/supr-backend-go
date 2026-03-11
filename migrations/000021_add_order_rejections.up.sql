-- Create order_rejections table to track provider rejections
CREATE TABLE IF NOT EXISTS order_rejections (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL REFERENCES service_orders(id) ON DELETE CASCADE,
    provider_id UUID NOT NULL,
    reason TEXT,
    rejected_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create composite unique index on order_id + provider_id to prevent duplicate rejections
CREATE UNIQUE INDEX IF NOT EXISTS idx_order_provider_rejection ON order_rejections(order_id, provider_id);

-- Create index for efficient lookups of rejections by provider and order
CREATE INDEX IF NOT EXISTS idx_order_rejections_provider ON order_rejections(provider_id, order_id);

-- Create index for finding recent rejections
CREATE INDEX IF NOT EXISTS idx_order_rejections_timestamp ON order_rejections(rejected_at DESC);
