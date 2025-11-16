
#!/bin/bash

# Configuration
BASE_URL="http://localhost:8080/api/v1"

echo "=== Testing Public Endpoints ==="
echo ""

# ==========================================
# AUTH ENDPOINTS
# ==========================================

# echo "1. Phone Signup (Rider)"
# curl -X POST "${BASE_URL}/auth/phone/signup" \
#   -H "Content-Type: application/json" \
#   -d '{
#     "name": "John Doe",
#     "phone": "+1234567890",
#     "role": "rider"
#   }'
# echo -e "\n"

# echo "2. Phone Signup (Driver)"
# curl -X POST "${BASE_URL}/auth/phone/signup" \
#   -H "Content-Type: application/json" \
#   -d '{
#     "name": "Jane Driver",
#     "phone": "+1234567891",
#     "role": "driver"
#   }'
# echo -e "\n"

# echo "3. Email Signup"
# curl -X POST "${BASE_URL}/auth/email/signup" \
#   -H "Content-Type: application/json" \
#   -d '{
#     "name": "Admin User",
#     "email": "admin@example.com",
#     "password": "securePassword123",
#     "role": "admin"
#   }'
# echo -e "\n"

echo "4. Phone Login (Rider)"
curl -X POST "${BASE_URL}/auth/phone/login" \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "+1234567890",
    "role": "rider"
  }'
echo -e "\n"

echo "5. Phone Login (Driver)"
curl -X POST "${BASE_URL}/auth/phone/login" \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "+1234567891",
    "role": "driver"
  }'
echo -e "\n"

echo "6. Email Login"
curl -X POST "${BASE_URL}/auth/email/login" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@example.com",
    "password": "securePassword123"
  }'
echo -e "\n"

echo "7. Refresh Token"
REFRESH_TOKEN="your_refresh_token_here"
curl -X POST "${BASE_URL}/auth/refresh" \
  -H "Content-Type: application/json" \
  -d "{
    \"refreshToken\": \"${REFRESH_TOKEN}\"
  }"
echo -e "\n"

# ==========================================
# VEHICLE TYPES (PUBLIC)
# ==========================================

echo "8. Get All Vehicle Types"
curl -X GET "${BASE_URL}/vehicles/types"
echo -e "\n"

echo "9. Get Active Vehicle Types"
curl -X GET "${BASE_URL}/vehicles/types/active"
echo -e "\n"

echo "10. Get Vehicle Type By ID"
VEHICLE_TYPE_ID="some-uuid-here"
curl -X GET "${BASE_URL}/vehicles/types/${VEHICLE_TYPE_ID}"
echo -e "\n"

# ==========================================
# PRICING (PUBLIC)
# ==========================================

echo "11. Get Fare Estimate"
curl -X POST "${BASE_URL}/pricing/estimate" \
  -H "Content-Type: application/json" \
  -d '{
    "pickupLat": 40.7128,
    "pickupLon": -74.0060,
    "dropoffLat": 40.7589,
    "dropoffLon": -73.9851,
    "vehicleTypeId": "some-uuid-here"
  }'
echo -e "\n"

echo "12. Get Surge Multiplier"
curl -X GET "${BASE_URL}/pricing/surge?latitude=40.7128&longitude=-74.0060"
echo -e "\n"

echo "13. Get All Surge Zones"
curl -X GET "${BASE_URL}/pricing/surge/zones"
echo -e "\n"

# ==========================================
# TRACKING (PUBLIC)
# ==========================================

echo "14. Find Nearby Drivers"
curl -X GET "${BASE_URL}/tracking/nearby?latitude=40.7128&longitude=-74.0060&radiusKm=5&limit=10"
echo -e "\n"

echo "15. Get Driver Location"
DRIVER_ID="some-uuid-here"
curl -X GET "${BASE_URL}/tracking/driver/${DRIVER_ID}"
echo -e "\n"

# ==========================================
# HOME SERVICES (PUBLIC)
# ==========================================

echo "16. List Service Categories"
curl -X GET "${BASE_URL}/services/categories"
echo -e "\n"

echo "17. List Services"
curl -X GET "${BASE_URL}/services?page=1&limit=20"
echo -e "\n"

echo "18. List Services with Filters"
curl -X GET "${BASE_URL}/services?page=1&limit=20&categoryId=1&search=cleaning&minPrice=10&maxPrice=100&isActive=true"
echo -e "\n"

echo "19. Get Service Details"
SERVICE_ID="1"
curl -X GET "${BASE_URL}/services/${SERVICE_ID}"
echo -e "\n"

echo "=== Public Endpoints Testing Complete ==="