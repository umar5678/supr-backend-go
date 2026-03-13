-- Rollback: Rename image_kit_file_id back to imagekit_file_id
ALTER TABLE documents RENAME COLUMN image_kit_file_id TO imagekit_file_id;

-- Rename image_kit_file_path back to imagekit_file_path
ALTER TABLE documents RENAME COLUMN image_kit_file_path TO imagekit_file_path;

-- Rename the index back
DROP INDEX IF EXISTS idx_documents_image_kit_file_id;
CREATE INDEX IF NOT EXISTS idx_documents_imagekit_file_id ON documents(imagekit_file_id);
