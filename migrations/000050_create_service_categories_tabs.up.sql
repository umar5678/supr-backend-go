-- Clean Home Services Schema - Part 1: Service Categories and Tabs
-- This migration creates the base service catalog structure

-- Service Categories
CREATE TABLE IF NOT EXISTS service_categories (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    icon_url VARCHAR(500),
    banner_image VARCHAR(500),
    is_active BOOLEAN DEFAULT TRUE,
    sort_order INT DEFAULT 0,
    highlights JSONB DEFAULT '[]',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_service_categories_is_active ON service_categories(is_active);
CREATE INDEX idx_service_categories_sort_order ON service_categories(sort_order);

-- Service Tabs (subcategories within categories)
CREATE TABLE IF NOT EXISTS service_tabs (
    id SERIAL PRIMARY KEY,
    category_id INT NOT NULL REFERENCES service_categories(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    icon_url VARCHAR(500),
    banner_title VARCHAR(255),
    banner_description VARCHAR(500),
    banner_image VARCHAR(500),
    is_active BOOLEAN DEFAULT TRUE,
    sort_order INT DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_service_tabs_category_id ON service_tabs(category_id);
CREATE INDEX idx_service_tabs_is_active ON service_tabs(is_active);
CREATE INDEX idx_service_tabs_category_active ON service_tabs(category_id, is_active);
