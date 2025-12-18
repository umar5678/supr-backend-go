-- ============================================================================
-- FIX: Assign Services to Provider (CORRECTED - Using Actual Provider IDs)
-- ============================================================================
-- 
-- Problem: The provider ID 749bd875-2336-41fa-a67d-06a511fe3213 does NOT exist
-- Solution: Use actual provider IDs from the database
--
-- Actual Providers in DB:
-- - 4b4f8116-9634-4922-ab43-6d1e514e810c (men-salon)
-- - 8943f1a0-7139-4d82-a7d0-d25a83e8ac9f (pest-control)
-- - 19ed9d7b-b193-4ade-aa64-c611e46c191a (men-spa)
-- - f0b376dc-37ff-432c-b61b-e57275ea4271 (men-spa)
--
-- ============================================================================

-- STEP 1: List all providers in database
SELECT id, user_id, service_type, service_category 
FROM service_provider_profiles
ORDER BY created_at DESC;

-- STEP 2: Check qualified services for each provider
SELECT 
  provider_id,
  COUNT(*) as qualified_service_count
FROM provider_qualified_services
GROUP BY provider_id;

-- ============================================================================
-- STEP 3: ASSIGN SERVICES TO ALL PROVIDERS
-- ============================================================================
-- This will assign ALL active services to EACH provider
INSERT INTO provider_qualified_services (provider_id, service_id)
SELECT 
  spp.id as provider_id,
  s.id as service_id
FROM service_provider_profiles spp
CROSS JOIN services s
WHERE s.is_active = true
  AND s.is_available = true
ON CONFLICT DO NOTHING;

-- Verify assignments
SELECT 
  provider_id,
  COUNT(*) as newly_assigned_services
FROM provider_qualified_services
GROUP BY provider_id
ORDER BY newly_assigned_services DESC;

-- ============================================================================
-- STEP 4: VERIFY - Check categories available to each provider
-- ============================================================================
SELECT 
  spp.id as provider_id,
  spp.service_type,
  COUNT(DISTINCT s.category_slug) as available_categories,
  STRING_AGG(DISTINCT s.category_slug, ', ' ORDER BY s.category_slug) as categories
FROM service_provider_profiles spp
LEFT JOIN provider_qualified_services pqs ON spp.id = pqs.provider_id
LEFT JOIN services s ON pqs.service_id = s.id
GROUP BY spp.id, spp.service_type
ORDER BY spp.created_at DESC;

-- ============================================================================
-- STEP 5: Check available orders for each provider
-- ============================================================================
SELECT 
  spp.id as provider_id,
  spp.service_type,
  COUNT(DISTINCT so.id) as available_orders,
  STRING_AGG(DISTINCT so.category_slug, ', ') as order_categories
FROM service_provider_profiles spp
JOIN provider_qualified_services pqs ON spp.id = pqs.provider_id
JOIN services s ON pqs.service_id = s.id
LEFT JOIN service_orders so ON so.category_slug = s.category_slug
  AND so.status IN ('pending', 'searching_provider')
  AND so.assigned_provider_id IS NULL
GROUP BY spp.id, spp.service_type
ORDER BY available_orders DESC;

-- ============================================================================
-- STEP 6: Example - Get orders for first provider
-- ============================================================================
WITH first_provider AS (
  SELECT id, service_type 
  FROM service_provider_profiles 
  ORDER BY created_at ASC 
  LIMIT 1
)
SELECT 
  fp.id as provider_id,
  fp.service_type,
  so.id as order_id,
  so.order_number,
  so.category_slug,
  so.status,
  so.total_price,
  so.created_at
FROM first_provider fp
JOIN provider_qualified_services pqs ON fp.id = pqs.provider_id
JOIN services s ON pqs.service_id = s.id
JOIN service_orders so ON so.category_slug = s.category_slug
  AND so.status IN ('pending', 'searching_provider')
  AND so.assigned_provider_id IS NULL
ORDER BY so.created_at DESC
LIMIT 10;

-- ============================================================================
-- ROLLBACK (if needed)
-- ============================================================================
/*
DELETE FROM provider_qualified_services;

-- Verify
SELECT COUNT(*) FROM provider_qualified_services;
-- Should return 0
*/
