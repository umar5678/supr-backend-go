-- Rollback: Remove imagekit_file_id column from documents table
DROP INDEX IF EXISTS idx_documents_imagekit_file_id;
ALTER TABLE documents DROP COLUMN IF EXISTS imagekit_file_id;
