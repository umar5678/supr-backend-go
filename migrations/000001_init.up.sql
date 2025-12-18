-- =====================================================
-- EXTENSIONS
-- =====================================================
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "postgis";

-- =====================================================
-- ENUM TYPES
-- =====================================================

CREATE TYPE user_role AS ENUM (
    'rider',
    'driver',
    'admin',
    'delivery_person',
    'service_provider',
    'handyman'
);

CREATE TYPE user_status AS ENUM (
    'active',
    'suspended',
    'banned',
    'pending_verification',
    'pending_approval'
);

CREATE TYPE wallet_type AS ENUM (
    'rider',
    'driver',
    'platform',
    'service_provider'
);

CREATE TYPE transaction_type AS ENUM (
    'credit',
    'debit',
    'refund',
    'hold',
    'release',
    'transfer'
);

CREATE TYPE transaction_status AS ENUM (
    'pending',
    'completed',
    'failed',
    'cancelled',
    'held',
    'released'
);

-- =====================================================
-- CORE TABLES
-- =====================================================

-- Users table
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255),
    phone VARCHAR(20),
    password VARCHAR(255),
    role user_role NOT NULL DEFAULT 'rider',
    status user_status NOT NULL DEFAULT 'active',
    profile_photo_url VARCHAR(500),
    last_login_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE UNIQUE INDEX idx_users_email ON users(email) WHERE email IS NOT NULL;
CREATE UNIQUE INDEX idx_users_phone ON users(phone) WHERE phone IS NOT NULL;
CREATE INDEX idx_users_deleted_at ON users(deleted_at);
CREATE INDEX idx_users_role ON users(role);
CREATE INDEX idx_users_status ON users(status);

-- =====================================================
-- WALLET TABLES
-- =====================================================

-- Wallets table
CREATE TABLE wallets (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL,
    wallet_type wallet_type NOT NULL,
    balance DECIMAL(12,2) NOT NULL DEFAULT 0.00,
    held_balance DECIMAL(12,2) NOT NULL DEFAULT 0.00,
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_wallets_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX idx_wallets_user_id ON wallets(user_id);
CREATE INDEX idx_wallets_wallet_type ON wallets(wallet_type);

-- Wallet transactions table
CREATE TABLE wallet_transactions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    wallet_id UUID NOT NULL,
    type transaction_type NOT NULL,
    amount DECIMAL(12,2) NOT NULL,
    balance_before DECIMAL(12,2) NOT NULL,
    balance_after DECIMAL(12,2) NOT NULL,
    status transaction_status NOT NULL DEFAULT 'pending',
    reference_type VARCHAR(50),
    reference_id VARCHAR(50) NOT NULL,
    description TEXT,
    metadata JSONB,
    processed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_wallet_transactions_wallet FOREIGN KEY (wallet_id) REFERENCES wallets(id) ON DELETE CASCADE
);

CREATE INDEX idx_wallet_transactions_wallet_id ON wallet_transactions(wallet_id);
CREATE INDEX idx_wallet_transactions_status ON wallet_transactions(status);
CREATE INDEX idx_wallet_transactions_reference ON wallet_transactions(reference_type, reference_id);
CREATE INDEX idx_wallet_transactions_created_at ON wallet_transactions(created_at);

-- Wallet holds table
CREATE TABLE wallet_holds (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    wallet_id UUID NOT NULL,
    amount DECIMAL(12,2) NOT NULL,
    reference_type VARCHAR(50) NOT NULL,
    reference_id UUID NOT NULL,
    status transaction_status NOT NULL DEFAULT 'held',
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    released_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_wallet_holds_wallet FOREIGN KEY (wallet_id) REFERENCES wallets(id) ON DELETE CASCADE
);

CREATE INDEX idx_wallet_holds_wallet_id ON wallet_holds(wallet_id);
CREATE INDEX idx_wallet_holds_status ON wallet_holds(status);
CREATE INDEX idx_wallet_holds_expires_at ON wallet_holds(expires_at);
CREATE INDEX idx_wallet_holds_reference ON wallet_holds(reference_type, reference_id);

-- =====================================================
-- RIDE-HAILING TABLES
-- =====================================================

