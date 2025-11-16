# API Testing Guide

This guide contains curl scripts to test all endpoints of your Go backend API.

## Files Overview

1. **public_endpoints.sh** - Endpoints that don't require authentication (19 tests)
2. **auth_endpoints.sh** - Endpoints requiring authentication (51 tests)
3. **admin_endpoints.sh** - Admin-only endpoints (8 tests)

## Prerequisites

- Bash shell (Linux/Mac) or Git Bash (Windows)
- curl installed
- Your backend server running (default: `http://localhost:8080`)

## Setup Instructions

### 1. Make Scripts Executable

```bash
chmod +x public_endpoints.sh
chmod +x auth_endpoints.sh
chmod +x admin_endpoints.sh
```

### 2. Update Configuration

#### For Public Endpoints
Edit `public_endpoints.sh` and update:
```bash
BASE_URL="http://localhost:8080"  # Change if your server runs on different port
```

#### For Authenticated Endpoints
Edit `auth_endpoints.sh` and update:
```bash
BASE_URL="http://localhost:8080"
ACCESS_TOKEN="your_access_token_here"  # Paste your token after login
```

#### For Admin Endpoints
Edit `admin_endpoints.sh` and update:
```bash
BASE_URL="http://localhost:8080"
ADMIN_TOKEN="your_admin_token_here"  # Paste your admin token
```

## Running the Tests

### Step 1: Test Public Endpoints

```bash
./public_endpoints.sh
```

This will test:
- Phone & Email Signup
- Phone & Email Login
- Vehicle Types
- Fare Estimates
- Surge Pricing
- Service Categories
- Nearby Drivers

### Step 2: Get Authentication Token

1. Run the login endpoints from public_endpoints.sh
2. Copy the `accessToken` from the response
3. Paste it into `auth_endpoints.sh` as the `ACCESS_TOKEN` value

Example response:
```json
{
  "success": true,
  "data": {
    "accessToken": "eyJhbGciOiJIUzI1NiIs...",
    "refreshToken": "def50200...",
    "user": {...}
  }
}
```

### Step 3: Test Authenticated Endpoints

```bash
./auth_endpoints.sh
```

This will test:
- User Profile
- Rider/Driver Operations
- Ride Management
- Wallet Operations
- Service Orders
- Ratings
- Todos

### Step 4: Test Admin Endpoints

```bash
./admin_endpoints.sh
```

This will test:
- Service Creation
- Service Updates
- Service Management

## Testing Specific Endpoints

You can also run individual curl commands from the scripts:

```bash
# Example: Test login
curl -X POST "http://localhost:8080/auth/email/login" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@example.com",
    "password": "securePassword123"
  }'
```

## Common Issues & Solutions

### Issue: "Connection refused"
**Solution:** Make sure your backend server is running on the correct port.

### Issue: "401 Unauthorized"
**Solution:** Your token may have expired. Login again and update the ACCESS_TOKEN.

### Issue: "404 Not Found"
**Solution:** Check that the endpoint path is correct and matches your API routes.

### Issue: UUID/ID not found
**Solution:** Replace placeholder IDs (like "some-uuid-here") with actual IDs from your database.

## Testing Workflow

### For Riders:
1. Run phone signup (rider)
2. Run phone login (rider)
3. Update ACCESS_TOKEN in auth_endpoints.sh
4. Test rider profile endpoints
5. Create a ride request
6. Check ride status

### For Drivers:
1. Run phone signup (driver)
2. Run phone login (driver)
3. Update ACCESS_TOKEN in auth_endpoints.sh
4. Test driver profile endpoints
5. Update driver status to "online"
6. Update location
7. Accept/reject rides

### For Service Providers:
1. Run email signup with role "service_provider"
2. Login and get token
3. Test provider order endpoints
4. Accept/reject service orders
5. Complete services

### For Admin:
1. Run email signup with role "admin"
2. Login and get admin token
3. Update ADMIN_TOKEN in admin_endpoints.sh
4. Create and manage services

## Environment Variables Alternative

Instead of editing the scripts, you can use environment variables:

```bash
# Set environment variables
export BASE_URL="http://localhost:8080"
export ACCESS_TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VySWQiOiI4YmJjY2RkYS0zMGVmLTQ4MWItOTEyMS0zYjZlNTBlYTQxMjUiLCJyb2xlIjoicmlkZXIiLCJleHAiOi0yODEyNTM1MDgwLCJuYmYiOjE3NjMzMTUzMzksImlhdCI6MTc2MzMxNTMzOX0.fkChfiiAgcTkH4OOA3BzjyAXEdZVqFBqrqX0PPtJ8Ko"
export ADMIN_TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VySWQiOiI4MzUyNTUwOC02ODA4LTQ2NDktYTRmMy05NzNhYTU0OTQ5ZTIiLCJyb2xlIjoiZHJpdmVyIiwiZXhwIjo3Mzk4MDkzODQyLCJuYmYiOjE3NjMzMTUzNDAsImlhdCI6MTc2MzMxNTM0MH0.botfwySeC4lULY_dDAS5-xKj5zrrYnBS4Nv0lhuXSEk"

# Then modify scripts to use them:
# BASE_URL="${BASE_URL:-http://localhost:8080}"
```

## Response Formatting

For better readability, pipe curl output to jq:

```bash
curl -X GET "http://localhost:8080/vehicles/types" | jq
```

Install jq:
```bash
# Ubuntu/Debian
sudo apt-get install jq

# Mac
brew install jq
```

## Saving Responses

To save responses to files:

```bash
./public_endpoints.sh > public_test_results.txt 2>&1
./auth_endpoints.sh > auth_test_results.txt 2>&1
./admin_endpoints.sh > admin_test_results.txt 2>&1
```

## API Endpoint Categories

### Authentication (6 endpoints)
- Phone signup/login for riders and drivers
- Email signup/login for all roles
- Token refresh
- Logout

### Riders (3 endpoints)
- Profile management
- Statistics
- Address management

### Drivers (8 endpoints)
- Profile and vehicle management
- Status updates (online/offline)
- Location tracking
- Dashboard and wallet

### Rides (10 endpoints)
- Create ride requests
- Accept/reject rides
- Track ride lifecycle
- Cancel rides

### Wallet (11 endpoints)
- Balance management
- Add/withdraw funds
- Transfers and holds
- Transaction history

### Home Services (14 endpoints)
- Browse categories and services
- Create service orders
- Provider order management
- Order lifecycle

### Pricing (3 endpoints)
- Fare estimates
- Surge pricing
- Dynamic pricing zones

### Vehicle Types (3 endpoints)
- List available vehicle types
- Vehicle type details

### Tracking (3 endpoints)
- Real-time driver location
- Nearby driver search

### Ratings (1 endpoint)
- Rate completed services

### Todos (5 endpoints)
- CRUD operations for todos

## Notes

- All timestamps should be in RFC3339 format
- Coordinates use standard lat/lon format
- Money amounts are in decimal format
- All responses follow the standard response format with `success`, `data`, `message`, and `meta` fields
- Some endpoints require specific user roles (rider, driver, admin, service_provider)

## Support

If you encounter issues:
1. Check server logs
2. Verify request payload matches the API schema
3. Ensure proper authentication token
4. Verify user has correct role/permissions