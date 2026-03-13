-- Drop imagekit_file_path column from documents table (rollback)
ALTER TABLE documents DROP COLUMN IF EXISTS imagekit_file_path;

-- Drop index if it exists
DROP INDEX IF EXISTS idx_documents_imagekit_file_id;