-- Vehicle types table
CREATE TABLE vehicle_types (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(50) NOT NULL UNIQUE,
    display_name VARCHAR(100) NOT NULL,
    base_fare DECIMAL(10,2) NOT NULL,
    per_km_rate DECIMAL(10,2) NOT NULL,
    per_minute_rate DECIMAL(10,2) NOT NULL,
    booking_fee DECIMAL(10,2) NOT NULL DEFAULT 0.50,
    capacity INTEGER NOT NULL,
    description TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    icon_url VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_vehicle_types_is_active ON vehicle_types(is_active);
CREATE INDEX idx_vehicle_types_deleted_at ON vehicle_types(deleted_at);

-- Driver profiles table
CREATE TABLE driver_profiles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL UNIQUE,
    license_number VARCHAR(100) NOT NULL UNIQUE,
    status VARCHAR(50) DEFAULT 'offline',
    current_location GEOMETRY(Point, 4326),
    heading INTEGER DEFAULT 0,
    rating DECIMAL(3,2) DEFAULT 5.0,
    total_trips INTEGER DEFAULT 0,
    total_earnings DECIMAL(10,2) DEFAULT 0,
    acceptance_rate DECIMAL(5,2) DEFAULT 100.0,
    cancellation_rate DECIMAL(5,2) DEFAULT 0.0,
    is_verified BOOLEAN DEFAULT TRUE,
    wallet_balance DECIMAL(10,2) DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    CONSTRAINT fk_driver_profiles_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX idx_driver_profiles_status ON driver_profiles(status);
CREATE INDEX idx_driver_profiles_is_verified ON driver_profiles(is_verified);
CREATE INDEX idx_driver_profiles_deleted_at ON driver_profiles(deleted_at);
CREATE INDEX idx_driver_profiles_location ON driver_profiles USING GIST(current_location);

-- Vehicles table
CREATE TABLE vehicles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    driver_id UUID NOT NULL UNIQUE,
    vehicle_type_id UUID NOT NULL,
    make VARCHAR(100) NOT NULL,
    model VARCHAR(100) NOT NULL,
    year INTEGER NOT NULL,
    color VARCHAR(50) NOT NULL,
    license_plate VARCHAR(50) NOT NULL UNIQUE,
    capacity INTEGER NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    CONSTRAINT fk_vehicles_driver FOREIGN KEY (driver_id) REFERENCES driver_profiles(id) ON DELETE CASCADE,
    CONSTRAINT fk_vehicles_vehicle_type FOREIGN KEY (vehicle_type_id) REFERENCES vehicle_types(id)
);

CREATE INDEX idx_vehicles_is_active ON vehicles(is_active);
CREATE INDEX idx_vehicles_deleted_at ON vehicles(deleted_at);

-- Driver locations table
CREATE TABLE driver_locations (
    id BIGSERIAL PRIMARY KEY,
    driver_id UUID NOT NULL,
    location GEOMETRY(Point, 4326) NOT NULL,
    latitude DECIMAL(10,8) NOT NULL,
    longitude DECIMAL(11,8) NOT NULL,
    heading INTEGER DEFAULT 0,
    speed DECIMAL(6,2) DEFAULT 0,
    accuracy DECIMAL(6,2) DEFAULT 0,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_driver_locations_driver FOREIGN KEY (driver_id) REFERENCES driver_profiles(id) ON DELETE CASCADE
);

CREATE INDEX idx_driver_locations_driver_id ON driver_locations(driver_id);
CREATE INDEX idx_driver_locations_timestamp ON driver_locations(timestamp);
CREATE INDEX idx_driver_locations_location ON driver_locations USING GIST(location);

-- Rider profiles table
CREATE TABLE rider_profiles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL UNIQUE,
    home_address JSONB,
    work_address JSONB,
    preferred_vehicle_type VARCHAR(50),
    rating DECIMAL(3,2) NOT NULL DEFAULT 5.0,
    total_rides INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_rider_profiles_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Rides table
