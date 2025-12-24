-- Laundry Service Schema Migrations
-- Updated: All products now have explicit prices in INR (Indian Rupees)
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
    -- ✅ NEW: Category slug for filtering by service category
    category_slug VARCHAR(100) NOT NULL DEFAULT 'laundry',
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

DROP INDEX IF EXISTS idx_laundry_service_catalog_slug CASCADE;
DROP INDEX IF EXISTS idx_laundry_service_catalog_is_active CASCADE;
DROP INDEX IF EXISTS idx_laundry_service_catalog_display_order CASCADE;
DROP INDEX IF EXISTS idx_laundry_service_catalog_category CASCADE;

CREATE INDEX idx_laundry_service_catalog_slug ON laundry_service_catalog(slug);
CREATE INDEX idx_laundry_service_catalog_is_active ON laundry_service_catalog(is_active);
CREATE INDEX idx_laundry_service_catalog_display_order ON laundry_service_catalog(display_order);
CREATE INDEX idx_laundry_service_catalog_category ON laundry_service_catalog(category_slug);

-- =====================================================
-- SEED: Service Catalog with INR Pricing
-- =====================================================
INSERT INTO laundry_service_catalog (
    slug, title, description, color_code, base_price, pricing_unit,
    turnaround_hours, express_fee, express_hours, display_order, category_slug, is_active
)
VALUES 
    ('wash-fold', 'Wash & Fold', 'Standard washing and folding service', '#4CAF50', 49.00, 'kg', 48, 99.00, 24, 1, 'laundry', TRUE),
    ('dry-clean', 'Dry Clean', 'Professional dry cleaning for delicate fabrics', '#2196F3', 149.00, 'item', 72, 199.00, 48, 2, 'laundry', TRUE),
    ('clean-press', 'Clean & Press', 'Washing with professional ironing and pressing', '#FF9800', 39.00, 'item', 48, 79.00, 24, 3, 'laundry', TRUE),
    ('shoe-care', 'Shoe Care', 'Professional shoe cleaning and polishing', '#9C27B0', 249.00, 'item', 72, 149.00, 48, 4, 'laundry', TRUE),
    ('home-linens', 'Home Linens', 'Bedsheets, curtains and household textiles', '#00BCD4', 59.00, 'kg', 48, 119.00, 24, 5, 'laundry', TRUE),
    ('carpet-cleaning', 'Carpet Cleaning', 'Deep cleaning for carpets and rugs', '#8B4513', 499.00, 'item', 96, 299.00, 72, 6, 'laundry', TRUE),
    ('leather-care', 'Leather Care', 'Specialized care for leather items', '#E91E63', 599.00, 'item', 96, 399.00, 72, 7, 'laundry', TRUE),
    ('bag', 'Bag Care', 'Cleaning and care for handbags, backpacks, and travel bags', '#795548', 299.00, 'item', 72, 199.00, 48, 8, 'laundry', TRUE)
ON CONFLICT (slug) DO UPDATE SET
    title = EXCLUDED.title,
    description = EXCLUDED.description,
    color_code = EXCLUDED.color_code,
    base_price = EXCLUDED.base_price,
    pricing_unit = EXCLUDED.pricing_unit,
    turnaround_hours = EXCLUDED.turnaround_hours,
    express_fee = EXCLUDED.express_fee,
    express_hours = EXCLUDED.express_hours,
    display_order = EXCLUDED.display_order,
    category_slug = EXCLUDED.category_slug,
    is_active = EXCLUDED.is_active,
    updated_at = CURRENT_TIMESTAMP;


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

    -- Pricing (NOW REQUIRED - every product has a price in INR)
    price DECIMAL(10,2) NOT NULL, -- Explicit price per item/unit in INR
    pricing_unit VARCHAR(20) NOT NULL DEFAULT 'item', -- item or kg

    -- Attributes
    typical_weight DECIMAL(8,3), -- Average weight in kg (for weight-based items)
    requires_special_care BOOLEAN DEFAULT FALSE,
    special_care_fee DECIMAL(10,2) DEFAULT 0,

    -- Display & Category
    display_order INTEGER DEFAULT 0,
    category_slug VARCHAR(100) NOT NULL DEFAULT 'laundry',
    is_active BOOLEAN DEFAULT TRUE,

    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT unique_service_product_slug UNIQUE(service_slug, slug)
);

DROP INDEX IF EXISTS idx_laundry_service_products_service CASCADE;
DROP INDEX IF EXISTS idx_laundry_service_products_is_active CASCADE;
DROP INDEX IF EXISTS idx_laundry_service_products_display_order CASCADE;
DROP INDEX IF EXISTS idx_laundry_service_products_slug CASCADE;
DROP INDEX IF EXISTS idx_laundry_service_products_category CASCADE;

CREATE INDEX idx_laundry_service_products_service ON laundry_service_products(service_slug);
CREATE INDEX idx_laundry_service_products_is_active ON laundry_service_products(is_active);
CREATE INDEX idx_laundry_service_products_display_order ON laundry_service_products(display_order);
CREATE INDEX idx_laundry_service_products_slug ON laundry_service_products(slug);
CREATE INDEX idx_laundry_service_products_category ON laundry_service_products(category_slug);

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
-- SEED: Products for Wash & Fold (INR Pricing)
-- Typical Indian laundry prices: ₹15-80 per item
-- =====================================================

