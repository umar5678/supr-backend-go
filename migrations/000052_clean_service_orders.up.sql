-- Clean Home Services Schema - Part 3: Service Orders (Customer Orders)
-- Orders reference providers via UUID FK (ProviderID), not UserID

CREATE TABLE IF NOT EXISTS service_orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_number VARCHAR(50) NOT NULL UNIQUE,
    
    -- Customer Information
    customer_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    customer_info JSONB NOT NULL,  -- Snapshot: {name, phone, email, address, lat, lng}
    
    -- Booking Information
    booking_info JSONB NOT NULL,  -- {day, date, time, preferredTime, quantityOfPros}
    
    -- Service Details (using slugs)
    category_slug VARCHAR(255) NOT NULL,
    selected_services JSONB NOT NULL,  -- [{serviceSlug, title, price, quantity}, ...]
    selected_addons JSONB,  -- [{addonSlug, title, price, quantity}, ...]
    special_notes TEXT,
    
    -- Pricing
    services_total DECIMAL(10,2) NOT NULL,
    addons_total DECIMAL(10,2) DEFAULT 0,
    subtotal DECIMAL(10,2) NOT NULL,
    platform_commission DECIMAL(10,2) NOT NULL,
    total_price DECIMAL(10,2) NOT NULL,
    
    -- Payment
    payment_info JSONB,
    wallet_hold_id UUID,
    
    -- Provider Assignment (CORRECT FK: references service_provider_profiles.id, not users.id)
    assigned_provider_id UUID REFERENCES service_provider_profiles(id) ON DELETE SET NULL,
    provider_accepted_at TIMESTAMP WITH TIME ZONE,
    provider_started_at TIMESTAMP WITH TIME ZONE,
    provider_completed_at TIMESTAMP WITH TIME ZONE,
    
    -- Order Status
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    
    -- Cancellation
    cancellation_info JSONB,
    
    -- Customer Rating (of Provider)
    customer_rating INTEGER CHECK (customer_rating >= 1 AND customer_rating <= 5),
    customer_review TEXT,
    customer_rated_at TIMESTAMP WITH TIME ZONE,
    
    -- Provider Rating (of Customer)
    provider_rating INTEGER CHECK (provider_rating >= 1 AND provider_rating <= 5),
    provider_review TEXT,
    provider_rated_at TIMESTAMP WITH TIME ZONE,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE
);

-- Indexes
CREATE INDEX idx_service_orders_customer_id ON service_orders(customer_id);
CREATE INDEX idx_service_orders_provider_id ON service_orders(assigned_provider_id);
CREATE INDEX idx_service_orders_status ON service_orders(status);
CREATE INDEX idx_service_orders_order_number ON service_orders(order_number);
CREATE INDEX idx_service_orders_category_slug ON service_orders(category_slug);
CREATE INDEX idx_service_orders_created_at ON service_orders(created_at DESC);
CREATE INDEX idx_service_orders_provider_status ON service_orders(assigned_provider_id, status) WHERE assigned_provider_id IS NOT NULL;
CREATE INDEX idx_service_orders_pending_category ON service_orders(category_slug, created_at) WHERE status IN ('pending', 'searching_provider');

-- Trigger for updated_at
CREATE OR REPLACE FUNCTION update_service_orders_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_service_orders_updated_at ON service_orders;
CREATE TRIGGER trigger_service_orders_updated_at
    BEFORE UPDATE ON service_orders
    FOR EACH ROW
    EXECUTE FUNCTION update_service_orders_updated_at();

-- Trigger for order_number generation
CREATE OR REPLACE FUNCTION generate_order_number()
RETURNS TRIGGER AS $$
DECLARE
    year_part VARCHAR(4);
    sequence_num INTEGER;
BEGIN
    year_part := TO_CHAR(CURRENT_DATE, 'YYYY');
    
    SELECT COALESCE(MAX(
        CAST(SUBSTRING(order_number FROM 'HS-' || year_part || '-(\d+)') AS INTEGER)
    ), 0) + 1
    INTO sequence_num
    FROM service_orders
    WHERE order_number LIKE 'HS-' || year_part || '-%';
    
    NEW.order_number := 'HS-' || year_part || '-' || LPAD(sequence_num::TEXT, 6, '0');
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_generate_order_number ON service_orders;
CREATE TRIGGER trigger_generate_order_number
    BEFORE INSERT ON service_orders
    FOR EACH ROW
    WHEN (NEW.order_number IS NULL OR NEW.order_number = '')
    EXECUTE FUNCTION generate_order_number();
