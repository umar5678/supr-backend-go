📊 K6 LOAD TESTING SETUP - COMPLETE ✅

═══════════════════════════════════════════════════════════════

## ✅ WHAT'S BEEN SET UP

Your `k6/` directory now contains:

📂 Test Scripts:
  ✓ basic-load-test.js          - Start here! (9 min, 50-100 VUs)
  ✓ realistic-user-journey.js   - Real user flows (10 min, 50 VUs)
  ✓ ramp-up-test.js             - Find breaking point (6 min, 10→100 VUs)
  ✓ spike-test.js               - Traffic spikes (8 min, 30→200→150 VUs)
  ✓ stress-test.js              - Find max capacity (30 min, 100→500 VUs)
  ✓ endurance-test.js           - Stability check (40 min, 50 VUs)

📚 Documentation:
  ✓ README.md                    - Comprehensive guide with all details
  ✓ QUICK-REFERENCE.md           - Commands cheat sheet
  ✓ TESTING-STRATEGY.md          - Complete testing workflow
  ✓ EXAMPLES.sh                  - 50+ command examples
  ✓ START-HERE.md                - This file

🛠️ Helper Scripts:
  ✓ run-k6-tests.bat             - Windows automation (simple: .\run-k6-tests.bat basic)
  ✓ run-k6-tests.sh              - Linux/Mac automation
  ✓ analyze_results.py           - Parse JSON results

═══════════════════════════════════════════════════════════════

## 🚀 QUICK START (5 MINUTES)

### 1. Install k6

Windows (PowerShell as Admin):
  choco install k6

Linux/Hostinger:
  curl https://dl.k6.io/key.gpg | sudo apt-key add -
  echo "deb https://dl.k6.io/deb stable main" | sudo tee /etc/apt/sources.list.d/k6.list
  sudo apt update && sudo apt install k6

Verify:
  k6 version

### 2. Start Your Backend

In one terminal:
  go run ./cmd/api/main.go

### 3. Run First Test

Windows:
  cd k6
  .\run-k6-tests.bat basic

Linux/Mac:
  cd k6
  chmod +x run-k6-tests.sh
  ./run-k6-tests.sh basic

Or directly:
  k6 run k6/basic-load-test.js

### 4. View Results

Look for output showing:
  ✅ http_req_duration: p(95)=<500ms  ← Want this!
  ✅ http_req_failed: <1%              ← Want this!
  ✅ http_requests: ~80 RPS            ← Good throughput

═══════════════════════════════════════════════════════════════

## 📊 WHICH TEST TO RUN?

┌─────────────────────────────────────────────────────────────┐
│ TEST              │ DURATION │ BEST FOR                      │
├─────────────────────────────────────────────────────────────┤
│ basic             │ 9 min    │ FIRST TEST - baseline metrics │
│ realistic         │ 10 min   │ Real user behavior            │
│ ramp-up           │ 6 min    │ Find breaking point           │
│ spike             │ 8 min    │ Traffic spikes               │
│ stress            │ 30 min   │ Max capacity (will crash!)    │
│ endurance         │ 40 min   │ Memory leaks & stability      │
└─────────────────────────────────────────────────────────────┘

Recommended Order:
  1. basic          (establish baseline)
  2. ramp-up        (find degradation point)
  3. spike          (test resilience)
  4. endurance      (test stability - run overnight)
  5. stress         (optional - will break things)

═══════════════════════════════════════════════════════════════

## 🎯 WHAT RESULTS MEAN

After running a test, you'll see:

  http_req_duration: avg=250ms p(95)=450ms p(99)=800ms
  
  - avg        = average response time (usually less important)
  - p(95)      = 95% of requests are faster than this ← FOCUS ON THIS
  - p(99)      = 99% of requests are faster than this ← ALSO IMPORTANT

  Examples:
    ✅ EXCELLENT:  p(95)=400ms, p(99)=700ms
    ⚠️  OKAY:      p(95)=600ms, p(99)=1000ms
    ❌ POOR:       p(95)=1500ms, p(99)=2000ms

  http_req_failed: 0.8%
  
    ✅ EXCELLENT:  < 1%
    ⚠️  WARNING:   1-5%
    ❌ BAD:        > 5%

═══════════════════════════════════════════════════════════════

## 🛠️ CUSTOMIZE FOR YOUR API