INSERT INTO laundry_service_products (id, service_slug, name, slug, description, price, pricing_unit, typical_weight, display_order, category_slug, is_active)
VALUES
    ('550e8400-e29b-41d4-a716-446655440001', 'wash-fold', 'T-Shirt', 't-shirt', 'Regular cotton t-shirt', 19.00, 'item', 0.150, 1, 'laundry', TRUE),
    ('550e8400-e29b-41d4-a716-446655440002', 'wash-fold', 'Shirt', 'shirt', 'Formal or casual shirt', 25.00, 'item', 0.200, 2, 'laundry', TRUE),
    ('550e8400-e29b-41d4-a716-446655440003', 'wash-fold', 'Pants', 'pants', 'Trousers/pants', 35.00, 'item', 0.350, 3, 'laundry', TRUE),
    ('550e8400-e29b-41d4-a716-446655440004', 'wash-fold', 'Shorts', 'shorts', 'Casual shorts', 22.00, 'item', 0.200, 4, 'laundry', TRUE),
    ('550e8400-e29b-41d4-a716-446655440005', 'wash-fold', 'Jeans', 'jeans', 'Denim jeans', 45.00, 'item', 0.600, 5, 'laundry', TRUE),
    ('550e8400-e29b-41d4-a716-446655440006', 'wash-fold', 'Dress', 'dress', 'Casual dress', 40.00, 'item', 0.300, 6, 'laundry', TRUE),
    ('550e8400-e29b-41d4-a716-446655440007', 'wash-fold', 'Skirt', 'skirt', 'Skirt', 30.00, 'item', 0.250, 7, 'laundry', TRUE),
    ('550e8400-e29b-41d4-a716-446655440008', 'wash-fold', 'Sweater', 'sweater', 'Pullover/sweater', 45.00, 'item', 0.400, 8, 'laundry', TRUE),
    ('550e8400-e29b-41d4-a716-446655440009', 'wash-fold', 'Hoodie', 'hoodie', 'Hooded sweatshirt', 55.00, 'item', 0.500, 9, 'laundry', TRUE),
    ('550e8400-e29b-41d4-a716-44665544000a', 'wash-fold', 'Jacket', 'jacket', 'Light jacket', 65.00, 'item', 0.600, 10, 'laundry', TRUE),
    ('550e8400-e29b-41d4-a716-44665544000b', 'wash-fold', 'Undergarments', 'undergarments', 'Underwear, socks (per piece)', 12.00, 'item', 0.050, 11, 'laundry', TRUE),
    ('550e8400-e29b-41d4-a716-44665544000c', 'wash-fold', 'Towel', 'towel', 'Bath towel', 45.00, 'item', 0.500, 12, 'laundry', TRUE),
    ('550e8400-e29b-41d4-a716-44665544000d', 'wash-fold', 'Saree (Cotton)', 'saree-cotton', 'Cotton saree', 55.00, 'item', 0.450, 13, 'laundry', TRUE),
    ('550e8400-e29b-41d4-a716-44665544000e', 'wash-fold', 'Salwar Suit', 'salwar-suit', 'Salwar kameez set (3 piece)', 65.00, 'item', 0.500, 14, 'laundry', TRUE),
    ('550e8400-e29b-41d4-a716-44665544000f', 'wash-fold', 'Dupatta', 'dupatta', 'Cotton or chiffon dupatta', 25.00, 'item', 0.150, 15, 'laundry', TRUE),
    ('550e8400-e29b-41d4-a716-446655440010', 'wash-fold', 'Lungi', 'lungi', 'Traditional lungi', 20.00, 'item', 0.200, 16, 'laundry', TRUE),
    ('550e8400-e29b-41d4-a716-446655440011', 'wash-fold', 'Dhoti', 'dhoti', 'Traditional dhoti', 25.00, 'item', 0.250, 17, 'laundry', TRUE),
    ('550e8400-e29b-41d4-a716-446655440012', 'wash-fold', 'Gym Wear', 'gym-wear', 'Sports/gym clothing', 22.00, 'item', 0.180, 18, 'laundry', TRUE)
ON CONFLICT (service_slug, slug) DO UPDATE SET
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    price = EXCLUDED.price,
    pricing_unit = EXCLUDED.pricing_unit,
    typical_weight = EXCLUDED.typical_weight,
    display_order = EXCLUDED.display_order,
    is_active = EXCLUDED.is_active,
    updated_at = CURRENT_TIMESTAMP;

-- =====================================================
-- SEED: Products for Dry Clean (INR Pricing)
-- Premium dry cleaning prices: ₹80-4000 per item
-- =====================================================

