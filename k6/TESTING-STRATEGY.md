# k6 Load Testing - Complete Setup Guide for Supr Backend

## Overview

This directory contains everything you need to load test your Careem-like home services API using **k6**, a modern, open-source load testing tool built in Go.

### Why k6?
- ✅ **Free & Open Source** - No expensive paid tiers
- ✅ **Developer Friendly** - JavaScript-based scripts
- ✅ **Highly Efficient** - Written in Go, lightweight on resources
- ✅ **Perfect for API Testing** - Built-in HTTP, auth, checks, thresholds
- ✅ **Industry Standard** - Widely recommended in 2025/2026

---

## 📋 Quick Start (5 minutes)

### Step 1: Install k6

**Windows (PowerShell as Admin):**
```powershell
choco install k6
```

**Linux/Hostinger (Ubuntu/Debian):**
```bash
sudo gpg -k
sudo apt-key adv --keyserver hkp://keyserver.ubuntu.com:80 --recv-keys C5AD17C747E3415A3642D57D77C6C491D6AC1D69
echo "deb https://dl.k6.io/deb stable main" | sudo tee /etc/apt/sources.list.d/k6.list
sudo apt update
sudo apt install k6
```

**Verify:**
```bash
k6 version
```

### Step 2: Start Your Backend

```bash
# In one terminal
go run ./cmd/api/main.go
```

Ensure it's running on `http://localhost:8080` (or adjust BASE_URL)

### Step 3: Run Your First Test

**Windows:**
```powershell
cd k6
.\run-k6-tests.bat basic
```

**Linux/Mac:**
```bash
cd k6
chmod +x run-k6-tests.sh
./run-k6-tests.sh basic
```

**Or directly:**
```bash
k6 run k6/basic-load-test.js
```

### Step 4: Review Results

Results will show:
- ✅ Requests completed
- ⏱️ Response times (p95, p99, avg)
- ❌ Error rate
- 📊 RPS (Requests Per Second)

---

## 📁 What's Included

### Test Scripts (`.js` files)

| File | Duration | Load Pattern | Use Case |
|------|----------|--------------|----------|
| `basic-load-test.js` | 9 min | 50→100 VUs | **Start here** - Baseline testing |
| `realistic-user-journey.js` | 10 min | 50 VUs | Realistic browsing → ordering flow |
| `ramp-up-test.js` | 6 min | 10→100 VUs | Find performance degradation point |
| `spike-test.js` | 8 min | 30→200→150 VUs | Test spike resilience |
| `stress-test.js` | 30 min | 100→500 VUs | Find breaking point ⚠️ |
| `endurance-test.js` | 40 min | 50 VUs constant | Detect memory leaks |

### Documentation

- **README.md** - Comprehensive guide with examples
- **QUICK-REFERENCE.md** - Commands and tips cheat sheet
- **TESTING-STRATEGY.md** - Detailed testing workflow (this file)

### Helper Scripts

- **run-k6-tests.bat** - Windows automation script
- **run-k6-tests.sh** - Linux/Mac automation script
- **analyze_results.py** - Parse and analyze JSON results

---

## 🎯 Testing Strategy (4+ hours total)

### Phase 1: Baseline Testing (30-40 minutes)

**Objective:** Establish performance baseline

```bash
# Terminal 1: Start backend
go run ./cmd/api/main.go

# Terminal 2: Run test
k6 run k6/basic-load-test.js

# Terminal 3 (optional): Monitor resources
watch -n 1 'free -h && ps aux | grep go'  # Linux
Get-Process | Select Name, CPU, Memory    # Windows
```

**What to record:**
- p(95) response time
- p(99) response time
- Error rate
- RPS (Requests per second)
- Peak CPU/Memory usage

**Acceptance Criteria:**
```
✅ p(95) < 500ms
✅ p(99) < 1000ms
✅ Error rate < 1%
✅ No timeouts
```

### Phase 2: Identify Breaking Point (45 minutes)

**Objective:** Find where performance degrades

```bash
k6 run k6/ramp-up-test.js
```

**Chart the results:**
```
VUs vs Response Time:
  10 VUs  → 250ms (excellent)
  25 VUs  → 280ms (excellent)
  50 VUs  → 320ms (good)
  75 VUs  → 450ms (acceptable)
  100 VUs → 750ms (degrading) ← Point of concern?
  150 VUs → 2000ms (poor)     ← Degradation begins
```

**Action:** Note the VU count where p95 exceeds 500ms

### Phase 3: Traffic Spike Resilience (15 minutes)

**Objective:** Ensure graceful handling of sudden spikes

```bash
k6 run k6/spike-test.js
```

