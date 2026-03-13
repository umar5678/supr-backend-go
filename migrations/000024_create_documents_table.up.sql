-- Create documents table
CREATE TABLE IF NOT EXISTS documents (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    driver_id UUID REFERENCES driver_profiles(id) ON DELETE CASCADE,
    service_provider_id UUID REFERENCES service_provider_profiles(id) ON DELETE CASCADE,
    
    -- Document details
    document_type VARCHAR(100) NOT NULL,
    file_name VARCHAR(255) NOT NULL,
    file_url VARCHAR(1000) NOT NULL,
    file_size BIGINT,
    mime_type VARCHAR(50),
    imagekit_file_id VARCHAR(255),
    imagekit_file_path VARCHAR(500),
    
    -- Verification status
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    verified_by UUID REFERENCES users(id) ON DELETE SET NULL,
    verified_at TIMESTAMP,
    rejection_reason TEXT,
    expiry_date TIMESTAMP,
    
    -- Metadata
    is_front BOOLEAN DEFAULT FALSE,
    metadata JSONB,
    
    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- Create indexes for better query performance
CREATE INDEX idx_documents_user_id ON documents(user_id);
CREATE INDEX idx_documents_driver_id ON documents(driver_id);
CREATE INDEX idx_documents_service_provider_id ON documents(service_provider_id);
CREATE INDEX idx_documents_document_type ON documents(document_type);
CREATE INDEX idx_documents_status ON documents(status);
CREATE INDEX idx_documents_created_at ON documents(created_at);
CREATE INDEX idx_documents_deleted_at ON documents(deleted_at);
-- Create index for imagekit_file_path if it doesn't exist
CREATE INDEX IF NOT EXISTS idx_documents_imagekit_file_id ON documents(imagekit_file_id);

-- Create document verification logs table
CREATE TABLE IF NOT EXISTS document_verification_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    document_id UUID NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    admin_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    
    action VARCHAR(50) NOT NULL,
    status VARCHAR(50) NOT NULL,
    comments TEXT,
    
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for verification logs
CREATE INDEX idx_document_verification_logs_document_id ON document_verification_logs(document_id);
CREATE INDEX idx_document_verification_logs_admin_id ON document_verification_logs(admin_id);
CREATE INDEX idx_document_verification_logs_created_at ON document_verification_logs(created_at);
