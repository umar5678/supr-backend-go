-- +goose Up
-- Make current_location nullable and ensure correct geometry type

ALTER TABLE driver_profiles
    ALTER COLUMN current_location DROP NOT NULL;

ALTER TABLE driver_profiles
    ALTER COLUMN current_location TYPE geometry(Point,4326)
        USING current_location::geometry;
