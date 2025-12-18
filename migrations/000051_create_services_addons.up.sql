-- Clean Home Services Schema - Part 2: Services and Add-ons
-- This migration creates service offerings and add-on options

-- Services (individual service offerings)
CREATE TABLE IF NOT EXISTS services (
    id SERIAL PRIMARY KEY,
    category_id INT NOT NULL REFERENCES service_categories(id) ON DELETE CASCADE,
    tab_id INT NOT NULL REFERENCES service_tabs(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    image_url VARCHAR(500),
    base_price DECIMAL(10,2) NOT NULL,
    original_price DECIMAL(10,2),
    base_duration_minutes INT DEFAULT 60,
    is_active BOOLEAN DEFAULT TRUE,
    is_featured BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_services_category_id ON services(category_id);
CREATE INDEX idx_services_tab_id ON services(tab_id);
CREATE INDEX idx_services_is_active ON services(is_active);
CREATE INDEX idx_services_category_active ON services(category_id, is_active);

-- Add-ons (optional additions to services)
CREATE TABLE IF NOT EXISTS addons (
    id SERIAL PRIMARY KEY,
    service_id INT NOT NULL REFERENCES services(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    price DECIMAL(10,2) NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_addons_service_id ON addons(service_id);
CREATE INDEX idx_addons_is_active ON addons(is_active);