CREATE TABLE rides (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    rider_id UUID NOT NULL,
    driver_id UUID,
    vehicle_type_id UUID NOT NULL,
    status VARCHAR(50) NOT NULL,
    
    -- Pickup location
    pickup_location GEOMETRY(Point, 4326) NOT NULL,
    pickup_lat DECIMAL(10,8) NOT NULL,
    pickup_lon DECIMAL(11,8) NOT NULL,
    pickup_address TEXT,
    
    -- Dropoff location
    dropoff_location GEOMETRY(Point, 4326) NOT NULL,
    dropoff_lat DECIMAL(10,8) NOT NULL,
    dropoff_lon DECIMAL(11,8) NOT NULL,
    dropoff_address TEXT,
    
    -- Estimates
    estimated_distance DECIMAL(10,2),
    estimated_duration INTEGER,
    estimated_fare DECIMAL(10,2),
    
    -- Actuals
    actual_distance DECIMAL(10,2),
    actual_duration INTEGER,
    actual_fare DECIMAL(10,2),
    
    -- Pricing
    surge_multiplier DECIMAL(3,2) DEFAULT 1.0,
    
    -- Wallet
    wallet_hold_id UUID,
    
    -- Notes
    rider_notes TEXT,
    cancellation_reason TEXT,
    cancelled_by VARCHAR(50),
    
    -- Timestamps
    requested_at TIMESTAMP WITH TIME ZONE NOT NULL,
    accepted_at TIMESTAMP WITH TIME ZONE,
    arrived_at TIMESTAMP WITH TIME ZONE,
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    cancelled_at TIMESTAMP WITH TIME ZONE,
    
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    
    CONSTRAINT fk_rides_rider FOREIGN KEY (rider_id) REFERENCES users(id),
    CONSTRAINT fk_rides_driver FOREIGN KEY (driver_id) REFERENCES users(id),
    CONSTRAINT fk_rides_vehicle_type FOREIGN KEY (vehicle_type_id) REFERENCES vehicle_types(id)
);

CREATE INDEX idx_rides_rider_id ON rides(rider_id);
CREATE INDEX idx_rides_driver_id ON rides(driver_id);
CREATE INDEX idx_rides_status ON rides(status);
CREATE INDEX idx_rides_requested_at ON rides(requested_at);
CREATE INDEX idx_rides_deleted_at ON rides(deleted_at);
CREATE INDEX idx_rides_pickup_location ON rides USING GIST(pickup_location);
CREATE INDEX idx_rides_dropoff_location ON rides USING GIST(dropoff_location);

-- Ride requests table
CREATE TABLE ride_requests (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    ride_id UUID NOT NULL,
    driver_id UUID NOT NULL,
    status VARCHAR(50) NOT NULL,
    sent_at TIMESTAMP WITH TIME ZONE NOT NULL,
    responded_at TIMESTAMP WITH TIME ZONE,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    rejection_reason TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_ride_requests_ride FOREIGN KEY (ride_id) REFERENCES rides(id) ON DELETE CASCADE,
    CONSTRAINT fk_ride_requests_driver FOREIGN KEY (driver_id) REFERENCES driver_profiles(id)
);

CREATE INDEX idx_ride_requests_ride_id ON ride_requests(ride_id);
CREATE INDEX idx_ride_requests_driver_id ON ride_requests(driver_id);
CREATE INDEX idx_ride_requests_status ON ride_requests(status);
CREATE INDEX idx_ride_requests_expires_at ON ride_requests(expires_at);