INSERT INTO laundry_service_products (id, service_slug, name, slug, description, price, pricing_unit, typical_weight, requires_special_care, special_care_fee, display_order, category_slug, is_active)
VALUES
    ('550e8400-e29b-41d4-a716-44665544010d', 'dry-clean', 'Suit (2-piece)', 'suit-2pc', 'Jacket and pants', 499.00, 'item', 1.200, TRUE, 100.00, 1, 'laundry', TRUE),
    ('550e8400-e29b-41d4-a716-44665544010e', 'dry-clean', 'Suit (3-piece)', 'suit-3pc', 'Jacket, pants, and vest', 699.00, 'item', 1.500, TRUE, 150.00, 2, 'laundry', TRUE),
    ('550e8400-e29b-41d4-a716-44665544010f', 'dry-clean', 'Blazer/Jacket', 'blazer', 'Formal blazer or jacket', 299.00, 'item', 0.700, TRUE, 75.00, 3, 'laundry', TRUE),
    ('550e8400-e29b-41d4-a716-446655440110', 'dry-clean', 'Dress Pants', 'dress-pants', 'Formal trousers', 179.00, 'item', 0.400, FALSE, 0.00, 4, 'laundry', TRUE),
    ('550e8400-e29b-41d4-a716-446655440111', 'dry-clean', 'Dress/Gown', 'dress-gown', 'Formal dress or gown', 399.00, 'item', 0.600, TRUE, 100.00, 5, 'laundry', TRUE),
    ('550e8400-e29b-41d4-a716-446655440112', 'dry-clean', 'Evening Gown', 'evening-gown', 'Long formal gown with embellishments', 799.00, 'item', 1.000, TRUE, 200.00, 6, 'laundry', TRUE),
    ('550e8400-e29b-41d4-a716-446655440113', 'dry-clean', 'Coat/Overcoat', 'coat', 'Winter coat or overcoat', 549.00, 'item', 1.200, TRUE, 100.00, 7, 'laundry', TRUE),
    ('550e8400-e29b-41d4-a716-446655440114', 'dry-clean', 'Tie', 'tie', 'Silk or polyester necktie', 89.00, 'item', 0.050, TRUE, 25.00, 8, 'laundry', TRUE),
    ('550e8400-e29b-41d4-a716-446655440115', 'dry-clean', 'Scarf', 'scarf', 'Silk or wool scarf', 129.00, 'item', 0.100, TRUE, 50.00, 9, 'laundry', TRUE),
    ('550e8400-e29b-41d4-a716-446655440116', 'dry-clean', 'Sherwani', 'sherwani', 'Traditional Indian formal wear', 999.00, 'item', 1.500, TRUE, 250.00, 10, 'laundry', TRUE),
    ('550e8400-e29b-41d4-a716-446655440117', 'dry-clean', 'Wedding Lehenga', 'wedding-lehenga', 'Bridal lehenga with dupatta', 2999.00, 'item', 3.000, TRUE, 500.00, 11, 'laundry', TRUE),
    ('550e8400-e29b-41d4-a716-446655440118', 'dry-clean', 'Wedding Sherwani', 'wedding-sherwani', 'Groom sherwani with accessories', 1499.00, 'item', 2.000, TRUE, 300.00, 12, 'laundry', TRUE),
    ('550e8400-e29b-41d4-a716-446655440119', 'dry-clean', 'Silk Saree', 'silk-saree', 'Pure silk saree', 599.00, 'item', 0.600, TRUE, 150.00, 13, 'laundry', TRUE),
    ('550e8400-e29b-41d4-a716-44665544011a', 'dry-clean', 'Designer Saree', 'designer-saree', 'Heavy work designer saree', 899.00, 'item', 0.800, TRUE, 200.00, 14, 'laundry', TRUE),
    ('550e8400-e29b-41d4-a716-44665544011b', 'dry-clean', 'Anarkali Suit', 'anarkali-suit', 'Embroidered anarkali dress', 699.00, 'item', 1.000, TRUE, 150.00, 15, 'laundry', TRUE),
    ('550e8400-e29b-41d4-a716-44665544011c', 'dry-clean', 'Kurta Pajama Set', 'kurta-pajama-set', 'Silk or embroidered kurta set', 399.00, 'item', 0.600, TRUE, 75.00, 16, 'laundry', TRUE)
ON CONFLICT (service_slug, slug) DO UPDATE SET
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    price = EXCLUDED.price,
    pricing_unit = EXCLUDED.pricing_unit,
    typical_weight = EXCLUDED.typical_weight,
    requires_special_care = EXCLUDED.requires_special_care,
    special_care_fee = EXCLUDED.special_care_fee,
    display_order = EXCLUDED.display_order,
    is_active = EXCLUDED.is_active,
    updated_at = CURRENT_TIMESTAMP;

-- =====================================================
-- SEED: Products for Clean & Press (INR Pricing)
-- Steam iron and press prices: ₹20-120 per item
-- =====================================================

INSERT INTO laundry_service_products (id, service_slug, name, slug, description, price, pricing_unit, typical_weight, display_order, category_slug, is_active)
VALUES
    ('550e8400-e29b-41d4-a716-446655440200', 'clean-press', 'Shirt', 'shirt', 'Formal shirt with steam ironing', 35.00, 'item', 0.200, 1, 'laundry', TRUE),
    ('550e8400-e29b-41d4-a716-446655440201', 'clean-press', 'Pants', 'pants', 'Trousers with crease pressing', 40.00, 'item', 0.350, 2, 'laundry', TRUE),
    ('550e8400-e29b-41d4-a716-446655440202', 'clean-press', 'T-Shirt', 't-shirt', 'T-shirt with light pressing', 25.00, 'item', 0.150, 3, 'laundry', TRUE),
    ('550e8400-e29b-41d4-a716-446655440203', 'clean-press', 'Kurta', 'kurta', 'Cotton or silk kurta', 45.00, 'item', 0.250, 4, 'laundry', TRUE),
    ('550e8400-e29b-41d4-a716-446655440204', 'clean-press', 'Kurta Pajama', 'kurta-pajama', 'Kurta with pajama set', 75.00, 'item', 0.450, 5, 'laundry', TRUE),
    ('550e8400-e29b-41d4-a716-446655440205', 'clean-press', 'Salwar Kameez', 'salwar-kameez', 'Complete salwar suit with dupatta', 89.00, 'item', 0.500, 6, 'laundry', TRUE),
    ('550e8400-e29b-41d4-a716-446655440206', 'clean-press', 'Saree (Cotton)', 'saree-cotton', 'Cotton saree with pleating', 79.00, 'item', 0.450, 7, 'laundry', TRUE),
    ('550e8400-e29b-41d4-a716-446655440207', 'clean-press', 'Saree (Silk)', 'saree-silk', 'Silk saree with careful pressing', 129.00, 'item', 0.550, 8, 'laundry', TRUE),
    ('550e8400-e29b-41d4-a716-446655440208', 'clean-press', 'Dress', 'dress', 'Western dress with pressing', 55.00, 'item', 0.300, 9, 'laundry', TRUE),
    ('550e8400-e29b-41d4-a716-446655440209', 'clean-press', 'Blouse', 'blouse', 'Saree blouse', 35.00, 'item', 0.150, 10, 'laundry', TRUE),
    ('550e8400-e29b-41d4-a716-44665544020a', 'clean-press', 'Dupatta', 'dupatta', 'Dupatta with light starch', 30.00, 'item', 0.150, 11, 'laundry', TRUE),
    ('550e8400-e29b-41d4-a716-44665544020b', 'clean-press', 'Table Cloth', 'table-cloth', 'Dining table cloth', 55.00, 'item', 0.400, 12, 'laundry', TRUE),
    ('550e8400-e29b-41d4-a716-44665544020c', 'clean-press', 'Napkins (Set of 6)', 'napkins-set', 'Cloth napkins set', 45.00, 'item', 0.200, 13, 'laundry', TRUE),
    ('550e8400-e29b-41d4-a716-44665544020d', 'clean-press', 'Blazer', 'blazer', 'Blazer with steam pressing', 99.00, 'item', 0.700, 14, 'laundry', TRUE)
