-- Remove resolved status fields from admin_support_chats table
ALTER TABLE admin_support_chats
DROP COLUMN resolved_at,
DROP COLUMN is_resolved;

-- Drop the index
DROP INDEX IF EXISTS idx_admin_support_chats_is_resolved;
