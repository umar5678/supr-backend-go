-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create enum types
CREATE TYPE user_role AS ENUM ('rider', 'driver', 'admin', 'delivery_person', 'service_provider', 'handyman');
CREATE TYPE user_status AS ENUM ('active', 'suspended', 'banned', 'pending_verification');

-- Create users table
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE,
    phone VARCHAR(20) UNIQUE,
    password VARCHAR(255),
    role user_role NOT NULL DEFAULT 'rider',
    status user_status NOT NULL DEFAULT 'active',
    profile_photo_url VARCHAR(500),
    last_login_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    
    -- Constraints
    CONSTRAINT email_or_phone_required CHECK (
        (email IS NOT NULL AND password IS NOT NULL) OR 
        (phone IS NOT NULL)
    ),
    CONSTRAINT email_for_non_riders CHECK (
        role IN ('rider', 'driver') OR 
        (email IS NOT NULL AND password IS NOT NULL)
    )
);

-- Create indexes
CREATE INDEX idx_users_email ON users(email) WHERE email IS NOT NULL;
CREATE INDEX idx_users_phone ON users(phone) WHERE phone IS NOT NULL;
CREATE INDEX idx_users_role ON users(role);
CREATE INDEX idx_users_status ON users(status);
CREATE INDEX idx_users_deleted_at ON users(deleted_at) WHERE deleted_at IS NOT NULL;

-- Comments
COMMENT ON TABLE users IS 'Main users table supporting multiple authentication methods';
COMMENT ON COLUMN users.phone IS 'Phone number for rider/driver authentication (unique)';
COMMENT ON COLUMN users.email IS 'Email for other role authentication (unique)';
COMMENT ON COLUMN users.password IS 'Hashed password for email-based authentication';
COMMENT ON COLUMN users.role IS 'User role: rider, driver, admin, etc.';
COMMENT ON COLUMN users.status IS 'User account status';

-- Trigger for updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();