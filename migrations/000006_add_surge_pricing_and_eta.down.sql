-- Migration: Drop surge pricing rules, demand tracking, and ETA estimates
-- Version: 000006

DROP TRIGGER IF EXISTS update_surge_history_updated_at ON surge_history;
DROP TRIGGER IF EXISTS update_eta_estimates_updated_at ON eta_estimates;
DROP TRIGGER IF EXISTS update_demand_tracking_updated_at ON demand_tracking;
DROP TRIGGER IF EXISTS update_surge_pricing_rules_updated_at ON surge_pricing_rules;

DROP TABLE IF EXISTS surge_history;
DROP TABLE IF EXISTS eta_estimates;
DROP TABLE IF EXISTS demand_tracking;
DROP TABLE IF EXISTS surge_pricing_rules;
