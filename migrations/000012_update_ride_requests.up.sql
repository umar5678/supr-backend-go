ALTER TABLE ride_requests
ADD COLUMN updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP;

ALTER TABLE ride_requests
DROP CONSTRAINT IF EXISTS ride_requests_status_check;

ALTER TABLE ride_requests
ADD CONSTRAINT ride_requests_status_check 
CHECK (status IN ('pending', 'accepted', 'rejected', 'expired', 'cancelled'));
