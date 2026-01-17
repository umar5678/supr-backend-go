-- Add driver_fare and rider_fare columns
-- Both driver and rider see the same amount (with promo discount applied if used)

ALTER TABLE rides
ADD COLUMN driver_fare DECIMAL(10, 2),
ADD COLUMN rider_fare DECIMAL(10, 2);

-- Add comments for clarity
COMMENT ON COLUMN rides.driver_fare IS 'Amount the driver earns (includes promo discount if applicable)';
COMMENT ON COLUMN rides.rider_fare IS 'Amount the rider pays (includes promo discount if applicable)';

