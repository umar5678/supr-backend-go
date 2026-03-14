-- Rollback document linking indices and comments
DROP INDEX IF EXISTS idx_documents_driver_id;
DROP INDEX IF EXISTS idx_documents_service_provider_id;