ON CONFLICT (service_slug, slug) DO UPDATE SET
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    price = EXCLUDED.price,
    pricing_unit = EXCLUDED.pricing_unit,
    typical_weight = EXCLUDED.typical_weight,
    display_order = EXCLUDED.display_order,
    is_active = EXCLUDED.is_active,
    updated_at = CURRENT_TIMESTAMP;

-- =====================================================
-- SEED: Products for Shoe Care (INR Pricing)
-- Professional shoe cleaning: ₹149-699 per pair
-- =====================================================

INSERT INTO laundry_service_products (id, service_slug, name, slug, description, price, pricing_unit, typical_weight, requires_special_care, special_care_fee, display_order, category_slug, is_active)
VALUES
    ('550e8400-e29b-41d4-a716-446655440300', 'shoe-care', 'Sneakers (Canvas)', 'sneakers-canvas', 'Canvas sports sneakers', 249.00, 'pair', 0.600, FALSE, 0.00, 1, 'laundry', TRUE),
    ('550e8400-e29b-41d4-a716-446655440301', 'shoe-care', 'Sneakers (Leather)', 'sneakers-leather', 'Leather sneakers with conditioning', 349.00, 'pair', 0.700, TRUE, 75.00, 2, 'laundry', TRUE),
    ('550e8400-e29b-41d4-a716-446655440302', 'shoe-care', 'Formal Shoes (Leather)', 'formal-leather', 'Leather formal shoes with polishing', 399.00, 'pair', 0.800, TRUE, 100.00, 3, 'laundry', TRUE),
    ('550e8400-e29b-41d4-a716-446655440303', 'shoe-care', 'Boots (Ankle)', 'boots-ankle', 'Ankle boots cleaning', 449.00, 'pair', 0.900, TRUE, 100.00, 4, 'laundry', TRUE),
    ('550e8400-e29b-41d4-a716-446655440304', 'shoe-care', 'Boots (Long)', 'boots-long', 'Knee-high or long boots', 599.00, 'pair', 1.200, TRUE, 150.00, 5, 'laundry', TRUE),
    ('550e8400-e29b-41d4-a716-446655440305', 'shoe-care', 'Sandals (Leather)', 'sandals-leather', 'Leather sandals cleaning', 199.00, 'pair', 0.400, TRUE, 50.00, 6, 'laundry', TRUE),
    ('550e8400-e29b-41d4-a716-446655440306', 'shoe-care', 'Sandals (Fabric)', 'sandals-fabric', 'Fabric or synthetic sandals', 149.00, 'pair', 0.300, FALSE, 0.00, 7, 'laundry', TRUE),
    ('550e8400-e29b-41d4-a716-446655440307', 'shoe-care', 'Heels', 'heels', 'High heels cleaning and care', 349.00, 'pair', 0.500, TRUE, 75.00, 8, 'laundry', TRUE),
    ('550e8400-e29b-41d4-a716-446655440308', 'shoe-care', 'Loafers', 'loafers', 'Leather loafers with conditioning', 349.00, 'pair', 0.600, TRUE, 75.00, 9, 'laundry', TRUE),
    ('550e8400-e29b-41d4-a716-446655440309', 'shoe-care', 'Sports Shoes', 'sports-shoes', 'Running/training shoes deep clean', 299.00, 'pair', 0.650, FALSE, 0.00, 10, 'laundry', TRUE),
    ('550e8400-e29b-41d4-a716-44665544030a', 'shoe-care', 'Mojari/Juttis', 'mojari-juttis', 'Traditional Indian footwear', 249.00, 'pair', 0.350, TRUE, 75.00, 11, 'laundry', TRUE),
    ('550e8400-e29b-41d4-a716-44665544030b', 'shoe-care', 'Kolhapuri Chappals', 'kolhapuri', 'Leather kolhapuri cleaning', 199.00, 'pair', 0.400, TRUE, 50.00, 12, 'laundry', TRUE),
    ('550e8400-e29b-41d4-a716-44665544030c', 'shoe-care', 'Suede Shoes', 'suede-shoes', 'Suede shoes specialized cleaning', 499.00, 'pair', 0.600, TRUE, 150.00, 13, 'laundry', TRUE),
    ('550e8400-e29b-41d4-a716-44665544030d', 'shoe-care', 'White Shoes (Any)', 'white-shoes', 'White shoes restoration and cleaning', 399.00, 'pair', 0.600, TRUE, 100.00, 14, 'laundry', TRUE)
ON CONFLICT (service_slug, slug) DO UPDATE SET
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    price = EXCLUDED.price,
    pricing_unit = EXCLUDED.pricing_unit,
    typical_weight = EXCLUDED.typical_weight,
    requires_special_care = EXCLUDED.requires_special_care,
    special_care_fee = EXCLUDED.special_care_fee,
    display_order = EXCLUDED.display_order,
    is_active = EXCLUDED.is_active,
    updated_at = CURRENT_TIMESTAMP;

-- =====================================================
-- SEED: Products for Home Linens (INR Pricing)
-- Home textile cleaning: ₹29-599 per item
-- =====================================================