1. Edit any `.js` file to add your endpoints:

   In `basic-load-test.js`, find:
     const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

   Change endpoints to match your API:
     let res = http.get(`${BASE_URL}/api/v1/your-endpoint`);

2. Add authentication if needed:

   Run test with token:
     k6 run -e AUTH_TOKEN=your_token basic-load-test.js

3. Adjust load in `export const options`:

   Change from:
     { duration: '5m', target: 100 }
   
   To custom load:
     { duration: '10m', target: 50 }

═══════════════════════════════════════════════════════════════

## ⚡ COMMON ISSUES & FIXES

Issue: "Connection refused"
Fix: 
  curl http://localhost:8080/health
  Make sure backend is running!

Issue: "High error rate"
Fix:
  - Check backend logs
  - Verify auth token (if using -e AUTH_TOKEN)
  - Check database connection

Issue: "High latency (p95 > 1000ms)"
Fix:
  - Add database indexes
  - Enable Redis caching
  - Check query performance
  - Increase connection pool size

Issue: "Out of memory"
Fix:
  - Reduce VUs: k6 run --vus 25 basic-load-test.js
  - Reduce duration: k6 run --duration 1m basic-load-test.js

═══════════════════════════════════════════════════════════════

## 📈 NEXT STEPS

1. ✅ Install k6 (if not done)
2. ✅ Run: `k6 run k6/basic-load-test.js`
3. ✅ Record baseline metrics
4. ✅ Fix any issues found
5. ✅ Re-run to verify fixes
6. ✅ Test all critical endpoints
7. ✅ Schedule daily tests

═══════════════════════════════════════════════════════════════

## 📚 WHERE TO FIND HELP

For specific commands:
  → See: QUICK-REFERENCE.md

For comprehensive guide:
  → See: README.md

For complete testing workflow:
  → See: TESTING-STRATEGY.md

For 50+ examples:
  → See: EXAMPLES.sh

For k6 official docs:
  → https://k6.io/docs/

═══════════════════════════════════════════════════════════════

## ⚠️ IMPORTANT NOTES

1. Same Machine Testing:
   - Backend and k6 on same 16GB VM = very fast loopback
   - Results will show faster than real-world
   - For realistic results: run k6 from your laptop pointing to Hostinger IP

2. Database Load:
   - k6 stresses your API, which stresses your database
   - Watch PostgreSQL connection pool
   - May need to increase: max_connections in postgresql.conf

3. Stress Test Warning:
   - stress-test.js WILL crash your API
   - Only run when ready
   - Good for capacity planning
   - DON'T run on production!

4. Consistent Results:
   - Run tests 2-3 times each
   - Network/system can cause variations
   - Average the results

═══════════════════════════════════════════════════════════════

## ✅ CHECKLIST

Before First Test:
  ☐ k6 installed (k6 version shows version)
  ☐ Backend running (curl http://localhost:8080/health)
  ☐ Open terminal in k6 directory

First Test:
  ☐ Run: k6 run basic-load-test.js
  ☐ Wait for completion
  ☐ Write down p(95) and error rate

After First Test:
  ☐ Review results
  ☐ Check if p(95) < 500ms
  ☐ Check if error rate < 1%
  ☐ Run 2-3 more times for consistency

═══════════════════════════════════════════════════════════════

## 🎓 KEY METRICS TO REMEMBER

p(95) response time
  ↳ Most important metric for user experience
  ↳ Want: < 500ms
  ↳ Acceptable: < 1000ms
  ↳ Bad: > 1000ms

Error rate
  ↳ Percentage of requests that failed
  ↳ Want: < 1%
  ↳ Acceptable: 1-5%
  ↳ Bad: > 5%

Requests per second (RPS)
  ↳ How many requests your API handles
  ↳ Watch for: consistency
  ↳ Bad: RPS dropping over time = degradation

Virtual Users (VUs)
  ↳ Concurrent users simulated
  ↳ Start: 10-50
  ↳ Scale: 50-100 for normal load
  ↳ Stress: 100+ for finding breaking point

═══════════════════════════════════════════════════════════════

READY TO START? Run this:

  Windows:    .\run-k6-tests.bat basic
  Linux/Mac:  ./run-k6-tests.sh basic
  Or direct:  k6 run k6/basic-load-test.js

═══════════════════════════════════════════════════════════════

Questions? Check the documentation files or k6 docs at:
https://k6.io/docs/

Good luck with your load testing! 🚀
