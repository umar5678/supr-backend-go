-- ============================================================================
-- FIND ALL PROVIDERS WITH MISSING SERVICES
-- ============================================================================
-- This query identifies all providers that need service assignments

-- Check all providers and their service counts
SELECT 
  spp.id as provider_id,
  spp.user_id,
  spp.service_type,
  COALESCE(COUNT(DISTINCT pqs.service_id), 0) as current_qualified_services,
  COALESCE(COUNT(DISTINCT CASE WHEN so.id IS NOT NULL THEN so.id END), 0) as available_orders
FROM service_provider_profiles spp
LEFT JOIN provider_qualified_services pqs ON spp.id = pqs.provider_id
LEFT JOIN services s ON pqs.service_id = s.id
LEFT JOIN service_orders so ON so.category_slug = s.category_slug
  AND so.status IN ('pending', 'searching_provider')
  AND so.assigned_provider_id IS NULL
GROUP BY spp.id, spp.user_id, spp.service_type
ORDER BY current_qualified_services ASC, spp.created_at DESC;

-- ============================================================================
-- IDENTIFY PROVIDERS NEEDING FIX (those with 0 services)
-- ============================================================================
SELECT 
  spp.id as provider_id,
  spp.user_id,
  spp.service_type,
  COUNT(pqs.service_id) as qualified_services_count
FROM service_provider_profiles spp
LEFT JOIN provider_qualified_services pqs ON spp.id = pqs.provider_id
GROUP BY spp.id, spp.user_id, spp.service_type
HAVING COUNT(pqs.service_id) = 0
ORDER BY spp.created_at DESC;

-- ============================================================================
-- FIX: Assign services to ALL providers with 0 services
-- ============================================================================
-- This will only add services to providers that don't have any
INSERT INTO provider_qualified_services (provider_id, service_id)
SELECT 
  spp.id as provider_id,
  s.id as service_id
FROM service_provider_profiles spp
CROSS JOIN services s
WHERE spp.id IN (
  -- Subquery: find providers with 0 services
  SELECT spp2.id 
  FROM service_provider_profiles spp2
  LEFT JOIN provider_qualified_services pqs2 ON spp2.id = pqs2.provider_id
  GROUP BY spp2.id
  HAVING COUNT(pqs2.service_id) = 0
)
AND s.is_active = true
AND s.is_available = true;

-- ============================================================================
-- VERIFY THE FIX
-- ============================================================================
-- Show all providers and their updated service counts
SELECT 
  spp.id as provider_id,
  spp.service_type,
  COUNT(DISTINCT pqs.service_id) as total_qualified_services,
  COUNT(DISTINCT s.category_slug) as accessible_categories,
  STRING_AGG(DISTINCT s.category_slug, ', ' ORDER BY s.category_slug) as categories
FROM service_provider_profiles spp
LEFT JOIN provider_qualified_services pqs ON spp.id = pqs.provider_id
LEFT JOIN services s ON pqs.service_id = s.id
GROUP BY spp.id, spp.service_type
ORDER BY total_qualified_services DESC, spp.created_at DESC;

-- ============================================================================
-- DETAILED PROVIDER STATUS REPORT
-- ============================================================================
SELECT 
  spp.id as provider_id,
  spp.user_id,
  spp.service_type,
  COUNT(DISTINCT pqs.service_id) as qualified_services,
  COUNT(DISTINCT s.category_slug) as accessible_categories,
  COUNT(DISTINCT CASE WHEN so.id IS NOT NULL THEN so.id END) as available_orders,
  CASE 
    WHEN COUNT(DISTINCT pqs.service_id) = 0 THEN '❌ NO SERVICES'
    WHEN COUNT(DISTINCT CASE WHEN so.id IS NOT NULL THEN so.id END) = 0 THEN '⚠️  NO ORDERS'
    ELSE '✅ OPERATIONAL'
  END as status
FROM service_provider_profiles spp
LEFT JOIN provider_qualified_services pqs ON spp.id = pqs.provider_id
LEFT JOIN services s ON pqs.service_id = s.id
LEFT JOIN service_orders so ON so.category_slug = s.category_slug
  AND so.status IN ('pending', 'searching_provider')
  AND so.assigned_provider_id IS NULL
GROUP BY spp.id, spp.user_id, spp.service_type
ORDER BY 
  CASE 
    WHEN COUNT(DISTINCT pqs.service_id) = 0 THEN 0
    WHEN COUNT(DISTINCT CASE WHEN so.id IS NOT NULL THEN so.id END) = 0 THEN 1
    ELSE 2
  END,
  spp.created_at DESC;

-- ============================================================================
-- SPECIFIC PROVIDER CHECK: The new provider
-- ============================================================================
-- Provider ID: 1bbe0f76-c324-4a7b-85da-3650359f5f6f
SELECT 
  spp.id as provider_id,
  spp.user_id,
  spp.service_type,
  spp.service_category,
  COUNT(DISTINCT pqs.service_id) as qualified_services,
  COUNT(DISTINCT s.category_slug) as accessible_categories,
  COUNT(DISTINCT so.id) as available_orders
FROM service_provider_profiles spp
LEFT JOIN provider_qualified_services pqs ON spp.id = pqs.provider_id
LEFT JOIN services s ON pqs.service_id = s.id
LEFT JOIN service_orders so ON so.category_slug = s.category_slug
  AND so.status IN ('pending', 'searching_provider')
  AND so.assigned_provider_id IS NULL
WHERE spp.id = '1bbe0f76-c324-4a7b-85da-3650359f5f6f'
GROUP BY spp.id, spp.user_id, spp.service_type, spp.service_category;

-- Before fix
-- Then after running the FIX section above, run this query again to verify

-- ============================================================================
-- ROLLBACK (if needed)
-- ============================================================================
/*
DELETE FROM provider_qualified_services
WHERE provider_id IN (
  SELECT spp.id 
  FROM service_provider_profiles spp
  LEFT JOIN provider_qualified_services pqs ON spp.id = pqs.provider_id
  GROUP BY spp.id
  HAVING COUNT(pqs.service_id) = 0
);

-- Verify
SELECT COUNT(*) FROM provider_qualified_services;
*/
