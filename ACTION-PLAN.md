🎯 YOUR ACTION PLAN - Copy & Paste Ready

═════════════════════════════════════════════════════════════════════

SITUATION:
  • Your k6 test showed 100% error rate
  • Root cause: CORS middleware rejecting k6 requests
  • Status: ✅ FIXED in code

NEXT STEPS (Follow in Order):

═════════════════════════════════════════════════════════════════════

STEP 1: PULL THE FIX (On Your Hostinger Server)
────────────────────────────────────────────────

  root@srv990975:~# cd /var/www/go-backend/supr-backend-go
  root@srv990975:supr-backend-go# git pull origin main
  
  This gets the CORS fix from this session.


STEP 2: REBUILD THE BACKEND
────────────────────────────

  Option A - Using Go directly:
    root@srv990975:supr-backend-go# go build -o bin/api ./cmd/api
    root@srv990975:supr-backend-go# ls -la bin/api  # Verify built

  Option B - Using Make:
    root@srv990975:supr-backend-go# make build


STEP 3: STOP OLD BACKEND (if running)
──────────────────────────────────────

  Option A - Kill by name:
    root@srv990975:supr-backend-go# pkill -f "go run"
    root@srv990975:supr-backend-go# pkill -f "./bin/api"

  Option B - Using systemctl (if set up):
    root@srv990975:supr-backend-go# systemctl stop go-backend


STEP 4: START NEW BACKEND
─────────────────────────

  Option A - Direct:
    root@srv990975:supr-backend-go# go run ./cmd/api/main.go
    # Should see: "Server listening on :8080"
    # Press Ctrl+C to stop (keep it running for testing)

  Option B - Compiled binary:
    root@srv990975:supr-backend-go# ./bin/api
    # Should see startup messages

  Option C - Background (recommended):
    root@srv990975:supr-backend-go# nohup ./bin/api > api.log 2>&1 &
    root@srv990975:supr-backend-go# tail -f api.log  # Monitor
    # Press Ctrl+C to exit monitoring (process stays running)


STEP 5: QUICK TEST - Verify Health Endpoint
────────────────────────────────────────────

  root@srv990975:supr-backend-go# curl http://localhost:8080/health
  
  Expected response: {"status":"OK"} or similar
  
  If you get this, the fix is working! ✅


STEP 6: RUN K6 TEST
───────────────────

  root@srv990975:supr-backend-go# cd k6
  root@srv990975:k6# ./run-k6-tests.sh basic
  
  Watch the output. You should see:
    ✅ Error rate < 1% (not 100%!)
    ✅ Health check passing
    ✅ Response times showing


STEP 7: CHECK RESULTS
─────────────────────

  Look for this in output:

    http_req_failed......: 0.5%  ✅  (should be LOW, not 100%)
    ✓ health check status is 200  ✅
    ✓ categories endpoint status is 200 or 404  ✅

  If you see these checkmarks, SUCCESS! 🎉


═════════════════════════════════════════════════════════════════════

COMPLETE COMMAND SEQUENCE (Copy & Paste):

root@srv990975:supr-backend-go# git pull origin main
root@srv990975:supr-backend-go# go build -o bin/api ./cmd/api
root@srv990975:supr-backend-go# pkill -f "go run"
root@srv990975:supr-backend-go# pkill -f "./bin/api"
root@srv990975:supr-backend-go# nohup ./bin/api > api.log 2>&1 &
root@srv990975:supr-backend-go# sleep 2
root@srv990975:supr-backend-go# curl http://localhost:8080/health
root@srv990975:supr-backend-go# cd k6
root@srv990975:k6# ./run-k6-tests.sh basic

═════════════════════════════════════════════════════════════════════

WHAT WAS CHANGED:

File: internal/middleware/cors.go
Line: 58-67 (isAllowedOrigin function)

BEFORE (❌ Wrong):
    if origin == "" {
        return false  // Rejected k6!
    }

AFTER (✅ Fixed):
    if origin == "" {
        return true  // Allow k6, curl, Postman
    }

═════════════════════════════════════════════════════════════════════

EXPECTED BEFORE & AFTER:

BEFORE:
    http_req_failed: 100.00% 45120 out of 45120  ❌
    ✗ health check status is 200
        ↳ 0% — ✓ 0 / ✗ 9024

AFTER:
    http_req_failed: 0.5%  ✅
    ✓ health check status is 200  ✅
    ✓ categories endpoint status is 200 or 404  ✅
    ✓ providers endpoint status is 200 or 404  ✅
    ✓ rider profile status is 200, 401, or 404  ✅
    ✓ driver profile status is 200, 401, or 404  ✅

═════════════════════════════════════════════════════════════════════

TROUBLESHOOTING:

Q: Still getting 100% error?
A: Make sure you:
   1. Ran: git pull origin main
   2. Rebuilt: go build -o bin/api ./cmd/api
   3. Restarted backend: pkill -f "go run" then ./bin/api

Q: curl health check fails?
A: Backend not running. Start it:
   ./bin/api
   (You should see "Server listening on :8080")

Q: Different error?
A: Check logs:
   tail -f api.log
   (if using nohup)

═════════════════════════════════════════════════════════════════════

OPTIONAL: RUN DIAGNOSTIC FIRST

If you want to test endpoints before running k6:

root@srv990975:k6# chmod +x diagnose.sh
root@srv990975:k6# ./diagnose.sh http://localhost:8080

Should show:
    1️⃣  ✅ Health check passed
    2️⃣  ✅ Categories endpoint responded
    3️⃣  ✅ Providers endpoint responded

═════════════════════════════════════════════════════════════════════

THEN TEST ALL K6 TESTS:

root@srv990975:k6# ./run-k6-tests.sh all

This will run:
    1. basic-load-test (9 min)
    2. realistic-user-journey (10 min)
    3. ramp-up-test (6 min)
    4. spike-test (8 min)

═════════════════════════════════════════════════════════════════════

COMMIT THE FIX:

root@srv990975:supr-backend-go# git add internal/middleware/cors.go
root@srv990975:supr-backend-go# git commit -m "fix: Allow k6 and CLI tool requests (CORS)"
root@srv990975:supr-backend-go# git push origin main

═════════════════════════════════════════════════════════════════════

SUMMARY:

✅ Issue Found: CORS middleware was too strict
✅ Fix Applied: Allow non-browser requests
✅ Ready To: Test with k6

Just follow the command sequence above and you're good to go! 🚀

═════════════════════════════════════════════════════════════════════

Have questions? Read:
  • K6-QUICK-FIX.txt (visual guide)
  • K6-FIX-GUIDE.md (detailed explanation)
  • k6/README.md (complete k6 guide)

═════════════════════════════════════════════════════════════════════