-- Surge pricing zones table
CREATE TABLE surge_pricing_zones (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    area_name VARCHAR(255) NOT NULL,
    area_geohash VARCHAR(12) NOT NULL,
    center_lat DECIMAL(10,8) NOT NULL,
    center_lon DECIMAL(11,8) NOT NULL,
    radius_km DECIMAL(6,2) NOT NULL,
    multiplier DECIMAL(3,2) NOT NULL DEFAULT 1.0,
    active_from TIMESTAMP WITH TIME ZONE NOT NULL,
    active_until TIMESTAMP WITH TIME ZONE NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_surge_pricing_zones_geohash ON surge_pricing_zones(area_geohash);
CREATE INDEX idx_surge_pricing_zones_is_active ON surge_pricing_zones(is_active);
CREATE INDEX idx_surge_pricing_zones_active_period ON surge_pricing_zones(active_from, active_until);
CREATE INDEX idx_surge_pricing_zones_deleted_at ON surge_pricing_zones(deleted_at);

-- =====================================================
-- HOME SERVICES TABLES
-- =====================================================

-- Services table (ServiceNew)
CREATE TABLE services (
    id UUID PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    long_title VARCHAR(500),
    service_slug VARCHAR(255) NOT NULL UNIQUE,
    category_slug VARCHAR(255) NOT NULL,
    description TEXT,
    long_description TEXT,
    highlights TEXT,
    whats_included TEXT[] NOT NULL DEFAULT '{}',
    terms_and_conditions TEXT[],
    banner_image VARCHAR(500),
    thumbnail VARCHAR(500),
    duration INTEGER,
    is_frequent BOOLEAN DEFAULT FALSE,
    frequency VARCHAR(100),
    sort_order INTEGER DEFAULT 0,
    is_active BOOLEAN DEFAULT TRUE,
    is_available BOOLEAN DEFAULT TRUE,
    base_price DECIMAL(10,2),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_services_category_slug ON services(category_slug);
CREATE INDEX idx_services_is_active ON services(is_active);
CREATE INDEX idx_services_is_available ON services(is_available);
CREATE INDEX idx_services_sort_order ON services(sort_order);
CREATE INDEX idx_services_deleted_at ON services(deleted_at);

-- Addons table
CREATE TABLE addons (
    id UUID PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    addon_slug VARCHAR(255) NOT NULL UNIQUE,
    category_slug VARCHAR(255) NOT NULL,
    description TEXT,
    whats_included TEXT[],
    notes TEXT[],
    image VARCHAR(500),
    price DECIMAL(10,2) NOT NULL,
    strikethrough_price DECIMAL(10,2),
    is_active BOOLEAN DEFAULT TRUE,
    is_available BOOLEAN DEFAULT TRUE,
    sort_order INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_addons_category_slug ON addons(category_slug);
CREATE INDEX idx_addons_is_active ON addons(is_active);
CREATE INDEX idx_addons_is_available ON addons(is_available);
CREATE INDEX idx_addons_sort_order ON addons(sort_order);
CREATE INDEX idx_addons_deleted_at ON addons(deleted_at);

-- Service provider profiles table
CREATE TABLE service_provider_profiles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL UNIQUE,
    business_name VARCHAR(255),
    description TEXT,
    service_category VARCHAR(100) NOT NULL,
    service_type VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending_approval',
    is_verified BOOLEAN DEFAULT FALSE,
    verification_docs JSONB,
    rating DECIMAL(3,2) DEFAULT 0,
    total_reviews INTEGER DEFAULT 0,
    completed_jobs INTEGER DEFAULT 0,
    is_available BOOLEAN DEFAULT TRUE,
    working_hours JSONB,
    service_areas JSONB,
    hourly_rate DECIMAL(10,2),
    currency VARCHAR(3) DEFAULT 'USD',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    CONSTRAINT fk_service_provider_profiles_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX idx_service_provider_profiles_service_type ON service_provider_profiles(service_type);
CREATE INDEX idx_service_provider_profiles_service_category ON service_provider_profiles(service_category);
CREATE INDEX idx_service_provider_profiles_status ON service_provider_profiles(status);
CREATE INDEX idx_service_provider_profiles_is_available ON service_provider_profiles(is_available);
CREATE INDEX idx_service_provider_profiles_deleted_at ON service_provider_profiles(deleted_at);

-- Provider service categories table
CREATE TABLE provider_service_categories (
    id UUID PRIMARY KEY,
    provider_id UUID NOT NULL,
    category_slug VARCHAR(255) NOT NULL,
    expertise_level VARCHAR(50) DEFAULT 'beginner',
    years_of_experience INTEGER DEFAULT 0,
    is_active BOOLEAN DEFAULT TRUE,
    completed_jobs INTEGER DEFAULT 0,
    total_earnings DECIMAL(12,2) DEFAULT 0,
    average_rating DECIMAL(3,2) DEFAULT 0,
    total_ratings INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_provider_service_categories_provider FOREIGN KEY (provider_id) REFERENCES service_provider_profiles(id) ON DELETE CASCADE,
    CONSTRAINT chk_expertise_level CHECK (expertise_level IN ('beginner', 'intermediate', 'expert'))
);

CREATE INDEX idx_provider_service_categories_provider_id ON provider_service_categories(provider_id);
CREATE INDEX idx_provider_service_categories_category_slug ON provider_service_categories(category_slug);
CREATE INDEX idx_provider_service_categories_is_active ON provider_service_categories(is_active);
CREATE UNIQUE INDEX idx_provider_service_categories_unique ON provider_service_categories(provider_id, category_slug);

-- Provider qualified services table (Which specific services can each provider do)
CREATE TABLE IF NOT EXISTS provider_qualified_services (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    provider_id UUID NOT NULL,
    service_id UUID NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_provider_qualified_services_provider FOREIGN KEY (provider_id) REFERENCES service_provider_profiles(id) ON DELETE CASCADE,
    CONSTRAINT fk_provider_qualified_services_service FOREIGN KEY (service_id) REFERENCES services(id) ON DELETE CASCADE,
    CONSTRAINT uk_provider_service_unique UNIQUE(provider_id, service_id)
);

CREATE INDEX idx_provider_qualified_services_provider_id ON provider_qualified_services(provider_id);
CREATE INDEX idx_provider_qualified_services_service_id ON provider_qualified_services(service_id);
CREATE INDEX idx_provider_qualified_services_is_active ON provider_qualified_services(is_active);

-- Service orders table (ServiceOrderNew)
CREATE TABLE service_orders (
    id UUID PRIMARY KEY,
    order_number VARCHAR(50) NOT NULL UNIQUE,
    
    -- Customer
    customer_id UUID NOT NULL,
    customer_info JSONB NOT NULL,
    
    -- Booking
    booking_info JSONB NOT NULL,
    
    -- Services
    category_slug VARCHAR(255) NOT NULL,
    selected_services JSONB NOT NULL,
    selected_addons JSONB,
    special_notes TEXT,
    
    -- Pricing
    services_total DECIMAL(10,2) NOT NULL,
    addons_total DECIMAL(10,2) DEFAULT 0,
    subtotal DECIMAL(10,2) NOT NULL,
    platform_commission DECIMAL(10,2) NOT NULL,
    total_price DECIMAL(10,2) NOT NULL,
    
    -- Payment
    payment_info JSONB,
    wallet_hold_id UUID,
    
    -- Provider
    assigned_provider_id UUID,
    provider_accepted_at TIMESTAMP WITH TIME ZONE,
    provider_started_at TIMESTAMP WITH TIME ZONE,
    provider_completed_at TIMESTAMP WITH TIME ZONE,
    
    -- Status
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    
    -- Cancellation
    cancellation_info JSONB,
    
    -- Customer Rating (of Provider)
    customer_rating INTEGER CHECK (customer_rating >= 1 AND customer_rating <= 5),
    customer_review TEXT,
    customer_rated_at TIMESTAMP WITH TIME ZONE,
    
    -- Provider Rating (of Customer)
    provider_rating INTEGER CHECK (provider_rating >= 1 AND provider_rating <= 5),
    provider_review TEXT,
    provider_rated_at TIMESTAMP WITH TIME ZONE,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    
    CONSTRAINT fk_service_orders_customer FOREIGN KEY (customer_id) REFERENCES users(id),
    CONSTRAINT fk_service_orders_provider FOREIGN KEY (assigned_provider_id) REFERENCES service_provider_profiles(id)
);

CREATE INDEX idx_service_orders_customer_id ON service_orders(customer_id);
CREATE INDEX idx_service_orders_assigned_provider_id ON service_orders(assigned_provider_id);
CREATE INDEX idx_service_orders_category_slug ON service_orders(category_slug);
CREATE INDEX idx_service_orders_status ON service_orders(status);
CREATE INDEX idx_service_orders_created_at ON service_orders(created_at);

-- Order status history table
CREATE TABLE order_status_history (
    id UUID PRIMARY KEY,
    order_id UUID NOT NULL,
    from_status VARCHAR(50),
    to_status VARCHAR(50) NOT NULL,
    changed_by UUID,
    changed_by_role VARCHAR(50),
    notes TEXT,
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_order_status_history_order FOREIGN KEY (order_id) REFERENCES service_orders(id) ON DELETE CASCADE,
    CONSTRAINT fk_order_status_history_user FOREIGN KEY (changed_by) REFERENCES users(id)
);

CREATE INDEX idx_order_status_history_order_id ON order_status_history(order_id);
CREATE INDEX idx_order_status_history_created_at ON order_status_history(created_at);

-- =====================================================
-- TRIGGERS FOR updated_at
-- =====================================================

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_wallets_updated_at BEFORE UPDATE ON wallets
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_vehicle_types_updated_at BEFORE UPDATE ON vehicle_types
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_driver_profiles_updated_at BEFORE UPDATE ON driver_profiles
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_vehicles_updated_at BEFORE UPDATE ON vehicles
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_rider_profiles_updated_at BEFORE UPDATE ON rider_profiles
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_rides_updated_at BEFORE UPDATE ON rides
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_ride_requests_updated_at BEFORE UPDATE ON ride_requests
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_surge_pricing_zones_updated_at BEFORE UPDATE ON surge_pricing_zones
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_services_updated_at BEFORE UPDATE ON services
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_addons_updated_at BEFORE UPDATE ON addons
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_service_provider_profiles_updated_at BEFORE UPDATE ON service_provider_profiles
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_provider_service_categories_updated_at BEFORE UPDATE ON provider_service_categories
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_service_orders_updated_at BEFORE UPDATE ON service_orders
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();