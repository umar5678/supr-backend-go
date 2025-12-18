-- Clean Home Services Schema - Part 2: Services with Slug-Based Relations
-- Using slug-based identifiers instead of foreign keys

CREATE TABLE IF NOT EXISTS services (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(255) NOT NULL,
    long_title VARCHAR(500),
    service_slug VARCHAR(255) NOT NULL UNIQUE,
    category_slug VARCHAR(255) NOT NULL,  -- Slug reference to category, not FK
    description TEXT,
    long_description TEXT,
    highlights TEXT,
    whats_included TEXT[] DEFAULT '{}',
    terms_and_conditions TEXT[],
    banner_image VARCHAR(500),
    thumbnail VARCHAR(500),
    duration INTEGER,  -- in minutes
    is_frequent BOOLEAN DEFAULT FALSE,
    frequency VARCHAR(100),
    sort_order INTEGER DEFAULT 0,
    is_active BOOLEAN DEFAULT TRUE,
    is_available BOOLEAN DEFAULT TRUE,
    base_price DECIMAL(10,2),
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_services_service_slug ON services(service_slug);
CREATE INDEX idx_services_category_slug ON services(category_slug);
CREATE INDEX idx_services_is_active ON services(is_active);
CREATE INDEX idx_services_category_active ON services(category_slug, is_active);

-- Add-ons (optional service additions)
CREATE TABLE IF NOT EXISTS addons (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(255) NOT NULL,
    addon_slug VARCHAR(255) NOT NULL UNIQUE,
    category_slug VARCHAR(255) NOT NULL,  -- Slug reference to category
    description TEXT,
    price DECIMAL(10,2) NOT NULL,
    sort_order INTEGER DEFAULT 0,
    is_active BOOLEAN DEFAULT TRUE,
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_addons_addon_slug ON addons(addon_slug);
CREATE INDEX idx_addons_category_slug ON addons(category_slug);
CREATE INDEX idx_addons_is_active ON addons(is_active);
