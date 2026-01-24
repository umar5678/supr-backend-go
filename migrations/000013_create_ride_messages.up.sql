-- Create ride_messages table
CREATE TABLE IF NOT EXISTS ride_messages (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    ride_id UUID NOT NULL,
    sender_id UUID NOT NULL,
    sender_type VARCHAR(50) NOT NULL CHECK (sender_type IN ('rider', 'driver')),
    message_type VARCHAR(50) DEFAULT 'text' NOT NULL CHECK (message_type IN ('text', 'location', 'status', 'system')),
    content TEXT NOT NULL,
    metadata JSONB,
    is_read BOOLEAN DEFAULT FALSE NOT NULL,
    read_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    deleted_at TIMESTAMP,
    
    -- Foreign Keys
    CONSTRAINT fk_ride_messages_ride 
        FOREIGN KEY (ride_id) REFERENCES rides(id) ON DELETE CASCADE,
    CONSTRAINT fk_ride_messages_sender 
        FOREIGN KEY (sender_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Create indexes for performance
CREATE INDEX idx_ride_messages_ride_id ON ride_messages(ride_id);
CREATE INDEX idx_ride_messages_sender_id ON ride_messages(sender_id);
CREATE INDEX idx_ride_messages_created_at ON ride_messages(created_at DESC);
CREATE INDEX idx_ride_messages_is_read_ride ON ride_messages(ride_id, is_read) 
    WHERE deleted_at IS NULL AND is_read = FALSE;
CREATE INDEX idx_ride_messages_deleted_at ON ride_messages(deleted_at) 
    WHERE deleted_at IS NULL;