INSERT INTO laundry_service_products (id, service_slug, name, slug, description, price, pricing_unit, typical_weight, display_order, category_slug, is_active)
VALUES
    ('550e8400-e29b-41d4-a716-446655440400', 'home-linens', 'Bed Sheet (Single)', 'bed-sheet-single', 'Single bed sheet wash and fold', 69.00, 'item', 0.500, 1, 'laundry', TRUE),
    ('550e8400-e29b-41d4-a716-446655440401', 'home-linens', 'Bed Sheet (Double)', 'bed-sheet-double', 'Double bed sheet wash and fold', 99.00, 'item', 0.800, 2, 'laundry', TRUE),
    ('550e8400-e29b-41d4-a716-446655440402', 'home-linens', 'Bed Sheet (King)', 'bed-sheet-king', 'King size bed sheet', 129.00, 'item', 1.000, 3, 'laundry', TRUE),
    ('550e8400-e29b-41d4-a716-446655440403', 'home-linens', 'Pillow Cover', 'pillow-cover', 'Standard pillow cover', 29.00, 'item', 0.100, 4, 'laundry', TRUE),
    ('550e8400-e29b-41d4-a716-446655440404', 'home-linens', 'Duvet Cover (Single)', 'duvet-cover-single', 'Single duvet/quilt cover', 149.00, 'item', 0.700, 5, 'laundry', TRUE),
    ('550e8400-e29b-41d4-a716-446655440405', 'home-linens', 'Duvet Cover (Double)', 'duvet-cover-double', 'Double duvet/quilt cover', 199.00, 'item', 1.000, 6, 'laundry', TRUE),
    ('550e8400-e29b-41d4-a716-446655440406', 'home-linens', 'Comforter (Single)', 'comforter-single', 'Single comforter/razai', 349.00, 'item', 2.000, 7, 'laundry', TRUE),
    ('550e8400-e29b-41d4-a716-446655440407', 'home-linens', 'Comforter (Double)', 'comforter-double', 'Double comforter/razai', 449.00, 'item', 3.000, 8, 'laundry', TRUE),
    ('550e8400-e29b-41d4-a716-446655440408', 'home-linens', 'Blanket (Light)', 'blanket-light', 'Light cotton blanket/kambal', 199.00, 'item', 1.500, 9, 'laundry', TRUE),
    ('550e8400-e29b-41d4-a716-446655440409', 'home-linens', 'Blanket (Heavy)', 'blanket-heavy', 'Heavy wool blanket', 299.00, 'item', 2.500, 10, 'laundry', TRUE),
    ('550e8400-e29b-41d4-a716-44665544040a', 'home-linens', 'Curtains (Per Panel)', 'curtains-panel', 'Window curtain per panel', 129.00, 'item', 0.800, 11, 'laundry', TRUE),
    ('550e8400-e29b-41d4-a716-44665544040b', 'home-linens', 'Curtains (Heavy/Blackout)', 'curtains-heavy', 'Heavy or blackout curtains per panel', 179.00, 'item', 1.200, 12, 'laundry', TRUE),
    ('550e8400-e29b-41d4-a716-44665544040c', 'home-linens', 'Sofa Cover (Single Seat)', 'sofa-cover-single', 'Single seater sofa cover', 149.00, 'item', 0.800, 13, 'laundry', TRUE),
    ('550e8400-e29b-41d4-a716-44665544040d', 'home-linens', 'Sofa Cover (3 Seater)', 'sofa-cover-3seater', 'Three seater sofa cover', 349.00, 'item', 2.000, 14, 'laundry', TRUE),
    ('550e8400-e29b-41d4-a716-44665544040e', 'home-linens', 'Table Cloth (4 Seater)', 'table-cloth-4', '4 seater dining table cloth', 89.00, 'item', 0.400, 15, 'laundry', TRUE),
    ('550e8400-e29b-41d4-a716-44665544040f', 'home-linens', 'Table Cloth (6 Seater)', 'table-cloth-6', '6 seater dining table cloth', 119.00, 'item', 0.600, 16, 'laundry', TRUE),
    ('550e8400-e29b-41d4-a716-446655440410', 'home-linens', 'Bath Towel (Large)', 'bath-towel-large', 'Large bath towel', 59.00, 'item', 0.500, 17, 'laundry', TRUE),
    ('550e8400-e29b-41d4-a716-446655440411', 'home-linens', 'Hand Towel', 'hand-towel', 'Small hand towel', 29.00, 'item', 0.200, 18, 'laundry', TRUE),
    ('550e8400-e29b-41d4-a716-446655440412', 'home-linens', 'Floor Mat', 'floor-mat', 'Bathroom or bedroom floor mat', 49.00, 'item', 0.400, 19, 'laundry', TRUE),
    ('550e8400-e29b-41d4-a716-446655440413', 'home-linens', 'Mattress Cover', 'mattress-cover', 'Mattress protector cover', 199.00, 'item', 1.000, 20, 'laundry', TRUE)
ON CONFLICT (service_slug, slug) DO UPDATE SET
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    price = EXCLUDED.price,
    pricing_unit = EXCLUDED.pricing_unit,
    typical_weight = EXCLUDED.typical_weight,
    display_order = EXCLUDED.display_order,
    is_active = EXCLUDED.is_active,
    updated_at = CURRENT_TIMESTAMP;

-- =====================================================
-- SEED: Products for Carpet Cleaning (INR Pricing)
-- Professional carpet cleaning: ₹299-2499 per item
-- =====================================================

