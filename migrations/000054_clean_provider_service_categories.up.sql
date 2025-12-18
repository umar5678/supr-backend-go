-- Clean Home Services Schema - Part 5: Provider Service Categories (Categories a provider offers)

CREATE TABLE IF NOT EXISTS provider_service_categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    provider_id UUID NOT NULL REFERENCES service_provider_profiles(id) ON DELETE CASCADE,
    category_slug VARCHAR(255) NOT NULL,  -- Slug reference to category
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    -- Ensure provider doesn't add same category twice
    UNIQUE(provider_id, category_slug)
);

CREATE INDEX idx_provider_service_categories_provider_id ON provider_service_categories(provider_id);
CREATE INDEX idx_provider_service_categories_category_slug ON provider_service_categories(category_slug);
CREATE INDEX idx_provider_service_categories_active ON provider_service_categories(is_active);
