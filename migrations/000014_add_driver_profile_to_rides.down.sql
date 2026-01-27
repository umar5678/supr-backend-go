-- Rollback: Remove driver_profile relationship support from rides table
-- This migration documents the removal of GORM relationship between rides and driver_profiles
-- No schema changes needed as the relationship is defined through GORM model only

-- No actual SQL changes required - this is a model-level relationship in GORM
