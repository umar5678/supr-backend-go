#!/bin/bash
# Diagnostic script to test backend before running k6

set -e

BASE_URL="${1:-http://localhost:8080}"

echo "🔍 Backend Diagnostic Test"
echo "================================"
echo "Testing URL: $BASE_URL"
echo ""

# Test 1: Health Check
echo "1️⃣  Testing Health Endpoint..."
HEALTH=$(curl -s -w "\n%{http_code}" "$BASE_URL/health")
HTTP_CODE=$(echo "$HEALTH" | tail -1)
BODY=$(echo "$HEALTH" | head -1)

if [ "$HTTP_CODE" = "200" ]; then
    echo "   ✅ Health check passed (HTTP $HTTP_CODE)"
    echo "   Response: $BODY"
else
    echo "   ❌ Health check failed (HTTP $HTTP_CODE)"
    echo "   Response: $BODY"
fi
echo ""

# Test 2: CORS Headers
echo "2️⃣  Testing CORS Headers..."
CORS=$(curl -s -i -X OPTIONS "$BASE_URL/health" 2>/dev/null | grep -i "access-control")

if [ -z "$CORS" ]; then
    echo "   ⚠️  No CORS headers detected (OK for non-browser requests)"
else
    echo "   ✅ CORS headers found:"
    echo "$CORS" | sed 's/^/      /'
fi
echo ""

# Test 3: Categories Endpoint
echo "3️⃣  Testing Categories Endpoint..."
CATEGORIES=$(curl -s -w "\n%{http_code}" "$BASE_URL/api/v1/homeservices/categories")
HTTP_CODE=$(echo "$CATEGORIES" | tail -1)

if [ "$HTTP_CODE" = "200" ] || [ "$HTTP_CODE" = "404" ] || [ "$HTTP_CODE" = "401" ]; then
    echo "   ✅ Categories endpoint responded (HTTP $HTTP_CODE)"
else
    echo "   ❌ Categories endpoint failed (HTTP $HTTP_CODE)"
fi
echo ""

# Test 4: Service Providers Endpoint
echo "4️⃣  Testing Service Providers Endpoint..."
PROVIDERS=$(curl -s -w "\n%{http_code}" "$BASE_URL/api/v1/serviceproviders")
HTTP_CODE=$(echo "$PROVIDERS" | tail -1)

if [ "$HTTP_CODE" = "200" ] || [ "$HTTP_CODE" = "404" ] || [ "$HTTP_CODE" = "401" ]; then
    echo "   ✅ Providers endpoint responded (HTTP $HTTP_CODE)"
else
    echo "   ❌ Providers endpoint failed (HTTP $HTTP_CODE)"
fi
echo ""

echo "================================"
echo "✅ Diagnostic complete!"
echo ""
echo "If all tests passed, run:"
echo "  k6 run k6/basic-load-test.js"
