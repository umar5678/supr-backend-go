# 🚗 Vehicle Types Insert Guide

## Quick Insert Commands

### Option 1: Using the Pre-made SQL File (Recommended)

```bash
# Connect to your database and run the SQL file
psql -U your_username -d your_database -f migrations/insert_vehicle_types.sql

# Or if you're on the server:
sudo -u postgres psql -d supr_db -f /var/www/go-backend/supr-backend-go/migrations/insert_vehicle_types.sql
```

### Option 2: Manual Insert (One by One)

```sql
-- Connect to your database first
psql -U your_username -d your_database

-- Then run these inserts:

INSERT INTO vehicle_types (name, display_name, base_fare, per_km_rate, per_minute_rate, booking_fee, capacity, description, is_active, icon_url) 
VALUES ('bike', 'Bike', 1.50, 0.50, 0.10, 0.25, 1, 'Fast and affordable bike service', TRUE, '/icons/bike.png');

INSERT INTO vehicle_types (name, display_name, base_fare, per_km_rate, per_minute_rate, booking_fee, capacity, description, is_active, icon_url) 
VALUES ('economy', 'Economy', 2.50, 0.75, 0.15, 0.50, 4, 'Budget-friendly ride with comfort', TRUE, '/icons/economy.png');

INSERT INTO vehicle_types (name, display_name, base_fare, per_km_rate, per_minute_rate, booking_fee, capacity, description, is_active, icon_url) 
VALUES ('premium', 'Premium', 4.00, 1.25, 0.25, 1.00, 4, 'Premium comfort ride with air conditioning', TRUE, '/icons/premium.png');

INSERT INTO vehicle_types (name, display_name, base_fare, per_km_rate, per_minute_rate, booking_fee, capacity, description, is_active, icon_url) 
VALUES ('suv', 'SUV', 5.00, 1.50, 0.30, 1.50, 6, 'Spacious SUV for larger groups', TRUE, '/icons/suv.png');

INSERT INTO vehicle_types (name, display_name, base_fare, per_km_rate, per_minute_rate, booking_fee, capacity, description, is_active, icon_url) 
VALUES ('luxury', 'Luxury', 8.00, 2.00, 0.50, 2.00, 4, 'Luxury ride experience with premium amenities', TRUE, '/icons/luxury.png');

INSERT INTO vehicle_types (name, display_name, base_fare, per_km_rate, per_minute_rate, booking_fee, capacity, description, is_active, icon_url) 
VALUES ('van', 'Van', 6.00, 1.75, 0.35, 1.75, 8, 'Large van for groups and luggage', TRUE, '/icons/van.png');

INSERT INTO vehicle_types (name, display_name, base_fare, per_km_rate, per_minute_rate, booking_fee, capacity, description, is_active, icon_url) 
VALUES ('delivery_van', 'Delivery Van', 3.00, 0.90, 0.20, 0.75, 2, 'Cargo delivery service', TRUE, '/icons/delivery_van.png');

INSERT INTO vehicle_types (name, display_name, base_fare, per_km_rate, per_minute_rate, booking_fee, capacity, description, is_active, icon_url) 
VALUES ('scooter', 'Scooter', 1.00, 0.40, 0.08, 0.20, 1, 'Quick and easy scooter rides', TRUE, '/icons/scooter.png');

INSERT INTO vehicle_types (name, display_name, base_fare, per_km_rate, per_minute_rate, booking_fee, capacity, description, is_active, icon_url) 
VALUES ('auto', 'Auto/Rickshaw', 1.80, 0.60, 0.12, 0.30, 3, 'Local auto-rickshaw service', TRUE, '/icons/auto.png');
```

## Pricing Breakdown

| Vehicle Type | Base Fare | Per KM | Per Minute | Booking Fee | Capacity |
|---|---|---|---|---|---|
| Bike | $1.50 | $0.50 | $0.10 | $0.25 | 1 |
| Economy | $2.50 | $0.75 | $0.15 | $0.50 | 4 |
| Premium | $4.00 | $1.25 | $0.25 | $1.00 | 4 |
| SUV | $5.00 | $1.50 | $0.30 | $1.50 | 6 |
| Luxury | $8.00 | $2.00 | $0.50 | $2.00 | 4 |
| Van | $6.00 | $1.75 | $0.35 | $1.75 | 8 |
| Delivery Van | $3.00 | $0.90 | $0.20 | $0.75 | 2 |
| Scooter | $1.00 | $0.40 | $0.08 | $0.20 | 1 |
| Auto/Rickshaw | $1.80 | $0.60 | $0.12 | $0.30 | 3 |

## Verify the Inserts

After running the INSERT statements, verify they were successful:

```sql
-- Count all vehicle types
SELECT COUNT(*) as total_types FROM vehicle_types;

-- List all vehicle types with pricing
SELECT id, name, display_name, base_fare, per_km_rate, per_minute_rate, capacity, is_active 
FROM vehicle_types 
ORDER BY base_fare;

-- Get details for a specific vehicle type
SELECT * FROM vehicle_types WHERE name = 'economy';
```

## On Hostinger Server

```bash
# SSH into your server
ssh root@srv990975

# Navigate to the backend
cd /var/www/go-backend/supr-backend-go

# Connect to PostgreSQL and run the insert file
sudo -u postgres psql -d supr_db -f migrations/insert_vehicle_types.sql

# Or run inserts directly:
sudo -u postgres psql -d supr_db -c "INSERT INTO vehicle_types (name, display_name, base_fare, per_km_rate, per_minute_rate, booking_fee, capacity, description, is_active) VALUES ('bike', 'Bike', 1.50, 0.50, 0.10, 0.25, 1, 'Fast and affordable bike service', TRUE);"
```

## Customize Pricing

Edit the `insert_vehicle_types.sql` file to adjust pricing before inserting:

- **base_fare**: Initial charge for booking
- **per_km_rate**: Charge per kilometer traveled
- **per_minute_rate**: Charge per minute of ride
- **booking_fee**: Additional booking fee
- **capacity**: How many passengers can fit

## Notes

- ✅ Using `ON CONFLICT (name) DO NOTHING` prevents duplicate key errors if you run the script multiple times
- ✅ All vehicle types are set to `is_active = TRUE` by default
- ✅ You can modify the `icon_url` to point to your actual icon files
- ✅ Capacity is important for the algorithm to match riders with appropriate vehicles
- ✅ Pricing is in USD - adjust currency in the application as needed

## Testing in Your API

Once inserted, you can fetch vehicle types:

```bash
# Get all active vehicle types
curl http://localhost:8080/api/v1/vehicles

# Or with remote URL
curl https://api.pittapizzahusrev.be/go/api/v1/vehicles
```

---

**Ready to insert? Run:**
```bash
psql -U your_username -d your_database -f migrations/insert_vehicle_types.sql
```
