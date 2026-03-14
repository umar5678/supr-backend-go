-- Link uploaded documents to driver and service provider profiles
-- When documents are uploaded, they are now automatically linked via driver_id or service_provider_id

-- These columns were already in the schema, but we ensure proper indexing for lookups
CREATE INDEX IF NOT EXISTS idx_documents_driver_id ON documents(driver_id) WHERE driver_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_documents_service_provider_id ON documents(service_provider_id) WHERE service_provider_id IS NOT NULL;

-- Add comment for documentation
COMMENT ON COLUMN documents.driver_id IS 'Links document to driver profile - auto-populated when driver uploads document';
COMMENT ON COLUMN documents.service_provider_id IS 'Links document to service provider profile - auto-populated when service provider uploads document';
