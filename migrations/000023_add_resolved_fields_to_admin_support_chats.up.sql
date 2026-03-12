-- Add resolved status fields to admin_support_chats table
ALTER TABLE admin_support_chats
ADD COLUMN is_resolved BOOLEAN DEFAULT FALSE,
ADD COLUMN resolved_at TIMESTAMP WITH TIME ZONE;

-- Create index for is_resolved for faster queries
CREATE INDEX idx_admin_support_chats_is_resolved ON admin_support_chats(is_resolved);
