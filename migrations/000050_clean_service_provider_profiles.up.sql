-- Clean Home Services Schema - Start Fresh with Slug-Based Relations
-- This replaces all messy migrations 000025-000041

-- Service Provider Profiles (Base table - providers)
CREATE TABLE IF NOT EXISTS service_provider_profiles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    
    -- Business Information
    business_name VARCHAR(255),
    description TEXT,
    service_category VARCHAR(100) NOT NULL,
    service_type VARCHAR(255) NOT NULL,
    
    -- Verification
    status VARCHAR(50) NOT NULL DEFAULT 'pending_approval',
    is_verified BOOLEAN DEFAULT FALSE,
    verification_docs JSONB DEFAULT '[]',
    
    -- Ratings
    rating DECIMAL(3,2) DEFAULT 0,
    total_reviews INTEGER DEFAULT 0,
    completed_jobs INTEGER DEFAULT 0,
    
    -- Availability
    is_available BOOLEAN DEFAULT TRUE,
    working_hours JSONB,
    service_areas JSONB DEFAULT '[]',
    
    -- Financial
    hourly_rate DECIMAL(10,2),
    currency VARCHAR(3) DEFAULT 'USD',
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_service_provider_profiles_user_id ON service_provider_profiles(user_id);
CREATE INDEX idx_service_provider_profiles_service_category ON service_provider_profiles(service_category);
CREATE INDEX idx_service_provider_profiles_is_verified ON service_provider_profiles(is_verified);
CREATE INDEX idx_service_provider_profiles_status ON service_provider_profiles(status);
CREATE INDEX idx_service_provider_profiles_deleted_at ON service_provider_profiles(deleted_at);
