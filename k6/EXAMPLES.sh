#!/usr/bin/env bash
# Comprehensive k6 examples - Copy and modify these commands

# ============================================================
# BASIC COMMANDS
# ============================================================

# 1. Run a simple test
k6 run basic-load-test.js

# 2. Run with custom base URL
k6 run -e BASE_URL=http://api.example.com basic-load-test.js

# 3. Run with auth token
k6 run -e BASE_URL=http://localhost:8080 -e AUTH_TOKEN=your_token realistic-user-journey.js

# 4. Run with custom VUs and duration
k6 run --vus 100 --duration 5m basic-load-test.js

# ============================================================
# OUTPUT & RESULTS
# ============================================================

# 5. Save results to JSON
k6 run -o json=results.json basic-load-test.js

# 6. Save to multiple formats
k6 run -o json=results.json --out csv=results.csv basic-load-test.js

# 7. Pretty print JSON results
k6 run -o json=results.json basic-load-test.js
cat results.json | jq '.'

# 8. Extract specific metrics
cat results.json | jq '.metrics."http_req_duration".values'

# ============================================================
# MONITORING & LOGGING
# ============================================================

# 9. Run with verbose output
k6 run --verbose basic-load-test.js

# 10. Run with debug mode
k6 run -v basic-load-test.js 2>&1 | tee test-debug.log

# 11. Monitor system while testing (separate terminal)
# Linux:
watch -n 1 'free -h; echo "---"; top -b -n1 | head -20'

# Windows PowerShell (separate terminal):
# $watch = { while($true) { Clear-Host; Get-Process | Where-Object {$_.ProcessName -eq 'go'} | Select-Object Name, CPU, Memory; Start-Sleep 1 } }; & $watch

# ============================================================
# SCALING & LOAD PROFILES
# ============================================================

# 12. Quick smoke test (30 seconds, 10 VUs)
k6 run --vus 10 --duration 30s basic-load-test.js

# 13. Sustained load (2 hours, 100 VUs)
k6 run --vus 100 --duration 2h basic-load-test.js

# 14. Ramp up to peak
k6 run --stage "2m:10" --stage "5m:50" --stage "2m:100" --stage "2m:0" basic-load-test.js

# ============================================================
# TEST-SPECIFIC COMMANDS
# ============================================================

# 15. Basic test (recommended first)
k6 run basic-load-test.js

# 16. Realistic user journey
k6 run -e AUTH_TOKEN=token realistic-user-journey.js

# 17. Ramp up test (find breaking point)
k6 run ramp-up-test.js

# 18. Spike test (traffic spike simulation)
k6 run spike-test.js

# 19. Stress test (find maximum capacity - WILL CRASH)
k6 run stress-test.js

# 20. Endurance test (30+ minute stability test)
k6 run endurance-test.js

# ============================================================
# ADVANCED: WITH HELPER SCRIPTS
# ============================================================

# 21. Windows - Run basic test
.\run-k6-tests.bat basic

# 22. Windows - Run realistic journey
.\run-k6-tests.bat realistic

# 23. Windows - Run all tests
.\run-k6-tests.bat all

# 24. Linux/Mac - Make script executable
chmod +x run-k6-tests.sh

# 25. Linux/Mac - Run basic test
./run-k6-tests.sh basic

# 26. Linux/Mac - Run with custom URL
BASE_URL=http://api.example.com ./run-k6-tests.sh realistic

# 27. Linux/Mac - Run all tests
./run-k6-tests.sh all

# ============================================================
# ANALYZING RESULTS
# ============================================================

# 28. Analyze results with Python script
python3 analyze_results.py results.json

# 29. View JSON results (formatted)
python3 -m json.tool results.json | less

# 30. Extract response times
cat results.json | jq '.metrics."http_req_duration"'

# 31. Extract error rate
cat results.json | jq '.metrics."http_req_failed"'

# 32. Compare two test runs
diff results1.json results2.json | less

# ============================================================
# AUTHENTICATION & SETUP
# ============================================================

# 33. Get auth token from your API
TOKEN=$(curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password"}' \
  -s | jq -r '.token')
echo "Token: $TOKEN"

# 34. Test with obtained token
k6 run -e BASE_URL=http://localhost:8080 -e AUTH_TOKEN=$TOKEN realistic-user-journey.js

# 35. Health check before running test
curl -v http://localhost:8080/health

# ============================================================
# DOCKER USAGE (if you prefer)
# ============================================================

# 36. Run k6 in Docker
docker run --rm -v $(pwd):/scripts loadimpact/k6 run /scripts/basic-load-test.js

# 37. Run k6 Docker with network access to local host
docker run --rm --network host -v $(pwd):/scripts loadimpact/k6 run /scripts/basic-load-test.js

# ============================================================
# CONTINUOUS INTEGRATION
# ============================================================

# 38. Export results for CI/CD
k6 run --out json=test-results-$(date +%Y%m%d-%H%M%S).json basic-load-test.js

# 39. Set exit status based on thresholds
k6 run basic-load-test.js
# Exit code 0 = all thresholds passed
# Exit code 64 = some thresholds failed

# 40. Run test and email results (Linux/Mac)
RESULT=$(k6 run -o json=results.json basic-load-test.js)
cat results.json | mail -s "Load Test Results" admin@example.com

# ============================================================
# DEBUGGING & PROFILING
# ============================================================

# 41. Enable verbose HTTP logging
k6 run --http-debug=full basic-load-test.js

# 42. Trace k6 execution
k6 run --log-output=stdout:full basic-load-test.js

# 43. Profile CPU usage (requires local Go installation)
go tool pprof -http=:8081 http://localhost:6060/debug/pprof/profile

# ============================================================
# BATCH TESTING
# ============================================================

# 44. Run multiple tests sequentially
for test in basic-load-test.js ramp-up-test.js spike-test.js; do
  echo "Running $test..."
  k6 run -o json=results-$test.json $test
  sleep 60  # Wait 1 minute between tests
done

# 45. Run tests and generate HTML report (requires grafana-reporter)
k6 run -o json=results.json basic-load-test.js
# Then use k6 results sharing: https://k6.io/docs/results-visualization/

# ============================================================
# PERFORMANCE TUNING
# ============================================================

# 46. Increase file descriptor limit (Linux)
ulimit -n 100000
k6 run stress-test.js

# 47. Disable keep-alive for stress testing
# Edit script: disableKeepAlives: true in http options

# 48. Limit memory usage
k6 run --limit-memory=1GB basic-load-test.js

# ============================================================
# PRODUCTION TESTING (SAFE)
# ============================================================

# 49. Minimal load test on production
k6 run --vus 5 --duration 1m basic-load-test.js

# 50. Daily automated test
# Add to crontab: 0 2 * * * cd /path/to/k6 && k6 run basic-load-test.js > test-$(date +\%Y\%m\%d).log 2>&1

# ============================================================
# NOTES
# ============================================================

# - Always test on staging first before production
# - Start with small VU counts and scale up gradually
# - Monitor backend resources (CPU, RAM, connections) while testing
# - Keep baseline results for comparison
# - Check backend logs for errors after each test
# - Use realistic think times between requests
# - Test during off-peak hours when possible
# - Run tests multiple times for consistency
# - Consider geographical distribution of load (if applicable)

# For more examples, see:
# - README.md (comprehensive guide)
# - QUICK-REFERENCE.md (commands cheat sheet)
# - TESTING-STRATEGY.md (testing workflow)
# - https://k6.io/docs/examples/
