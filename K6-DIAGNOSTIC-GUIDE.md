# 🔍 k6 Endpoint Diagnostic Guide

## The Problem

You're seeing 100% failure rate, but with blazing-fast response times (2-3ms). This means:
- ❌ NOT a backend performance issue
- ❌ NOT a rate limiting issue (those would be 429s)
- ✅ ONE endpoint is returning a non-2xx status code consistently

## The Solution: Run Diagnostics

### Step 1: Run the diagnostic script

```bash
cd k6/

# Local testing
./run-k6-tests.sh diagnose

# Remote testing (Hostinger)
BASE_URL=https://api.pittapizzahusrev.be/go ./run-k6-tests.sh diagnose
```

### Step 2: Look at the console output

The script will print status codes for each endpoint:

```
=== DIAGNOSTIC: Testing all endpoints ===

✓ Health: 200 http://localhost:8080/health
  Response: {"status":"ok","time":"2026-01-03T...

✓ Categories: 404 http://localhost:8080/api/v1/homeservices/categories
  Response: {"error":"not found"}

✓ Providers: 200 http://localhost:8080/api/v1/serviceproviders
  Response: [...]

✓ Rider Profile: 401 http://localhost:8080/api/v1/riders/profile
  Response: {"error":"unauthorized"}

✓ Driver Profile: 200 http://localhost:8080/api/v1/drivers/profile
  Response: [...]
```

### Step 3: Analyze the results

| Status Code | Meaning | Fix |
|---|---|---|
| **200** | ✅ Endpoint works | Keep it |
| **404** | ❌ Endpoint doesn't exist | Fix URL or route |
| **401** | ⚠️ Needs authentication | Add valid token or accept 401 in checks |
| **500** | ❌ Server error | Check backend logs |
| **429** | ⚠️ Rate limited | Increase rate limit or reduce load |

## Quick Fixes

### Option A: Accept Non-2xx Status Codes in Your Test

If an endpoint legitimately returns 401 (unauthenticated) or 404 (not found), update the check:

**File:** `k6/basic-load-test.js`

**OLD (too strict):**
```javascript
check(healthRes, {
  'health check status is 200': (r) => r.status === 200,
});
```

**NEW (tolerant):**
```javascript
check(healthRes, {
  'health check status is 200': (r) => r.status === 200,
});

check(categoriesRes, {
  'categories status is 200 or 404': (r) => [200, 404].includes(r.status),
});

check(riderRes, {
  'rider profile status is 200 or 401 or 404': (r) => [200, 401, 404].includes(r.status),
});
```

### Option B: Fix the Endpoints

If a URL is wrong, find the correct endpoint:

```bash
# Check available routes
curl -i https://api.pittapizzahusrev.be/go/api/v1/

# Or check your code
grep -r "RegisterRoutes\|router.GET\|router.POST" internal/modules/
```

### Option C: Provide Valid Authentication

```bash
# Get a valid token first
curl -X POST https://api.pittapizzahusrev.be/go/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@test.com","password":"password"}'

# Run test with token
AUTH_TOKEN=your_actual_token BASE_URL=https://api.pittapizzahusrev.be/go ./run-k6-tests.sh diagnose
```

## What to Report

After running diagnostics, you should see something like:

```
Status Codes Found:
- /health → 200 ✅
- /api/v1/homeservices/categories → 404 ❌
- /api/v1/serviceproviders → 200 ✅
- /api/v1/riders/profile → 401 ⚠️
- /api/v1/drivers/profile → 200 ✅
```

**Tell me:** Which endpoints are 404, 401, or 500? Then we'll fix the test script.

---

## Expected Outcome

After fixing the checks/endpoints:
- ✅ `http_req_failed: 0%` (no failures)
- ✅ `http_req_duration p(95)=5ms` (still blazing fast)
- ✅ All checks passing
- ✅ Actual load test results you can trust

---

## Pro Tips

1. **Always run diagnostics first** before running full load tests
2. **1 VU = easier debugging** (use `vus: 1` to see what's happening)
3. **Console logs are your friend** in k6 (visible with `k6 run`)
4. **Test locally first** before testing on production

Run the diagnostic and let me know the output! 🚀
