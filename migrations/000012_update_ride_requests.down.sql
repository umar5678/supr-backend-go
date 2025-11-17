ALTER TABLE ride_requests
DROP COLUMN IF EXISTS updated_at;

ALTER TABLE ride_requests
DROP CONSTRAINT IF EXISTS ride_requests_status_check;

ALTER TABLE ride_requests
ADD CONSTRAINT ride_requests_status_check 
CHECK (status IN ('pending', 'expired', 'rejected', 'accepted'));