**Expected behavior:**
- Brief latency spike (normal)
- Quick recovery (within 1-2 seconds)
- Minimal errors (<5%)

**Red flags:**
- ❌ Cascading failures
- ❌ Recovery takes >30 seconds
- ❌ Error rate stays >20%

### Phase 4: Long-Term Stability (40+ minutes)

**Objective:** Detect memory leaks and resource issues

```bash
# Monitor first
watch -n 5 'free -h'

# Then run test
k6 run k6/endurance-test.js
```

**Watch for:**
- 📈 Steadily increasing memory usage → **LEAK**
- 📉 Response times slowly degrading → **LEAK**
- 📊 Consistent metrics → **HEALTHY**

**Memory check:**
```
Start:  8.5 GB used
After 10 min:  8.6 GB used  ✅ (0.1 GB increase)
After 20 min:  8.8 GB used  ✅ (0.2 GB total)
After 30 min:  9.2 GB used  ✅ (0.3 GB increase) - slight slope

After 10 min:  8.5 GB used
After 20 min:  9.5 GB used  ⚠️ (1 GB increase)
After 30 min: 11.0 GB used  ❌ MEMORY LEAK DETECTED
```

### Phase 5: Maximum Capacity (30 minutes) - OPTIONAL

**Objective:** Find absolute breaking point

⚠️ **WARNING:** This test WILL crash your API!

```bash
k6 run k6/stress-test.js
```

**This is useful for:**
- Capacity planning
- Budget estimation (cloud resources)
- Identifying bottlenecks

**DO NOT run on production!**

---

## 🔍 Understanding Results

### Sample Test Output

```
data_received..................: 2.5 MB   42 kB/s
data_sent......................: 850 kB   14 kB/s
http_req_blocked...............: avg=10ms    p(95)=20ms
http_req_connecting............: avg=5ms     p(95)=10ms
http_req_duration..............: avg=250ms   p(95)=450ms  p(99)=800ms    ✅ EXCELLENT
http_req_failed................: 0.5%                                   ✅ GOOD
http_req_receiving.............: avg=20ms    p(95)=40ms
http_req_sending...............: avg=5ms     p(95)=10ms
http_req_tls_handshaking.......: avg=0s      p(95)=0s
http_req_waiting...............: avg=220ms   p(95)=400ms
http_requests..................: 5000      83 req/sec
iteration_duration.............: avg=5.2s   p(95)=6.5s
iterations.....................: 1000      17 iter/sec
vus............................: 50        min=50  max=50
vus_max........................: 50        min=50  max=50
```

### Key Metrics Interpretation

| Metric | Good | Warning | Bad |
|--------|------|---------|-----|
| **p(95)** | <500ms | 500-1000ms | >1000ms |
| **p(99)** | <1s | 1-2s | >2s |
| **Error rate** | <1% | 1-5% | >5% |
| **RPS** | Consistent | Declining | Collapsing |

### Performance Rating

```
Score Calculation:
- p(95) < 500ms     → 30 points
- Error rate < 1%   → 30 points
- No timeouts       → 20 points
- Stable metrics    → 20 points

90+ points  → ✅ EXCELLENT - Ready for production
70-90 points → ⚠️  GOOD - Monitor and optimize
<70 points  → ❌ NEEDS WORK - Investigate issues
```

---

## 🛠️ Optimization Workflow

When tests reveal issues:

### High Error Rate (>5%)
1. ✅ Check backend logs: `tail -f app.log`
2. ✅ Verify auth token if using: `-e AUTH_TOKEN=valid_token`
3. ✅ Check endpoint accessibility: `curl http://localhost:8080/health`
4. ✅ Review database connection pool
5. ✅ Check for connection timeouts

### Slow Response Times (p95 > 1s)
1. ✅ Add database indexes
2. ✅ Enable caching (Redis)
3. ✅ Optimize SQL queries (`EXPLAIN ANALYZE`)
4. ✅ Increase database connection pool
5. ✅ Check for N+1 queries

### Memory Growth
1. ✅ Check for goroutine leaks: `import _ "net/http/pprof"`
2. ✅ Review database connection handling
3. ✅ Check for unbounded queues/channels
4. ✅ Monitor garbage collection

### CPU Maxing Out
1. ✅ Profile with `pprof`: `go tool pprof http://localhost:6060/debug/pprof/profile`
2. ✅ Look for hot functions
3. ✅ Optimize algorithms
4. ✅ Consider horizontal scaling

---

## 📊 Analyzing Results

