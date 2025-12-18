-- Clean Home Services Schema - Part 6: Provider Qualified Services (Which specific services a provider can do)

CREATE TABLE IF NOT EXISTS provider_qualified_services (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    provider_id UUID NOT NULL REFERENCES service_provider_profiles(id) ON DELETE CASCADE,
    service_id UUID NOT NULL REFERENCES services(id) ON DELETE CASCADE,
    is_available BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    -- Ensure provider doesn't have duplicate service assignments
    UNIQUE(provider_id, service_id)
);

CREATE INDEX idx_provider_qualified_services_provider_id ON provider_qualified_services(provider_id);
CREATE INDEX idx_provider_qualified_services_service_id ON provider_qualified_services(service_id);
CREATE INDEX idx_provider_qualified_services_provider_available ON provider_qualified_services(provider_id, is_available);
