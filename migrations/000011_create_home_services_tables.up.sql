-- Enable extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "postgis";

-- Ensure wallet enum types exist (they should from your existing system)
-- If they don't exist, create them
DO $$ BEGIN
    CREATE TYPE wallet_type AS ENUM ('rider', 'driver', 'platform', 'service_provider');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

DO $$ BEGIN
    CREATE TYPE transaction_type AS ENUM ('credit', 'debit', 'refund', 'hold', 'release', 'transfer');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

DO $$ BEGIN
    CREATE TYPE transaction_status AS ENUM ('pending', 'completed', 'failed', 'cancelled', 'held', 'released');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;


-- SERVICE CATALOG --
CREATE TABLE "service_categories" (
    "id" SERIAL PRIMARY KEY,
    "name" VARCHAR(100) NOT NULL,
    "description" TEXT,
    "icon_url" VARCHAR(255),
    "is_active" BOOLEAN NOT NULL DEFAULT TRUE,
    "sort_order" INT NOT NULL DEFAULT 0,
    "created_at" TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    "updated_at" TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE "services" (
    "id" SERIAL PRIMARY KEY,
    "category_id" INT NOT NULL,
    "name" VARCHAR(150) NOT NULL,
    "description" TEXT,
    "image_url" VARCHAR(255),
    "base_price" DECIMAL(10, 2) NOT NULL,
    "pricing_model" VARCHAR(50) NOT NULL,
    "base_duration_minutes" INT NOT NULL,
    "is_active" BOOLEAN NOT NULL DEFAULT TRUE,
    "created_at" TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    "updated_at" TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_category FOREIGN KEY("category_id") REFERENCES "service_categories"("id") ON DELETE CASCADE
);
CREATE INDEX idx_services_category ON "services"("category_id");
CREATE INDEX idx_services_active ON "services"("is_active");

CREATE TABLE "service_options" (
    "id" SERIAL PRIMARY KEY,
    "service_id" INT NOT NULL,
    "name" VARCHAR(100) NOT NULL,
    "type" VARCHAR(50) NOT NULL,
    "is_required" BOOLEAN NOT NULL DEFAULT FALSE,
    "created_at" TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_service FOREIGN KEY("service_id") REFERENCES "services"("id") ON DELETE CASCADE
);
CREATE INDEX idx_service_options_service ON "service_options"("service_id");

CREATE TABLE "service_option_choices" (
    "id" SERIAL PRIMARY KEY,
    "option_id" INT NOT NULL,
    "label" VARCHAR(100) NOT NULL,
    "price_modifier" DECIMAL(10, 2) NOT NULL DEFAULT 0.00,
    "duration_modifier_minutes" INT NOT NULL DEFAULT 0,
    "created_at" TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_option FOREIGN KEY("option_id") REFERENCES "service_options"("id") ON DELETE CASCADE
);
CREATE INDEX idx_service_option_choices_option ON "service_option_choices"("option_id");


-- SERVICE PROVIDERS --
CREATE TABLE "service_providers" (
    "id" UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    "user_id" UUID NOT NULL UNIQUE,
    "photo" VARCHAR(255),
    "rating" DECIMAL(3, 2) NOT NULL DEFAULT 5.00,
    "status" VARCHAR(50) NOT NULL DEFAULT 'offline',
    "location" GEOGRAPHY(Point, 4326),
    "last_active" TIMESTAMPTZ,
    "is_verified" BOOLEAN NOT NULL DEFAULT FALSE,
    "total_jobs" INT NOT NULL DEFAULT 0,
    "completed_jobs" INT NOT NULL DEFAULT 0,
    "created_at" TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    "updated_at" TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_user FOREIGN KEY("user_id") REFERENCES "users"("id") ON DELETE CASCADE
);
CREATE INDEX idx_service_providers_user ON "service_providers"("user_id");
CREATE INDEX idx_service_providers_status ON "service_providers"("status");
CREATE INDEX idx_service_providers_location ON "service_providers" USING GIST ("location");

CREATE TABLE "provider_qualified_services" (
    "provider_id" UUID NOT NULL,
    "service_id" INT NOT NULL,
    PRIMARY KEY ("provider_id", "service_id"),
    CONSTRAINT fk_provider FOREIGN KEY("provider_id") REFERENCES "service_providers"("id") ON DELETE CASCADE,
    CONSTRAINT fk_service FOREIGN KEY("service_id") REFERENCES "services"("id") ON DELETE CASCADE
);
CREATE INDEX idx_pqs_service ON "provider_qualified_services"("service_id");


-- ORDERS --
CREATE TABLE "service_orders" (
    "id" UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    "code" VARCHAR(20) UNIQUE NOT NULL,
    "user_id" UUID NOT NULL,
    "provider_id" UUID,
    "status" VARCHAR(50) NOT NULL,
    "address" TEXT NOT NULL,
    "location" GEOGRAPHY(Point, 4326) NOT NULL,
    "service_date" TIMESTAMPTZ NOT NULL,
    "notes" TEXT,
    "frequency" VARCHAR(20) NOT NULL DEFAULT 'once',
    "subtotal" DECIMAL(10, 2) NOT NULL,
    "discount" DECIMAL(10, 2) NOT NULL DEFAULT 0.00,
    "surge_fee" DECIMAL(10, 2) NOT NULL DEFAULT 0.00,
    "platform_fee" DECIMAL(10, 2) NOT NULL DEFAULT 0.00,
    "total" DECIMAL(10, 2) NOT NULL,
    "coupon_code" VARCHAR(50),
    "wallet_hold_id" UUID,
    "created_at" TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    "accepted_at" TIMESTAMPTZ,
    "started_at" TIMESTAMPTZ,
    "completed_at" TIMESTAMPTZ,
    "cancelled_at" TIMESTAMPTZ,
    CONSTRAINT fk_user FOREIGN KEY("user_id") REFERENCES "users"("id"),
    CONSTRAINT fk_provider FOREIGN KEY("provider_id") REFERENCES "service_providers"("id"),
    CONSTRAINT fk_wallet_hold FOREIGN KEY("wallet_hold_id") REFERENCES "wallet_holds"("id")
);
CREATE INDEX idx_service_orders_user ON "service_orders"("user_id");
CREATE INDEX idx_service_orders_provider ON "service_orders"("provider_id");
CREATE INDEX idx_service_orders_status ON "service_orders"("status");
CREATE INDEX idx_service_orders_created ON "service_orders"("created_at");

CREATE TABLE "order_items" (
    "id" SERIAL PRIMARY KEY,
    "order_id" UUID NOT NULL,
    "service_id" INT NOT NULL,
    "service_name" VARCHAR(150) NOT NULL,
    "base_price" DECIMAL(10, 2) NOT NULL,
    "calculated_price" DECIMAL(10, 2) NOT NULL,
    "duration_minutes" INT NOT NULL,
    "selected_options" JSONB,
    CONSTRAINT fk_order FOREIGN KEY("order_id") REFERENCES "service_orders"("id") ON DELETE CASCADE,
    CONSTRAINT fk_service FOREIGN KEY("service_id") REFERENCES "services"("id")
);
CREATE INDEX idx_order_items_order ON "order_items"("order_id");


-- RATINGS --
CREATE TABLE "ratings" (
    "id" UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    "order_id" UUID NOT NULL UNIQUE,
    "user_id" UUID NOT NULL,
    "provider_id" UUID NOT NULL,
    "score" INT NOT NULL CHECK (score >= 1 AND score <= 5),
    "comment" TEXT,
    "created_at" TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_order FOREIGN KEY("order_id") REFERENCES "service_orders"("id"),
    CONSTRAINT fk_user FOREIGN KEY("user_id") REFERENCES "users"("id"),
    CONSTRAINT fk_provider FOREIGN KEY("provider_id") REFERENCES "service_providers"("id")
);
CREATE INDEX idx_ratings_provider ON "ratings"("provider_id");
CREATE INDEX idx_ratings_user ON "ratings"("user_id");


-- SURGE ZONES (optional) --
CREATE TABLE "surge_zones" (
    "id" SERIAL PRIMARY KEY,
    "name" VARCHAR(100) NOT NULL,
    "zone" GEOGRAPHY(Polygon, 4326) NOT NULL,
    "surge_multiplier" DECIMAL(3, 2) NOT NULL DEFAULT 1.00,
    "is_active" BOOLEAN NOT NULL DEFAULT TRUE,
    "valid_from" TIMESTAMPTZ,
    "valid_to" TIMESTAMPTZ,
    "created_at" TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    "updated_at" TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_surge_zones_zone ON "surge_zones" USING GIST ("zone");
CREATE INDEX idx_surge_zones_active ON "surge_zones"("is_active");