### View Raw JSON
```bash
# Pretty print
cat k6-results/basic-load-test_*.json | jq '.'

# Extract specific metrics
cat k6-results/basic-load-test_*.json | jq '.metrics."http_req_duration".values'

# Python analysis
python3 analyze_results.py k6-results/basic-load-test_*.json
```

### Save Results for Comparison

```bash
# After test, copy results
cp k6-results/basic-load-test_*.json baseline-20240103.json

# After optimization, compare
diff baseline-20240103.json k6-results/basic-load-test_*.json
```

---

## 🚀 Advanced Scenarios

### Load Test with Authentication

```bash
# Get auth token first
TOKEN=$(curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password"}' \
  | jq -r '.token')

# Run with token
k6 run -e BASE_URL=http://localhost:8080 -e AUTH_TOKEN=$TOKEN realistic-user-journey.js
```

### Load Test Different Endpoints

**Create custom-endpoints.js:**
```javascript
import http from 'k6/http';
import { sleep } from 'k6';

export const options = {
  stages: [
    { duration: '2m', target: 50 },
    { duration: '5m', target: 50 },
    { duration: '2m', target: 0 },
  ],
};

export default function () {
  // Test your specific endpoints
  http.get('http://localhost:8080/api/v1/your-endpoint');
  sleep(1);
}
```

### Run from Another Machine (Realistic Network)

Since your backend and k6 are on the same machine, tests measure loopback (very fast, unrealistic).

**Better approach:**
1. Install k6 on your laptop
2. Point to your Hostinger public IP/domain
3. More realistic network latency

```bash
# From your laptop
k6 run -e BASE_URL=http://your-domain.com basic-load-test.js
```

---

## 📈 Continuous Load Testing

### Schedule Regular Tests

**Linux Cron:**
```bash
# Run daily at 2 AM
0 2 * * * cd /home/user/supr-backend-go/k6 && k6 run basic-load-test.js > test-results-$(date +\%Y\%m\%d).log 2>&1
```

**GitHub Actions:**
```yaml
name: Daily Load Test
on:
  schedule:
    - cron: '0 2 * * *'
jobs:
  load-test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - uses: grafana/setup-k6-action@v1
      - run: k6 run k6/basic-load-test.js
```

---

## 🎓 Tips & Best Practices

### ✅ DO

- ✅ Start with small VU counts (10-50)
- ✅ Gradually increase load
- ✅ Monitor system resources while testing
- ✅ Run tests multiple times for consistency
- ✅ Keep baseline results for comparison
- ✅ Test during low-traffic periods
- ✅ Check logs after each test

### ❌ DON'T

- ❌ Run stress tests on production
- ❌ Use unrealistic user journeys
- ❌ Ignore high error rates
- ❌ Test without monitoring resources
- ❌ Change code while testing
- ❌ Run tests from the same server (use different machine for realism)

---

## 🐛 Troubleshooting

### "Connection refused"
```bash
# Verify backend is running
curl http://localhost:8080/health

# Check if port is correct
netstat -tuln | grep 8080  # Linux
netstat -ano | findstr :8080  # Windows
```

### "Too many open connections"
```bash
# Increase file descriptor limit
ulimit -n 10000

# Or reduce VUs
k6 run --vus 25 basic-load-test.js
```

### "Out of memory"
```bash
# Monitor memory
free -h  # Linux
Get-Process | Sort Memory | tail -20  # Windows

# Reduce load
k6 run --vus 10 --duration 1m basic-load-test.js
```

### "High latency/errors only in load test"
- Database connection pool too small
- Missing database indexes
- Network bandwidth bottleneck
- Backend goroutine limit reached

---

## 📚 Resources

- [k6 Official Documentation](https://k6.io/docs/)
- [k6 GitHub Repository](https://github.com/grafana/k6)
- [HTTP Load Testing Best Practices](https://k6.io/docs/examples/http-requests/)
- [k6 Thresholds & Checks](https://k6.io/docs/using-k6/thresholds/)

---

## ✅ Checklist: Complete Load Testing

- [ ] Install k6
- [ ] Run `basic-load-test.js`
- [ ] Record baseline metrics
- [ ] Run `ramp-up-test.js` to find breaking point
- [ ] Run `spike-test.js` to test spikes
- [ ] Run `endurance-test.js` for stability
- [ ] Analyze results with `analyze_results.py`
- [ ] Create performance report
- [ ] Optimize based on findings
- [ ] Re-test after optimization
- [ ] Schedule regular load tests

---

## Questions?

- Check `README.md` for comprehensive guide
- Check `QUICK-REFERENCE.md` for commands
- Review test script comments for customization
- Check k6 docs: https://k6.io/docs/

**Happy load testing! 🚀**
