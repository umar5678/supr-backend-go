-- Create admin_support_chats table
CREATE TABLE admin_support_chats (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    
    -- Conversation threading
    conversation_id UUID NOT NULL,
    parent_message_id UUID,
    
    -- Message details
    sender_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    sender_role VARCHAR(50) NOT NULL,
    content TEXT NOT NULL,
    
    -- Metadata
    metadata JSONB,
    
    -- Status
    is_read BOOLEAN DEFAULT FALSE,
    read_by_admin_at TIMESTAMP,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Indexes
CREATE INDEX idx_admin_support_chats_conversation_id ON admin_support_chats(conversation_id);
CREATE INDEX idx_admin_support_chats_sender_id ON admin_support_chats(sender_id);
CREATE INDEX idx_admin_support_chats_parent_message_id ON admin_support_chats(parent_message_id);
CREATE INDEX idx_admin_support_chats_created_at ON admin_support_chats(created_at DESC);
CREATE INDEX idx_admin_support_chats_is_read ON admin_support_chats(is_read);
