-- Add imagekit_file_path column to documents table
ALTER TABLE documents ADD COLUMN IF NOT EXISTS imagekit_file_path VARCHAR(500);

-- Create index for imagekit_file_path if it doesn't exist
CREATE INDEX IF NOT EXISTS idx_documents_imagekit_file_id ON documents(imagekit_file_id);
