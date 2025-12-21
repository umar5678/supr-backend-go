-- Laundry Service Schema Migrations
-- Updated: All products now have explicit prices
-- Note: PostGIS extensions are optional and may require superuser privileges
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- =====================================================
-- TABLE 1: laundry_service_catalog
-- =====================================================
CREATE TABLE IF NOT EXISTS laundry_service_catalog (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slug VARCHAR(100) NOT NULL UNIQUE,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    color_code VARCHAR(20),
    base_price DECIMAL(10,2) NOT NULL,
    pricing_unit VARCHAR(20) NOT NULL DEFAULT 'kg', -- kg, item, load
    turnaround_hours INTEGER DEFAULT 48,
    express_fee DECIMAL(10,2) DEFAULT 0,
    express_hours INTEGER DEFAULT 24,
    display_order INTEGER DEFAULT 0,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

DROP INDEX IF EXISTS idx_laundry_service_catalog_slug CASCADE;
DROP INDEX IF EXISTS idx_laundry_service_catalog_is_active CASCADE;
DROP INDEX IF EXISTS idx_laundry_service_catalog_display_order CASCADE;

CREATE INDEX idx_laundry_service_catalog_slug ON laundry_service_catalog(slug);
CREATE INDEX idx_laundry_service_catalog_is_active ON laundry_service_catalog(is_active);
CREATE INDEX idx_laundry_service_catalog_display_order ON laundry_service_catalog(display_order);

-- Pre-seed service catalog
INSERT INTO laundry_service_catalog (
    slug, title, description, color_code, base_price, pricing_unit,
    turnaround_hours, express_fee, express_hours, display_order, is_active
)
VALUES 
    ('wash-fold', 'Wash & Fold', 'Standard washing and folding', '#4CAF50', 5.00, 'kg',   48, 15.00, 24, 1, TRUE),
    ('dry-clean', 'Dry Clean', 'Professional dry cleaning',     '#2196F3', 10.00, 'item', 72, 25.00, 48, 2, TRUE),
    ('clean-press', 'Clean & Press', 'Washing with ironing',    '#FF9800', 8.00, 'item',  48, 20.00, 24, 3, TRUE),
    ('shoe-care', 'Shoe Care', 'Professional shoe cleaning',    '#9C27B0', 15.00, 'item', 72, 30.00, 48, 4, TRUE),
    ('home-linens', 'Home Linens', 'Bedsheets and curtains',    '#00BCD4', 6.00, 'kg',    48, 18.00, 24, 5, TRUE),
    ('carpet-cleaning', 'Carpet Cleaning', 'Professional carpet cleaning', '#8B4513', 20.00, 'item', 96, 40.00, 72, 6, TRUE),
    ('leather-care', 'Leather Care', 'Leather jacket and item care', '#E91E63', 25.00, 'item', 96, 50.00, 72, 7, TRUE),
    ('bag', 'Bag Care', 'Cleaning and care for handbags, backpacks, and travel bags',
        '#795548', 7.00, 'item', 72, 20.00, 48, 8, TRUE)
ON CONFLICT (slug) DO NOTHING;


-- =====================================================
-- TABLE 2: laundry_service_products
-- =====================================================

CREATE TABLE IF NOT EXISTS laundry_service_products (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    service_slug VARCHAR(100) NOT NULL REFERENCES laundry_service_catalog(slug) ON DELETE CASCADE,

    -- Product Details
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(100) NOT NULL,
    description TEXT,
    icon_url VARCHAR(500),

    -- Pricing (NOW REQUIRED - every product has a price)
    price DECIMAL(10,2) NOT NULL, -- Explicit price per item/unit
    pricing_unit VARCHAR(20) NOT NULL DEFAULT 'item', -- item or kg

    -- Attributes
    typical_weight DECIMAL(8,3), -- Average weight in kg (for weight-based items)
    requires_special_care BOOLEAN DEFAULT FALSE,
    special_care_fee DECIMAL(10,2) DEFAULT 0,

    -- Display
    display_order INTEGER DEFAULT 0,
    is_active BOOLEAN DEFAULT TRUE,

    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT unique_service_product_slug UNIQUE(service_slug, slug)
);

DROP INDEX IF EXISTS idx_laundry_service_products_service CASCADE;
DROP INDEX IF EXISTS idx_laundry_service_products_is_active CASCADE;
DROP INDEX IF EXISTS idx_laundry_service_products_display_order CASCADE;
DROP INDEX IF EXISTS idx_laundry_service_products_slug CASCADE;

CREATE INDEX idx_laundry_service_products_service ON laundry_service_products(service_slug);
CREATE INDEX idx_laundry_service_products_is_active ON laundry_service_products(is_active);
CREATE INDEX idx_laundry_service_products_display_order ON laundry_service_products(display_order);
CREATE INDEX idx_laundry_service_products_slug ON laundry_service_products(slug);

-- =====================================================
-- TABLE: laundry_orders
-- =====================================================
CREATE TABLE IF NOT EXISTS laundry_orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_number VARCHAR(50) NOT NULL UNIQUE,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    category_slug VARCHAR(100) NOT NULL,
    status VARCHAR(50) DEFAULT 'pending',
    address TEXT NOT NULL,
    latitude DECIMAL(10, 8),
    longitude DECIMAL(11, 8),
    service_date TIMESTAMP WITH TIME ZONE,
    total DECIMAL(10,2) NOT NULL,
    provider_id UUID REFERENCES service_provider_profiles(id),
    is_express BOOLEAN DEFAULT FALSE,
    special_notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

DROP INDEX IF EXISTS idx_laundry_orders_user_id CASCADE;
DROP INDEX IF EXISTS idx_laundry_orders_provider_id CASCADE;
DROP INDEX IF EXISTS idx_laundry_orders_status CASCADE;
DROP INDEX IF EXISTS idx_laundry_orders_category_slug CASCADE;
DROP INDEX IF EXISTS idx_laundry_orders_created_at CASCADE;

CREATE INDEX idx_laundry_orders_user_id ON laundry_orders(user_id);
CREATE INDEX idx_laundry_orders_provider_id ON laundry_orders(provider_id);
CREATE INDEX idx_laundry_orders_status ON laundry_orders(status);
CREATE INDEX idx_laundry_orders_category_slug ON laundry_orders(category_slug);
CREATE INDEX idx_laundry_orders_created_at ON laundry_orders(created_at DESC);

-- =====================================================
-- SEED: Products for Wash & Fold (UPDATED - All have prices)
-- Price calculated as: typical_weight × service.base_price (Rs 5/kg)
-- =====================================================

INSERT INTO laundry_service_products (id, service_slug, name, slug, description, price, pricing_unit, typical_weight, display_order)
VALUES
    ('550e8400-e29b-41d4-a716-446655440001', 'wash-fold', 'T-Shirt', 't-shirt', 'Regular cotton t-shirt', 0.75, 'item', 0.15, 1),
    ('550e8400-e29b-41d4-a716-446655440002', 'wash-fold', 'Shirt', 'shirt', 'Formal or casual shirt', 1.00, 'item', 0.20, 2),
    ('550e8400-e29b-41d4-a716-446655440003', 'wash-fold', 'Pants', 'pants', 'Trousers/pants', 1.75, 'item', 0.35, 3),
    ('550e8400-e29b-41d4-a716-446655440004', 'wash-fold', 'Shorts', 'shorts', 'Casual shorts', 1.00, 'item', 0.20, 4),
    ('550e8400-e29b-41d4-a716-446655440005', 'wash-fold', 'Jeans', 'jeans', 'Denim jeans', 3.00, 'item', 0.60, 5),
    ('550e8400-e29b-41d4-a716-446655440006', 'wash-fold', 'Dress', 'dress', 'Casual dress', 1.50, 'item', 0.30, 6),
    ('550e8400-e29b-41d4-a716-446655440007', 'wash-fold', 'Skirt', 'skirt', 'Skirt', 1.25, 'item', 0.25, 7),
    ('550e8400-e29b-41d4-a716-446655440008', 'wash-fold', 'Sweater', 'sweater', 'Pullover/sweater', 2.00, 'item', 0.40, 8),
    ('550e8400-e29b-41d4-a716-446655440009', 'wash-fold', 'Hoodie', 'hoodie', 'Hooded sweatshirt', 2.50, 'item', 0.50, 9),
    ('550e8400-e29b-41d4-a716-44665544000a', 'wash-fold', 'Jacket', 'jacket', 'Light jacket', 3.00, 'item', 0.60, 10),
    ('550e8400-e29b-41d4-a716-44665544000b', 'wash-fold', 'Undergarments', 'undergarments', 'Underwear, socks (per piece)', 0.25, 'item', 0.05, 11),
    ('550e8400-e29b-41d4-a716-44665544000c', 'wash-fold', 'Towel', 'towel', 'Bath towel', 2.50, 'item', 0.50, 12)
ON CONFLICT (service_slug, slug) DO NOTHING;

-- =====================================================
-- SEED: Products for Dry Clean (Already had prices)
-- =====================================================

INSERT INTO laundry_service_products (id, service_slug, name, slug, description, price, pricing_unit, display_order)
VALUES
    ('550e8400-e29b-41d4-a716-44665544010d', 'dry-clean', 'Suit (2-piece)', 'suit-2pc', 'Jacket and pants', 25.00, 'item', 1),
    ('550e8400-e29b-41d4-a716-44665544010e', 'dry-clean', 'Suit (3-piece)', 'suit-3pc', 'Jacket, pants, and vest', 35.00, 'item', 2),
    ('550e8400-e29b-41d4-a716-44665544010f', 'dry-clean', 'Blazer/Jacket', 'blazer', 'Formal blazer or jacket', 15.00, 'item', 3),
    ('550e8400-e29b-41d4-a716-446655440110', 'dry-clean', 'Dress Pants', 'dress-pants', 'Formal trousers', 10.00, 'item', 4),
    ('550e8400-e29b-41d4-a716-446655440111', 'dry-clean', 'Dress/Gown', 'dress-gown', 'Formal dress or gown', 20.00, 'item', 5),
    ('550e8400-e29b-41d4-a716-446655440112', 'dry-clean', 'Evening Gown', 'evening-gown', 'Long formal gown', 30.00, 'item', 6),
    ('550e8400-e29b-41d4-a716-446655440113', 'dry-clean', 'Coat/Overcoat', 'coat', 'Winter coat', 20.00, 'item', 7),
    ('550e8400-e29b-41d4-a716-446655440114', 'dry-clean', 'Tie', 'tie', 'Necktie', 5.00, 'item', 8),
    ('550e8400-e29b-41d4-a716-446655440115', 'dry-clean', 'Scarf', 'scarf', 'Silk or wool scarf', 8.00, 'item', 9),
    ('550e8400-e29b-41d4-a716-446655440116', 'dry-clean', 'Sherwani', 'sherwani', 'Traditional Pakistani formal wear', 40.00, 'item', 10),
    ('550e8400-e29b-41d4-a716-446655440117', 'dry-clean', 'Wedding Dress', 'wedding-dress', 'Bridal dress', 100.00, 'item', 11)
ON CONFLICT (service_slug, slug) DO NOTHING;

-- =====================================================
-- SEED: Products for Clean & Press (Already had prices)
-- =====================================================

INSERT INTO laundry_service_products (id, service_slug, name, slug, description, price, pricing_unit, display_order)
VALUES
    ('550e8400-e29b-41d4-a716-446655440118', 'clean-press', 'Shirt', 'shirt', 'Dress shirt with ironing', 8.00, 'item', 1),
    ('550e8400-e29b-41d4-a716-446655440119', 'clean-press', 'Pants', 'pants', 'Trousers with pressing', 8.00, 'item', 2),
    ('550e8400-e29b-41d4-a716-44665544011a', 'clean-press', 'Shalwar Kameez', 'shalwar-kameez', 'Traditional Pakistani outfit', 12.00, 'item', 3),
    ('550e8400-e29b-41d4-a716-44665544011b', 'clean-press', 'Kurta', 'kurta', 'Traditional shirt', 8.00, 'item', 4),
    ('550e8400-e29b-41d4-a716-44665544011c', 'clean-press', 'Dress', 'dress', 'Dress with pressing', 10.00, 'item', 5),
    ('550e8400-e29b-41d4-a716-44665544011d', 'clean-press', 'Blouse', 'blouse', 'Ladies blouse', 8.00, 'item', 6),
    ('550e8400-e29b-41d4-a716-44665544011e', 'clean-press', 'Table Cloth', 'table-cloth', 'Dining table cloth', 8.00, 'item', 7)
ON CONFLICT (service_slug, slug) DO NOTHING;

-- =====================================================
-- SEED: Products for Shoe Care (Already had prices)
-- =====================================================

INSERT INTO laundry_service_products (id, service_slug, name, slug, description, price, pricing_unit, display_order)
VALUES
    ('550e8400-e29b-41d4-a716-44665544011f', 'shoe-care', 'Sneakers', 'sneakers', 'Sports/casual sneakers', 15.00, 'item', 1),
    ('550e8400-e29b-41d4-a716-446655440120', 'shoe-care', 'Formal Shoes', 'formal-shoes', 'Leather formal shoes', 20.00, 'item', 2),
    ('550e8400-e29b-41d4-a716-446655440121', 'shoe-care', 'Boots', 'boots', 'Ankle or knee-high boots', 25.00, 'item', 3),
    ('550e8400-e29b-41d4-a716-446655440122', 'shoe-care', 'Sandals', 'sandals', 'Open-toe sandals', 12.00, 'item', 4),
    ('550e8400-e29b-41d4-a716-446655440123', 'shoe-care', 'Heels', 'heels', 'High heels', 18.00, 'item', 5),
    ('550e8400-e29b-41d4-a716-446655440124', 'shoe-care', 'Loafers', 'loafers', 'Slip-on shoes', 18.00, 'item', 6),
    ('550e8400-e29b-41d4-a716-446655440125', 'shoe-care', 'Sports Shoes', 'sports-shoes', 'Running/training shoes', 15.00, 'item', 7),
    ('550e8400-e29b-41d4-a716-446655440126', 'shoe-care', 'Khussas', 'khussas', 'Traditional Pakistani shoes', 15.00, 'item', 8)
ON CONFLICT (service_slug, slug) DO NOTHING;

-- =====================================================
-- SEED: Products for Home Linens (UPDATED - All have prices)
-- Price calculated as: typical_weight × service.base_price (Rs 6/kg)
-- =====================================================

INSERT INTO laundry_service_products (id, service_slug, name, slug, description, price, pricing_unit, typical_weight, display_order)
VALUES
    ('550e8400-e29b-41d4-a716-446655440127', 'home-linens', 'Bed Sheet (Single)', 'bed-sheet-single', 'Single bed sheet', 4.80, 'item', 0.80, 1),
    ('550e8400-e29b-41d4-a716-446655440128', 'home-linens', 'Bed Sheet (Double)', 'bed-sheet-double', 'Double bed sheet', 7.20, 'item', 1.20, 2),
    ('550e8400-e29b-41d4-a716-446655440129', 'home-linens', 'Bed Sheet (King)', 'bed-sheet-king', 'King size bed sheet', 9.00, 'item', 1.50, 3),
    ('550e8400-e29b-41d4-a716-44665544012a', 'home-linens', 'Duvet Cover', 'duvet-cover', 'Comforter cover', 6.00, 'item', 1.00, 4),
    ('550e8400-e29b-41d4-a716-44665544012b', 'home-linens', 'Pillow Case', 'pillow-case', 'Pillow cover', 0.90, 'item', 0.15, 5),
    ('550e8400-e29b-41d4-a716-44665544012c', 'home-linens', 'Comforter', 'comforter', 'Heavy blanket/comforter', 15.00, 'item', 2.50, 6),
    ('550e8400-e29b-41d4-a716-44665544012d', 'home-linens', 'Blanket', 'blanket', 'Regular blanket', 9.00, 'item', 1.50, 7),
    ('550e8400-e29b-41d4-a716-44665544012e', 'home-linens', 'Curtains (per panel)', 'curtains', 'Window curtains', 6.00, 'item', 1.00, 8),
    ('550e8400-e29b-41d4-a716-44665544012f', 'home-linens', 'Table Cloth', 'table-cloth', 'Dining table cloth', 3.60, 'item', 0.60, 9),
    ('550e8400-e29b-41d4-a716-446655440130', 'home-linens', 'Towel Set', 'towel-set', 'Bath towel set', 7.20, 'item', 1.20, 10)
ON CONFLICT (service_slug, slug) DO NOTHING;

-- =====================================================
-- SEED: Products for Carpet Cleaning (Already had prices)
-- =====================================================

INSERT INTO laundry_service_products (id, service_slug, name, slug, description, price, pricing_unit, display_order)
VALUES
    ('550e8400-e29b-41d4-a716-446655440131', 'carpet-cleaning', 'Small Rug', 'small-rug', 'Up to 4x6 feet', 20.00, 'item', 1),
    ('550e8400-e29b-41d4-a716-446655440132', 'carpet-cleaning', 'Medium Rug', 'medium-rug', '5x7 to 6x9 feet', 35.00, 'item', 2),
    ('550e8400-e29b-41d4-a716-446655440133', 'carpet-cleaning', 'Large Rug', 'large-rug', '8x10 feet or larger', 50.00, 'item', 3),
    ('550e8400-e29b-41d4-a716-446655440134', 'carpet-cleaning', 'Runner', 'runner', 'Hallway runner', 25.00, 'item', 4),
    ('550e8400-e29b-41d4-a716-446655440135', 'carpet-cleaning', 'Prayer Mat', 'prayer-mat', 'Small prayer rug', 15.00, 'item', 5),
    ('550e8400-e29b-41d4-a716-446655440136', 'carpet-cleaning', 'Door Mat', 'door-mat', 'Entrance mat', 12.00, 'item', 6)
ON CONFLICT (service_slug, slug) DO NOTHING;

-- =====================================================
-- SEED: Products for Leather Care (Already had prices)
-- =====================================================

INSERT INTO laundry_service_products (id, service_slug, name, slug, description, price, pricing_unit, requires_special_care, display_order)
VALUES
    ('550e8400-e29b-41d4-a716-446655440137', 'leather-care', 'Leather Jacket', 'leather-jacket', 'Full leather jacket', 25.00, 'item', TRUE, 1),
    ('550e8400-e29b-41d4-a716-446655440138', 'leather-care', 'Leather Pants', 'leather-pants', 'Leather trousers', 20.00, 'item', TRUE, 2),
    ('550e8400-e29b-41d4-a716-446655440139', 'leather-care', 'Leather Bag', 'leather-bag', 'Handbag or purse', 15.00, 'item', TRUE, 3),
    ('550e8400-e29b-41d4-a716-44665544013a', 'leather-care', 'Leather Shoes', 'leather-shoes', 'Professional shoe care', 20.00, 'item', TRUE, 4),
    ('550e8400-e29b-41d4-a716-44665544013b', 'leather-care', 'Leather Gloves', 'leather-gloves', 'Pair of leather gloves', 10.00, 'item', TRUE, 5),
    ('550e8400-e29b-41d4-a716-44665544013c', 'leather-care', 'Leather Sofa Cushion', 'leather-sofa-cushion', 'Single cushion', 30.00, 'item', TRUE, 6)
ON CONFLICT (service_slug, slug) DO NOTHING;

INSERT INTO laundry_service_products (
    id, service_slug, name, slug, description, price,
    pricing_unit, typical_weight, display_order
)
VALUES
    ('550e8400-e29b-41d4-a716-446655440050', 'bag', 'Handbag', 'handbag',
     'Regular handbag (fabric or leather)', 5.00, 'item', 0.80, 1),
    ('550e8400-e29b-41d4-a716-446655440051', 'bag', 'Backpack', 'backpack',
     'Everyday backpack', 6.00, 'item', 1.00, 2),
    ('550e8400-e29b-41d4-a716-446655440052', 'bag', 'Laptop Bag', 'laptop-bag',
     'Padded laptop bag', 5.50, 'item', 0.90, 3),
    ('550e8400-e29b-41d4-a716-446655440053', 'bag', 'Travel Duffel', 'travel-duffel',
     'Medium travel duffel bag', 7.00, 'item', 1.50, 4)
ON CONFLICT (service_slug, slug) DO NOTHING;

-- =====================================================
-- =====================================================
-- TABLE 3: laundry_order_items
-- =====================================================
CREATE TABLE IF NOT EXISTS laundry_order_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL REFERENCES laundry_orders(id) ON DELETE CASCADE,
    service_slug VARCHAR(100) NOT NULL,
    product_slug VARCHAR(100),
    item_type VARCHAR(100) NOT NULL,
    quantity INTEGER DEFAULT 1,
    weight DECIMAL(8,3),
    qr_code VARCHAR(255) NOT NULL UNIQUE,
    status VARCHAR(50) DEFAULT 'pending',
    has_issue BOOLEAN DEFAULT FALSE,
    issue_description TEXT,
    price DECIMAL(10,2) NOT NULL,
    received_at TIMESTAMP WITH TIME ZONE,
    packed_at TIMESTAMP WITH TIME ZONE,
    delivered_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

DROP INDEX IF EXISTS idx_laundry_order_items_order_id CASCADE;
DROP INDEX IF EXISTS idx_laundry_order_items_qr_code CASCADE;
DROP INDEX IF EXISTS idx_laundry_order_items_status CASCADE;
DROP INDEX IF EXISTS idx_laundry_order_items_service_slug CASCADE;
DROP INDEX IF EXISTS idx_laundry_order_items_product_slug CASCADE;
DROP INDEX IF EXISTS idx_laundry_order_items_service_product CASCADE;

CREATE INDEX idx_laundry_order_items_order_id ON laundry_order_items(order_id);
CREATE INDEX idx_laundry_order_items_qr_code ON laundry_order_items(qr_code);
CREATE INDEX idx_laundry_order_items_status ON laundry_order_items(status);
CREATE INDEX idx_laundry_order_items_service_slug ON laundry_order_items(service_slug);
CREATE INDEX idx_laundry_order_items_product_slug ON laundry_order_items(product_slug);
CREATE INDEX idx_laundry_order_items_service_product ON laundry_order_items(service_slug, product_slug);

-- =====================================================
-- TABLE 4: laundry_pickups
-- =====================================================
CREATE TABLE IF NOT EXISTS laundry_pickups (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL UNIQUE REFERENCES laundry_orders(id) ON DELETE CASCADE,
    provider_id UUID NOT NULL REFERENCES service_provider_profiles(id) ON DELETE CASCADE,
    scheduled_at TIMESTAMP WITH TIME ZONE NOT NULL,
    arrived_at TIMESTAMP WITH TIME ZONE,
    picked_up_at TIMESTAMP WITH TIME ZONE,
    status VARCHAR(50) DEFAULT 'scheduled',
    photo_url VARCHAR(500),
    notes TEXT,
    bag_count INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

DROP INDEX IF EXISTS idx_laundry_pickups_order_id CASCADE;
DROP INDEX IF EXISTS idx_laundry_pickups_provider_id CASCADE;
DROP INDEX IF EXISTS idx_laundry_pickups_status CASCADE;
DROP INDEX IF EXISTS idx_laundry_pickups_scheduled_at CASCADE;

CREATE INDEX idx_laundry_pickups_order_id ON laundry_pickups(order_id);
CREATE INDEX idx_laundry_pickups_provider_id ON laundry_pickups(provider_id);
CREATE INDEX idx_laundry_pickups_status ON laundry_pickups(status);
CREATE INDEX idx_laundry_pickups_scheduled_at ON laundry_pickups(scheduled_at);

-- =====================================================
-- TABLE 5: laundry_deliveries
-- =====================================================
CREATE TABLE IF NOT EXISTS laundry_deliveries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL UNIQUE REFERENCES laundry_orders(id) ON DELETE CASCADE,
    provider_id UUID NOT NULL REFERENCES service_provider_profiles(id) ON DELETE CASCADE,
    scheduled_at TIMESTAMP WITH TIME ZONE NOT NULL,
    arrived_at TIMESTAMP WITH TIME ZONE,
    delivered_at TIMESTAMP WITH TIME ZONE,
    status VARCHAR(50) DEFAULT 'scheduled',
    photo_url VARCHAR(500),
    recipient_name VARCHAR(255),
    recipient_signature TEXT,
    notes TEXT,
    reschedule_count INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

DROP INDEX IF EXISTS idx_laundry_deliveries_order_id CASCADE;
DROP INDEX IF EXISTS idx_laundry_deliveries_provider_id CASCADE;
DROP INDEX IF EXISTS idx_laundry_deliveries_status CASCADE;
DROP INDEX IF EXISTS idx_laundry_deliveries_scheduled_at CASCADE;

CREATE INDEX idx_laundry_deliveries_order_id ON laundry_deliveries(order_id);
CREATE INDEX idx_laundry_deliveries_provider_id ON laundry_deliveries(provider_id);
CREATE INDEX idx_laundry_deliveries_status ON laundry_deliveries(status);
CREATE INDEX idx_laundry_deliveries_scheduled_at ON laundry_deliveries(scheduled_at);

-- =====================================================
-- TABLE 6: laundry_issues
-- =====================================================
CREATE TABLE IF NOT EXISTS laundry_issues (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL REFERENCES laundry_orders(id) ON DELETE CASCADE,
    customer_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider_id UUID NOT NULL REFERENCES service_provider_profiles(id) ON DELETE CASCADE,
    issue_type VARCHAR(100) NOT NULL,
    description TEXT NOT NULL,
    priority VARCHAR(20) DEFAULT 'medium',
    status VARCHAR(50) DEFAULT 'open',
    resolution TEXT,
    refund_amount DECIMAL(10,2),
    compensation_type VARCHAR(100),
    resolved_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

DROP INDEX IF EXISTS idx_laundry_issues_order_id CASCADE;
DROP INDEX IF EXISTS idx_laundry_issues_customer_id CASCADE;
DROP INDEX IF EXISTS idx_laundry_issues_provider_id CASCADE;
DROP INDEX IF EXISTS idx_laundry_issues_status CASCADE;
DROP INDEX IF EXISTS idx_laundry_issues_priority CASCADE;
DROP INDEX IF EXISTS idx_laundry_issues_created_at CASCADE;

CREATE INDEX idx_laundry_issues_order_id ON laundry_issues(order_id);
CREATE INDEX idx_laundry_issues_customer_id ON laundry_issues(customer_id);
CREATE INDEX idx_laundry_issues_provider_id ON laundry_issues(provider_id);
CREATE INDEX idx_laundry_issues_status ON laundry_issues(status);
CREATE INDEX idx_laundry_issues_priority ON laundry_issues(priority);
CREATE INDEX idx_laundry_issues_created_at ON laundry_issues(created_at DESC);

-- =====================================================
-- TRIGGERS for updated_at
-- =====================================================

CREATE OR REPLACE FUNCTION update_laundry_service_catalog_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_update_laundry_service_catalog_updated_at ON laundry_service_catalog;
CREATE TRIGGER trigger_update_laundry_service_catalog_updated_at
    BEFORE UPDATE ON laundry_service_catalog
    FOR EACH ROW
    EXECUTE FUNCTION update_laundry_service_catalog_updated_at();

CREATE OR REPLACE FUNCTION update_laundry_table_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_update_laundry_service_products_updated_at ON laundry_service_products;
CREATE TRIGGER trigger_update_laundry_service_products_updated_at
    BEFORE UPDATE ON laundry_service_products
    FOR EACH ROW
    EXECUTE FUNCTION update_laundry_table_updated_at();

DROP TRIGGER IF EXISTS trigger_update_laundry_orders_updated_at ON laundry_orders;
CREATE TRIGGER trigger_update_laundry_orders_updated_at
    BEFORE UPDATE ON laundry_orders
    FOR EACH ROW
    EXECUTE FUNCTION update_laundry_table_updated_at();

DROP TRIGGER IF EXISTS trigger_update_laundry_order_items_updated_at ON laundry_order_items;
CREATE TRIGGER trigger_update_laundry_order_items_updated_at
    BEFORE UPDATE ON laundry_order_items
    FOR EACH ROW
    EXECUTE FUNCTION update_laundry_table_updated_at();

DROP TRIGGER IF EXISTS trigger_update_laundry_pickups_updated_at ON laundry_pickups;
CREATE TRIGGER trigger_update_laundry_pickups_updated_at
    BEFORE UPDATE ON laundry_pickups
    FOR EACH ROW
    EXECUTE FUNCTION update_laundry_table_updated_at();

DROP TRIGGER IF EXISTS trigger_update_laundry_deliveries_updated_at ON laundry_deliveries;
CREATE TRIGGER trigger_update_laundry_deliveries_updated_at
    BEFORE UPDATE ON laundry_deliveries
    FOR EACH ROW
    EXECUTE FUNCTION update_laundry_table_updated_at();

DROP TRIGGER IF EXISTS trigger_update_laundry_issues_updated_at ON laundry_issues;
CREATE TRIGGER trigger_update_laundry_issues_updated_at
    BEFORE UPDATE ON laundry_issues
    FOR EACH ROW
    EXECUTE FUNCTION update_laundry_table_updated_at();

-- =====================================================
-- INDEXES for performance
-- =====================================================

DROP INDEX IF EXISTS idx_laundry_issues_order_provider CASCADE;
DROP INDEX IF EXISTS idx_laundry_pickups_provider_status CASCADE;
DROP INDEX IF EXISTS idx_laundry_deliveries_provider_status CASCADE;
DROP INDEX IF EXISTS idx_laundry_order_items_status_created CASCADE;

CREATE INDEX idx_laundry_issues_order_provider ON laundry_issues(order_id, provider_id);
CREATE INDEX idx_laundry_pickups_provider_status ON laundry_pickups(provider_id, status) WHERE status IN ('scheduled', 'en_route', 'arrived');
CREATE INDEX idx_laundry_deliveries_provider_status ON laundry_deliveries(provider_id, status) WHERE status IN ('scheduled', 'en_route', 'arrived');
CREATE INDEX idx_laundry_order_items_status_created ON laundry_order_items(status, created_at DESC);

-- =====================================================
-- VIEW: Services with Products
-- =====================================================

CREATE OR REPLACE VIEW v_laundry_services_with_products AS
SELECT 
    lsc.id,
    lsc.slug,
    lsc.title,
    lsc.description,
    lsc.color_code,
    lsc.base_price,
    lsc.pricing_unit,
    lsc.turnaround_hours,
    lsc.express_fee,
    lsc.express_hours,
    lsc.is_active,
    COUNT(lsp.id) AS product_count,
    json_agg(
        json_build_object(
            'id', lsp.id,
            'name', lsp.name,
            'slug', lsp.slug,
            'description', lsp.description,
            'price', lsp.price,
            'pricing_unit', lsp.pricing_unit,
            'typical_weight', lsp.typical_weight,
            'requires_special_care', lsp.requires_special_care,
            'special_care_fee', lsp.special_care_fee,
            'display_order', lsp.display_order
        ) ORDER BY lsp.display_order
    ) FILTER (WHERE lsp.id IS NOT NULL) AS products
FROM laundry_service_catalog lsc
LEFT JOIN laundry_service_products lsp ON lsc.slug = lsp.service_slug AND lsp.is_active = TRUE
WHERE lsc.is_active = TRUE
GROUP BY lsc.id, lsc.slug, lsc.title, lsc.description, lsc.color_code, 
         lsc.base_price, lsc.pricing_unit, lsc.turnaround_hours, 
         lsc.express_fee, lsc.express_hours, lsc.is_active
ORDER BY lsc.display_order;

-- =====================================================
-- COMMENTS
-- =====================================================

COMMENT ON TABLE laundry_service_catalog IS 'Master catalog of laundry services available';
COMMENT ON TABLE laundry_service_products IS 'Available products/items for each laundry service - ALL products have explicit prices';
COMMENT ON TABLE laundry_order_items IS 'Individual items in a laundry order with QR tracking';
COMMENT ON TABLE laundry_pickups IS 'Pickup schedules and tracking';
COMMENT ON TABLE laundry_deliveries IS 'Delivery schedules and tracking';
COMMENT ON TABLE laundry_issues IS 'Customer complaints and issue tracking';
COMMENT ON VIEW v_laundry_services_with_products IS 'Services with their product listings and counts';

COMMENT ON COLUMN laundry_service_products.price IS 'Explicit price per item - REQUIRED for all products';
COMMENT ON COLUMN laundry_service_products.typical_weight IS 'Average weight in kg - kept for reference and bulk calculations';
COMMENT ON COLUMN laundry_order_items.product_slug IS 'Links to the specific product in laundry_service_products';
COMMENT ON COLUMN laundry_order_items.qr_code IS 'Unique QR code for item tracking throughout the process';