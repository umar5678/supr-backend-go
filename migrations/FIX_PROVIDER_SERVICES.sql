-- ============================================================================
-- FIX: Assign Services to Provider (Provider No Qualified Services)
-- ============================================================================
-- 
-- Problem: Provider 749bd875-2336-41fa-a67d-06a511fe3213 has 0 qualified services
-- Solution: Bulk-insert services into provider_qualified_services table
--
-- ============================================================================

-- STEP 1: Verify provider exists
SELECT id, service_type, service_category FROM service_provider_profiles
WHERE id = '749bd875-2336-41fa-a67d-06a511fe3213';
-- Expected: Should return 1 row

-- STEP 2: Check current qualified services
SELECT COUNT(*) as current_qualified_services 
FROM provider_qualified_services
WHERE provider_id = '749bd875-2336-41fa-a67d-06a511fe3213';
-- Expected: Should return 0

-- STEP 3: List available services to assign
SELECT id, title, category_slug FROM services
WHERE is_active = true AND is_available = true
LIMIT 20;
-- This shows you what services exist in the DB

-- ============================================================================
-- STEP 4: RUN THIS TO FIX - Option A: Assign ALL active services
-- ============================================================================
INSERT INTO provider_qualified_services (provider_id, service_id)
SELECT 
  '749bd875-2336-41fa-a67d-06a511fe3213' as provider_id,
  s.id as service_id
FROM services s
WHERE s.is_active = true
  AND s.is_available = true
ON CONFLICT DO NOTHING;

-- Verify how many were assigned
SELECT COUNT(*) as newly_assigned FROM provider_qualified_services
WHERE provider_id = '749bd875-2336-41fa-a67d-06a511fe3213';

-- ============================================================================
-- ALTERNATIVE: Option B - Assign only CLEANING SERVICES
-- ============================================================================
-- Uncomment if you only want to assign specific services
/*
INSERT INTO provider_qualified_services (provider_id, service_id)
SELECT 
  '749bd875-2336-41fa-a67d-06a511fe3213' as provider_id,
  s.id as service_id
FROM services s
WHERE s.category_slug = 'cleaning-services'
  AND s.is_active = true
ON CONFLICT DO NOTHING;
*/

-- ============================================================================
-- STEP 5: Verify the fix worked
-- ============================================================================

-- Check categories provider now has
SELECT DISTINCT s.category_slug as provider_category, COUNT(*) as service_count
FROM provider_qualified_services pqs
JOIN services s ON pqs.service_id = s.id
WHERE pqs.provider_id = '749bd875-2336-41fa-a67d-06a511fe3213'
GROUP BY s.category_slug
ORDER BY s.category_slug;
-- Expected: Should show multiple rows with different categories

-- Check orders available in those categories
SELECT so.id, so.order_number, so.category_slug, so.status, so.total_price
FROM service_orders so
WHERE so.category_slug IN (
  SELECT DISTINCT s.category_slug
  FROM provider_qualified_services pqs
  JOIN services s ON pqs.service_id = s.id
  WHERE pqs.provider_id = '749bd875-2336-41fa-a67d-06a511fe3213'
)
AND so.status IN ('pending', 'searching_provider')
AND so.assigned_provider_id IS NULL
ORDER BY so.created_at DESC
LIMIT 10;
-- Expected: Should show 4-8 available orders

-- ============================================================================
-- STEP 6: Test the API
-- ============================================================================
-- Call: GET /api/v1/provider/orders/available?page=1&limit=100
-- 
-- Expected response should now show:
-- {
--   "success": true,
--   "message": "Found 4 available orders matching your qualifications",
--   "data": {
--     "orders": [ /* array with actual orders */ ],
--     "metadata": {
--       "qualifiedCategories": ["cleaning-services", "women-spa", "men-spa", "men-salon"],
--       "totalCategoriesCount": 4,
--       "ordersFound": true
--     },
--     "totalCount": 4
--   }
-- }

-- ============================================================================
-- ROLLBACK (if something goes wrong)
-- ============================================================================
/*
DELETE FROM provider_qualified_services
WHERE provider_id = '749bd875-2336-41fa-a67d-06a511fe3213';

-- Verify rollback
SELECT COUNT(*) FROM provider_qualified_services
WHERE provider_id = '749bd875-2336-41fa-a67d-06a511fe3213';
-- Should return 0
*/

-- ============================================================================
-- BATCH FIX: Apply to ALL providers with 0 services (if needed)
-- ============================================================================
/*
-- Find all providers with no services
SELECT spp.id, spp.service_type, COUNT(pqs.service_id) as service_count
FROM service_provider_profiles spp
LEFT JOIN provider_qualified_services pqs ON spp.id = pqs.provider_id
GROUP BY spp.id, spp.service_type
HAVING COUNT(pqs.service_id) = 0;

-- Batch assign all services to these providers (careful with this!)
INSERT INTO provider_qualified_services (provider_id, service_id)
SELECT 
  spp.id as provider_id,
  s.id as service_id
FROM service_provider_profiles spp
CROSS JOIN services s
WHERE spp.id IN (
  SELECT spp2.id FROM service_provider_profiles spp2
  LEFT JOIN provider_qualified_services pqs2 ON spp2.id = pqs2.provider_id
  GROUP BY spp2.id
  HAVING COUNT(pqs2.service_id) = 0
)
AND s.is_active = true
AND s.is_available = true
ON CONFLICT DO NOTHING;
*/

-- ============================================================================
-- Notes:
-- - All queries use ON CONFLICT DO NOTHING to prevent duplicates
-- - It's safe to run these multiple times
-- - Assignments are permanent; use ROLLBACK section if you need to undo
-- - After running, restart API or wait for next request to see changes
-- ============================================================================
