-- Rename imagekit_file_id to image_kit_file_id to match GORM naming conventions
ALTER TABLE documents RENAME COLUMN imagekit_file_id TO image_kit_file_id;

-- Rename imagekit_file_path to image_kit_file_path to match GORM naming conventions
ALTER TABLE documents RENAME COLUMN imagekit_file_path TO image_kit_file_path;

-- Rename the indexes accordingly
DROP INDEX IF EXISTS idx_documents_imagekit_file_id;
CREATE INDEX IF NOT EXISTS idx_documents_image_kit_file_id ON documents(image_kit_file_id);
