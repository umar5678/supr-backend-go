# Quick Reference Guide for k6 Load Testing

## Installation Quick Start

### Windows (PowerShell)
```powershell
# Using Chocolatey
choco install k6

# Verify
k6 version
```

### Linux/Hostinger
```bash
sudo apt-key adv --keyserver hkp://keyserver.ubuntu.com:80 --recv-keys C5AD17C747E3415A3642D57D77C6C491D6AC1D69
echo "deb https://dl.k6.io/deb stable main" | sudo tee /etc/apt/sources.list.d/k6.list
sudo apt update && sudo apt install k6
```

## Quick Commands

### Run a Test
```bash
# Basic run
k6 run basic-load-test.js

# With environment variables
k6 run -e BASE_URL=http://localhost:8080 -e AUTH_TOKEN=token basic-load-test.js

# Output JSON results
k6 run -o json=results.json basic-load-test.js

# Custom VUs and duration
k6 run --vus 100 --duration 5m basic-load-test.js
```

### Using Helper Scripts

**Windows:**
```powershell
# Run basic test
.\run-k6-tests.bat basic

# Run with custom URL
$env:BASE_URL="http://api.example.com"; .\run-k6-tests.bat realistic
```

**Linux/Mac:**
```bash
# Make executable
chmod +x run-k6-tests.sh

# Run basic test
./run-k6-tests.sh basic

# Run with custom URL
BASE_URL=http://api.example.com ./run-k6-tests.sh realistic
```

## Test Selection Guide

| Test | Duration | VUs | Best For | Command |
|------|----------|-----|----------|---------|
| **basic** | 9 min | 50-100 | First test, baseline | `k6 run basic-load-test.js` |
| **realistic** | 10 min | 50 | Real user flows | `k6 run realistic-user-journey.js` |
| **ramp** | 6 min | 10→100 | Degradation point | `k6 run ramp-up-test.js` |
| **spike** | 8 min | 30→200→150 | Traffic spikes | `k6 run spike-test.js` |
| **stress** | 30 min | 100→500 | Max capacity | `k6 run stress-test.js` ⚠️ |
| **endurance** | 40 min | 50 | Memory leaks | `k6 run endurance-test.js` |

## Reading Results

### Key Thresholds to Monitor

```
✅ GOOD                    ⚠️  WARNING              ❌ BAD
p(95) < 500ms             p(95) 500-1000ms        p(95) > 1000ms
p(99) < 1000ms            p(99) 1000-2000ms       p(99) > 2000ms
Error rate < 1%           Error rate 1-5%         Error rate > 5%
RPS constant              RPS declining           RPS collapsing
```

### Example Good Results
```
http_req_duration: avg=250ms p(95)=450ms p(99)=800ms
http_req_failed: 0.5%
http_requests: 5000 req/sec ✅
```

### Example Problem Results
```
http_req_duration: avg=2s p(95)=3.5s p(99)=5s
http_req_failed: 15%
Timeouts increasing over time ❌
```

## Troubleshooting

| Problem | Solution |
|---------|----------|
| "Connection refused" | `curl http://localhost:8080/health` - verify backend is running |
| "Too many open files" | `ulimit -n 10000` on Linux |
| "Out of memory" | Reduce VUs: `k6 run --vus 25 basic-load-test.js` |
| "High error rate" | Check backend logs, verify auth token, check endpoints |
| "Very slow response" | Backend overloaded - reduce VUs or check database |

## Performance Testing Workflow

1. **Baseline** (10 min)
   ```bash
   k6 run basic-load-test.js
   ```
   Record all metrics for comparison

2. **Identify Issues** (20 min)
   ```bash
   k6 run ramp-up-test.js
   ```
   Find where performance degrades

3. **Spike Resilience** (10 min)
   ```bash
   k6 run spike-test.js
   ```
   Ensure graceful handling of traffic spikes

4. **Long-term Stability** (40 min)
   ```bash
   k6 run endurance-test.js
   ```
   Check for leaks and degradation

5. **Max Capacity** (30 min) - *Optional*
   ```bash
   k6 run stress-test.js
   ```
   Find breaking point (for planning)

## Tips for Better Tests

1. **Run from different machine** for realistic network latency
2. **Monitor system resources** while testing:
   ```bash
   # Linux
   watch -n 1 'free -h; ps aux | grep go'
   
   # Windows
   Get-Process | Select-Object Name, CPU, Memory
   ```

3. **Check backend logs** for errors
4. **Gradual load increase** is safer than sudden spikes
5. **Run multiple times** - results can vary

## Results Analysis

### JSON Results Inspection
```bash
# View in browser (online)
# Upload to https://k6.io/docs/results-visualization/

# View summary
cat results.json | grep -E '"(metric|name|value"'

# Pretty print with jq
cat results.json | jq '.data.samples[] | select(.metric == "http_req_duration")'
```

## Common Metrics Explained

- **http_req_duration**: Total time for HTTP request
- **http_req_failed**: Requests with status >= 400
- **http_requests**: Total requests made
- **iteration_duration**: Time per test iteration
- **vus**: Concurrent virtual users
- **p(95)**: 95th percentile (95% of requests faster than this)
- **p(99)**: 99th percentile (99% of requests faster than this)

## Next Steps

1. ✅ Install k6
2. ✅ Start backend: `go run ./cmd/api/main.go`
3. ✅ Run: `k6 run k6/basic-load-test.js`
4. ✅ Analyze results
5. ✅ Optimize based on findings
6. ✅ Schedule regular load tests

## Resources

- [k6 Docs](https://k6.io/docs/)
- [k6 GitHub](https://github.com/grafana/k6)
- [REST API Load Testing](https://k6.io/docs/examples/http-requests/)
- [Thresholds & Checks](https://k6.io/docs/using-k6/thresholds/)