INSERT INTO laundry_service_products (id, service_slug, name, slug, description, price, pricing_unit, typical_weight, requires_special_care, special_care_fee, display_order, category_slug, is_active)
VALUES
    ('550e8400-e29b-41d4-a716-446655440500', 'carpet-cleaning', 'Small Rug (3x5 ft)', 'small-rug', 'Small area rug up to 3x5 feet', 499.00, 'item', 3.000, FALSE, 0.00, 1, 'laundry', TRUE),
    ('550e8400-e29b-41d4-a716-446655440501', 'carpet-cleaning', 'Medium Rug (5x7 ft)', 'medium-rug', 'Medium rug 5x7 feet', 799.00, 'item', 6.000, FALSE, 0.00, 2, 'laundry', TRUE),
    ('550e8400-e29b-41d4-a716-446655440502', 'carpet-cleaning', 'Large Rug (8x10 ft)', 'large-rug', 'Large area rug 8x10 feet', 1299.00, 'item', 12.000, FALSE, 0.00, 3, 'laundry', TRUE),
    ('550e8400-e29b-41d4-a716-446655440503', 'carpet-cleaning', 'Extra Large Rug (10x12 ft)', 'xl-rug', 'Extra large rug 10x12 feet or bigger', 1799.00, 'item', 18.000, TRUE, 300.00, 4, 'laundry', TRUE),
    ('550e8400-e29b-41d4-a716-446655440504', 'carpet-cleaning', 'Runner (2.5x8 ft)', 'runner', 'Hallway runner carpet', 599.00, 'item', 4.000, FALSE, 0.00, 5, 'laundry', TRUE),
    ('550e8400-e29b-41d4-a716-446655440505', 'carpet-cleaning', 'Prayer Mat/Janamaz', 'prayer-mat', 'Small prayer rug', 299.00, 'item', 1.500, FALSE, 0.00, 6, 'laundry', TRUE),
    ('550e8400-e29b-41d4-a716-446655440506', 'carpet-cleaning', 'Door Mat', 'door-mat', 'Entrance or door mat', 199.00, 'item', 1.000, FALSE, 0.00, 7, 'laundry', TRUE),
    ('550e8400-e29b-41d4-a716-446655440507', 'carpet-cleaning', 'Handmade Carpet (Small)', 'handmade-small', 'Handmade or antique carpet up to 4x6 ft', 999.00, 'item', 5.000, TRUE, 250.00, 8, 'laundry', TRUE),
    ('550e8400-e29b-41d4-a716-446655440508', 'carpet-cleaning', 'Handmade Carpet (Large)', 'handmade-large', 'Handmade or antique carpet 6x9 ft+', 1999.00, 'item', 12.000, TRUE, 500.00, 9, 'laundry', TRUE),
    ('550e8400-e29b-41d4-a716-446655440509', 'carpet-cleaning', 'Kashmiri Carpet (Small)', 'kashmiri-small', 'Kashmiri silk/wool carpet small', 1499.00, 'item', 4.000, TRUE, 400.00, 10, 'laundry', TRUE),
    ('550e8400-e29b-41d4-a716-44665544050a', 'carpet-cleaning', 'Kashmiri Carpet (Large)', 'kashmiri-large', 'Kashmiri silk/wool carpet large', 2499.00, 'item', 10.000, TRUE, 600.00, 11, 'laundry', TRUE),
    ('550e8400-e29b-41d4-a716-44665544050b', 'carpet-cleaning', 'Bathroom Rug', 'bathroom-rug', 'Bathroom floor rug/mat', 249.00, 'item', 1.500, FALSE, 0.00, 12, 'laundry', TRUE),
    ('550e8400-e29b-41d4-a716-44665544050c', 'carpet-cleaning', 'Kids Play Mat', 'play-mat', 'Children play mat or rug', 399.00, 'item', 2.500, FALSE, 0.00, 13, 'laundry', TRUE)
ON CONFLICT (service_slug, slug) DO UPDATE SET
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    price = EXCLUDED.price,
    pricing_unit = EXCLUDED.pricing_unit,
    typical_weight = EXCLUDED.typical_weight,
    requires_special_care = EXCLUDED.requires_special_care,
    special_care_fee = EXCLUDED.special_care_fee,
    display_order = EXCLUDED.display_order,
    is_active = EXCLUDED.is_active,
    updated_at = CURRENT_TIMESTAMP;

-- =====================================================
-- SEED: Products for Leather Care (INR Pricing)
-- Leather cleaning and conditioning: ₹399-2499 per item
-- =====================================================

