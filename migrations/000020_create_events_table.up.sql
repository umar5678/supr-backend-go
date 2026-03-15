-- Create events table for notification event sourcing
CREATE TABLE IF NOT EXISTS events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_type VARCHAR(100) NOT NULL,
    source_module VARCHAR(50) NOT NULL,
    payload JSONB NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    retry_count INTEGER DEFAULT 0,
    failed_reason TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    published_at TIMESTAMP WITH TIME ZONE
);

-- Create index on event_type for faster lookups
CREATE INDEX IF NOT EXISTS idx_events_event_type ON events(event_type);

-- Create index on status for filtering pending/failed events
CREATE INDEX IF NOT EXISTS idx_events_status ON events(status);

-- Create index on created_at for time-based queries
CREATE INDEX IF NOT EXISTS idx_events_created_at ON events(created_at);

-- Create processed_events table for tracking which consumers have processed which events
CREATE TABLE IF NOT EXISTS processed_events (
    event_id UUID NOT NULL,
    consumer_group VARCHAR(100) NOT NULL,
    processed_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (event_id, consumer_group)
);

-- Create index on consumer_group for faster lookups
CREATE INDEX IF NOT EXISTS idx_processed_events_consumer_group ON processed_events(consumer_group);