INSERT INTO laundry_service_products (id, service_slug, name, slug, description, price, pricing_unit, typical_weight, requires_special_care, special_care_fee, display_order, category_slug, is_active)
VALUES
    ('550e8400-e29b-41d4-a716-446655440600', 'leather-care', 'Leather Jacket (Regular)', 'leather-jacket', 'Full leather jacket cleaning and conditioning', 899.00, 'item', 1.500, TRUE, 200.00, 1, 'laundry', TRUE),
    ('550e8400-e29b-41d4-a716-446655440601', 'leather-care', 'Leather Jacket (Premium)', 'leather-jacket-premium', 'Premium leather jacket with restoration', 1299.00, 'item', 1.800, TRUE, 300.00, 2, 'laundry', TRUE),
    ('550e8400-e29b-41d4-a716-446655440602', 'leather-care', 'Leather Pants', 'leather-pants', 'Leather trousers cleaning', 699.00, 'item', 1.000, TRUE, 150.00, 3, 'laundry', TRUE),
    ('550e8400-e29b-41d4-a716-446655440603', 'leather-care', 'Leather Skirt', 'leather-skirt', 'Leather skirt cleaning', 599.00, 'item', 0.600, TRUE, 100.00, 4, 'laundry', TRUE),
    ('550e8400-e29b-41d4-a716-446655440604', 'leather-care', 'Leather Bag (Small)', 'leather-bag-small', 'Small leather handbag or clutch', 499.00, 'item', 0.400, TRUE, 100.00, 5, 'laundry', TRUE),
    ('550e8400-e29b-41d4-a716-446655440605', 'leather-care', 'Leather Bag (Medium)', 'leather-bag-medium', 'Medium leather handbag', 699.00, 'item', 0.700, TRUE, 150.00, 6, 'laundry', TRUE),
    ('550e8400-e29b-41d4-a716-446655440606', 'leather-care', 'Leather Bag (Large)', 'leather-bag-large', 'Large leather bag or briefcase', 899.00, 'item', 1.200, TRUE, 200.00, 7, 'laundry', TRUE),
    ('550e8400-e29b-41d4-a716-446655440607', 'leather-care', 'Leather Wallet', 'leather-wallet', 'Leather wallet cleaning and conditioning', 299.00, 'item', 0.150, TRUE, 50.00, 8, 'laundry', TRUE),
    ('550e8400-e29b-41d4-a716-446655440608', 'leather-care', 'Leather Belt', 'leather-belt', 'Leather belt cleaning', 199.00, 'item', 0.200, TRUE, 50.00, 9, 'laundry', TRUE),
    ('550e8400-e29b-41d4-a716-446655440609', 'leather-care', 'Leather Shoes', 'leather-shoes', 'Premium leather shoes deep conditioning', 599.00, 'pair', 0.800, TRUE, 150.00, 10, 'laundry', TRUE),
    ('550e8400-e29b-41d4-a716-44665544060a', 'leather-care', 'Leather Gloves', 'leather-gloves', 'Pair of leather gloves', 349.00, 'pair', 0.200, TRUE, 75.00, 11, 'laundry', TRUE),
    ('550e8400-e29b-41d4-a716-44665544060b', 'leather-care', 'Leather Sofa Cushion', 'leather-sofa-cushion', 'Single leather sofa cushion cover', 799.00, 'item', 1.500, TRUE, 200.00, 12, 'laundry', TRUE),
    ('550e8400-e29b-41d4-a716-44665544060c', 'leather-care', 'Leather Chair', 'leather-chair', 'Leather office or dining chair', 1199.00, 'item', 4.000, TRUE, 300.00, 13, 'laundry', TRUE),
    ('550e8400-e29b-41d4-a716-44665544060d', 'leather-care', 'Suede Jacket', 'suede-jacket', 'Suede jacket specialized cleaning', 999.00, 'item', 1.200, TRUE, 250.00, 14, 'laundry', TRUE),
    ('550e8400-e29b-41d4-a716-44665544060e', 'leather-care', 'Suede Bag', 'suede-bag', 'Suede handbag cleaning', 599.00, 'item', 0.600, TRUE, 150.00, 15, 'laundry', TRUE),
    ('550e8400-e29b-41d4-a716-44665544060f', 'leather-care', 'Leather Car Seat (Per Seat)', 'leather-car-seat', 'Leather car seat cleaning per seat', 999.00, 'item', 5.000, TRUE, 250.00, 16, 'laundry', TRUE)
ON CONFLICT (service_slug, slug) DO UPDATE SET
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    price = EXCLUDED.price,
    pricing_unit = EXCLUDED.pricing_unit,
    typical_weight = EXCLUDED.typical_weight,
    requires_special_care = EXCLUDED.requires_special_care,
    special_care_fee = EXCLUDED.special_care_fee,
    display_order = EXCLUDED.display_order,
    is_active = EXCLUDED.is_active,
    updated_at = CURRENT_TIMESTAMP;

-- =====================================================
-- SEED: Products for Bag Care (INR Pricing)
-- Bag cleaning and care: ₹199-799 per item
-- =====================================================

INSERT INTO laundry_service_products (id, service_slug, name, slug, description, price, pricing_unit, typical_weight, requires_special_care, special_care_fee, display_order, category_slug, is_active)
VALUES
    ('550e8400-e29b-41d4-a716-446655440700', 'bag', 'Handbag (Fabric)', 'handbag-fabric', 'Regular fabric or canvas handbag', 299.00, 'item', 0.500, FALSE, 0.00, 1, 'laundry', TRUE),
    ('550e8400-e29b-41d4-a716-446655440701', 'bag', 'Handbag (Leather)', 'handbag-leather', 'Leather handbag with conditioning', 499.00, 'item', 0.700, TRUE, 100.00, 2, 'laundry', TRUE),
    ('550e8400-e29b-41d4-a716-446655440702', 'bag', 'Designer Handbag', 'designer-handbag', 'Premium designer bag specialized care', 799.00, 'item', 0.800, TRUE, 200.00, 3, 'laundry', TRUE),
    ('550e8400-e29b-41d4-a716-446655440703', 'bag', 'Backpack (Small)', 'backpack-small', 'Small backpack or daypack', 249.00, 'item', 0.600, FALSE, 0.00, 4, 'laundry', TRUE),
    ('550e8400-e29b-41d4-a716-446655440704', 'bag', 'Backpack (Large)', 'backpack-large', 'Large backpack or hiking bag', 399.00, 'item', 1.200, FALSE, 0.00, 5, 'laundry', TRUE),
    ('550e8400-e29b-41d4-a716-446655440705', 'bag', 'Laptop Bag', 'laptop-bag', 'Padded laptop bag or sleeve', 349.00, 'item', 0.800, FALSE, 0.00, 6, 'laundry', TRUE),
    ('550e8400-e29b-41d4-a716-446655440706', 'bag', 'Office Bag/Briefcase', 'office-bag', 'Professional office bag or briefcase', 449.00, 'item', 1.000, TRUE, 75.00, 7, 'laundry', TRUE),
    ('550e8400-e29b-41d4-a716-446655440707', 'bag', 'Travel Duffel (Small)', 'duffel-small', 'Small travel duffel bag', 399.00, 'item', 1.000, FALSE, 0.00, 8, 'laundry', TRUE),
    ('550e8400-e29b-41d4-a716-446655440708', 'bag', 'Travel Duffel (Large)', 'duffel-large', 'Large travel duffel bag', 549.00, 'item', 1.500, FALSE, 0.00, 9, 'laundry', TRUE),
    ('550e8400-e29b-41d4-a716-446655440709', 'bag', 'Suitcase (Cabin)', 'suitcase-cabin', 'Cabin size suitcase exterior cleaning', 399.00, 'item', 3.000, FALSE, 0.00, 10, 'laundry', TRUE),
    ('550e8400-e29b-41d4-a716-44665544070a', 'bag', 'Suitcase (Large)', 'suitcase-large', 'Large suitcase exterior cleaning', 549.00, 'item', 5.000, FALSE, 0.00, 11, 'laundry', TRUE),
    ('550e8400-e29b-41d4-a716-44665544070b', 'bag', 'Gym Bag', 'gym-bag', 'Sports or gym bag', 299.00, 'item', 0.700, FALSE, 0.00, 12, 'laundry', TRUE),
    ('550e8400-e29b-41d4-a716-44665544070c', 'bag', 'School Bag', 'school-bag', 'Children school bag', 249.00, 'item', 0.600, FALSE, 0.00, 13, 'laundry', TRUE),
    ('550e8400-e29b-41d4-a716-44665544070d', 'bag', 'Tote Bag (Canvas)', 'tote-canvas', 'Canvas tote or shopping bag', 199.00, 'item', 0.400, FALSE, 0.00, 14, 'laundry', TRUE),
    ('550e8400-e29b-41d4-a716-44665544070e', 'bag', 'Clutch/Pouch', 'clutch', 'Small clutch or pouch', 199.00, 'item', 0.200, FALSE, 0.00, 15, 'laundry', TRUE),
    ('550e8400-e29b-41d4-a716-44665544070f', 'bag', 'Camera Bag', 'camera-bag', 'Padded camera bag', 399.00, 'item', 0.800, TRUE, 75.00, 16, 'laundry', TRUE)
ON CONFLICT (service_slug, slug) DO UPDATE SET
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    price = EXCLUDED.price,
    pricing_unit = EXCLUDED.pricing_unit,
    typical_weight = EXCLUDED.typical_weight,
    requires_special_care = EXCLUDED.requires_special_care,
    special_care_fee = EXCLUDED.special_care_fee,
    display_order = EXCLUDED.display_order,
    is_active = EXCLUDED.is_active,
    updated_at = CURRENT_TIMESTAMP;

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
-- VIEW: Services with Products (INR Pricing)
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
    COALESCE(MIN(lsp.price), lsc.base_price) AS min_price,
    COALESCE(MAX(lsp.price), lsc.base_price) AS max_price,
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
WHERE lsc.is_active = TRUE AND lsc.deleted_at IS NULL
GROUP BY lsc.id, lsc.slug, lsc.title, lsc.description, lsc.color_code, 
         lsc.base_price, lsc.pricing_unit, lsc.turnaround_hours, 
         lsc.express_fee, lsc.express_hours, lsc.is_active
ORDER BY lsc.display_order;

-- =====================================================
-- VIEW: Price Summary per Service
-- =====================================================

CREATE OR REPLACE VIEW v_laundry_price_summary AS
SELECT 
    lsc.slug AS service_slug,
    lsc.title AS service_title,
    lsc.base_price,
    lsc.pricing_unit,
    lsc.express_fee,
    COUNT(lsp.id) AS total_products,
    MIN(lsp.price) AS min_product_price,
    MAX(lsp.price) AS max_product_price,
    ROUND(AVG(lsp.price), 2) AS avg_product_price,
    STRING_AGG(DISTINCT lsp.pricing_unit, ', ') AS product_pricing_units
FROM laundry_service_catalog lsc
LEFT JOIN laundry_service_products lsp ON lsc.slug = lsp.service_slug AND lsp.is_active = TRUE
WHERE lsc.is_active = TRUE AND lsc.deleted_at IS NULL
GROUP BY lsc.slug, lsc.title, lsc.base_price, lsc.pricing_unit, lsc.express_fee, lsc.display_order
ORDER BY lsc.display_order;

-- =====================================================
-- COMMENTS (Updated for INR)
-- =====================================================

COMMENT ON TABLE laundry_service_catalog IS 'Master catalog of laundry services available - Prices in INR';
COMMENT ON TABLE laundry_service_products IS 'Available products/items for each laundry service - ALL products have explicit prices in INR';
COMMENT ON TABLE laundry_order_items IS 'Individual items in a laundry order with QR tracking';
COMMENT ON TABLE laundry_pickups IS 'Pickup schedules and tracking';
COMMENT ON TABLE laundry_deliveries IS 'Delivery schedules and tracking';
COMMENT ON TABLE laundry_issues IS 'Customer complaints and issue tracking';
COMMENT ON VIEW v_laundry_services_with_products IS 'Services with their product listings, counts, and price ranges in INR';
COMMENT ON VIEW v_laundry_price_summary IS 'Quick price summary per service category in INR';

COMMENT ON COLUMN laundry_service_catalog.base_price IS 'Base price in INR per unit (kg or item)';
COMMENT ON COLUMN laundry_service_catalog.express_fee IS 'Additional express service fee in INR';
COMMENT ON COLUMN laundry_service_products.price IS 'Explicit price per item in INR - REQUIRED for all products';
COMMENT ON COLUMN laundry_service_products.typical_weight IS 'Average weight in kg - used for reference and bulk calculations';
COMMENT ON COLUMN laundry_service_products.special_care_fee IS 'Additional fee in INR for items requiring special care';
COMMENT ON COLUMN laundry_order_items.product_slug IS 'Links to the specific product in laundry_service_products';
COMMENT ON COLUMN laundry_order_items.qr_code IS 'Unique QR code for item tracking throughout the process';
COMMENT ON COLUMN laundry_order_items.price IS 'Final price in INR charged for this item';

-- =====================================================
-- USEFUL QUERIES FOR VERIFICATION
-- =====================================================

-- Check price ranges per service
-- SELECT * FROM v_laundry_price_summary;

-- Get all products for a specific service
-- SELECT * FROM laundry_service_products WHERE service_slug = 'wash-fold' ORDER BY display_order;

-- Get services with product counts
-- SELECT slug, title, base_price, (SELECT COUNT(*) FROM laundry_service_products WHERE service_slug = slug) as products FROM laundry_service_catalog;