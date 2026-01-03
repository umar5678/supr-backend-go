# k6 Load Testing Guide for Supr Backend

This directory contains k6 load testing scripts for your Go backend API. k6 is a modern, open-source load testing tool perfect for testing API performance.

## Installation

### On Windows (PowerShell)
```powershell
# Using Chocolatey
choco install k6

# Or download directly from https://github.com/grafana/k6/releases
```

### On Ubuntu/Debian (Hostinger VM)
```bash
sudo gpg -k
sudo apt-key adv --keyserver hkp://keyserver.ubuntu.com:80 --recv-keys C5AD17C747E3415A3642D57D77C6C491D6AC1D69
echo "deb https://dl.k6.io/deb stable main" | sudo tee /etc/apt/sources.list.d/k6.list
sudo apt update
sudo apt install k6
```

### Verify Installation
```bash
k6 version
```

## Test Scripts Overview

### 1. **basic-load-test.js** - Start Here!
A balanced test covering multiple user journeys:
- 50-100 concurrent virtual users (VUs)
- Tests health check, categories, profiles, and wallet endpoints
- Validates response times (<500ms for 95%, <1s for 99%)
- Best for: Initial baseline testing

**Run:**
```bash
k6 run basic-load-test.js
```

**With authentication token:**
```bash
k6 run -e AUTH_TOKEN=your_token_here basic-load-test.js
```

### 2. **realistic-user-journey.js** - Most Realistic
Simulates actual user behavior:
- Browse services → Create order → Track order → Rate service → Check wallet
- 50 concurrent VUs
- Custom metrics for error rate and API duration tracking
- Best for: Understanding real-world performance

**Run:**
```bash
k6 run realistic-user-journey.js
```

### 3. **spike-test.js** - Sudden Traffic Spike
Tests system behavior during unexpected load spikes:
- Rapid increases from 30 → 200 → 150 VUs
- Allows 20% error rate during spikes
- Best for: Understanding breaking points

**Run:**
```bash
k6 run spike-test.js
```

### 4. **stress-test.js** - Find Breaking Point
Gradually increases load until system fails:
- 100 → 500 concurrent VUs
- Identifies maximum capacity
- Best for: Capacity planning

⚠️ **Warning:** This test will likely crash your API. Only run when ready!

**Run:**
```bash
k6 run stress-test.js
```

### 5. **endurance-test.js** - Long-Running Stability
Runs 30 minutes of steady load:
- 50 concurrent VUs for extended period
- Detects memory leaks and resource exhaustion
- Best for: Stability and leak detection

**Run:**
```bash
k6 run endurance-test.js
```

### 6. **ramp-up-test.js** - Progressive Load
Gradually increases to system capacity:
- 10 → 25 → 50 → 100 VUs
- Custom metrics for detailed analysis
- Best for: Identifying performance degradation threshold

**Run:**
```bash
k6 run ramp-up-test.js
```

## Running Tests with Custom Configuration

### Set Base URL
```bash
k6 run -e BASE_URL=http://your-domain.com:8080 basic-load-test.js
```

### Use Custom Duration
```bash
k6 run -e BASE_URL=http://localhost:8080 basic-load-test.js --duration 10m --vus 50
```

### Output Results to JSON
```bash
k6 run -o json=results.json basic-load-test.js
```

### Output Results to CSV
```bash
k6 run --out csv=results.csv basic-load-test.js
```

### Output to Multiple Formats
```bash
k6 run -o json=results.json --out csv=results.csv basic-load-test.js
```

## Monitoring During Tests

### In Another Terminal - Monitor System Resources
```bash
# On Linux
watch -n 1 'free -h; echo "---"; ps aux | grep go'

# On Windows PowerShell
while ($true) { Get-Process | Select-Object Name, CPU, Memory | head -20; Start-Sleep -Seconds 1; cls }
```

### Check API Logs in Real-Time
```bash
# If using Docker
docker logs -f container_name

# If running locally
tail -f your-log-file.log
```

## Understanding Results

### Key Metrics
- **p(95) and p(99)**: 95th and 99th percentile response times
- **RPS**: Requests per second
- **Error Rate**: Percentage of failed requests
- **Duration**: Total time of request

### Sample Output Interpretation
```
data_received..................: 2.5 MB  42 kB/s
data_sent......................: 850 kB  14 kB/s
http_req_blocked...............: avg=10ms    min=0ms     med=1ms     max=150ms   p(90)=20ms   p(95)=40ms
http_req_connecting............: avg=5ms     min=0ms     med=0ms     max=100ms   p(90)=10ms   p(95)=20ms
http_req_duration..............: avg=250ms   min=50ms    med=200ms   max=2s      p(90)=500ms  p(95)=800ms
http_req_failed................: 1.5%   ← Error rate
http_req_receiving.............: avg=20ms    min=1ms     med=10ms    max=500ms   p(90)=40ms   p(95)=80ms
http_req_sending...............: avg=5ms     min=0ms     med=0ms     max=100ms   p(90)=10ms   p(95)=20ms
http_req_tls_handshaking.......: avg=0s      min=0s      med=0s      max=0s      p(90)=0s     p(95)=0s
http_req_waiting...............: avg=220ms   min=40ms    med=180ms   max=1.9s    p(90)=450ms  p(95)=750ms
http_requests..................: 5000   83.33 req/sec
iteration_duration.............: avg=5.2s    min=3.5s    med=5s      max=10s     p(90)=6.5s   p(95)=7.2s
iterations.....................: 1000   16.67 iter/sec
vus............................: 50     min=50  max=50
vus_max........................: 50     min=50  max=50
```

**Good Signs:**
- ✅ p(95) < 500ms
- ✅ http_req_failed < 1%
- ✅ Consistent performance across iterations

**Red Flags:**
- ❌ p(95) > 1000ms
- ❌ http_req_failed > 5%
- ❌ Increasing error rate over time

## Testing Strategy

### Phase 1: Baseline (30 minutes)
1. Run `basic-load-test.js` with 50 VUs
2. Record baseline metrics
3. Monitor server resources (CPU, RAM, disk)

### Phase 2: Stress Testing (1 hour)
1. Run `spike-test.js` to test spikes
2. Run `stress-test.js` to find breaking point
3. Identify maximum safe concurrent users

### Phase 3: Endurance (40+ minutes)
1. Run `endurance-test.js` at 70% of breaking point
2. Monitor for memory leaks
3. Watch for performance degradation

### Phase 4: Real-World Simulation (1 hour)
1. Run `realistic-user-journey.js`
2. Verify all business flows work under load
3. Check database query performance

## Optimization Tips

### If Tests Fail With "Too Many Open Files"
```bash
# Increase file descriptor limit
ulimit -n 10000

# Or in k6 options:
{
  maxRedirects: 5,
  noVUConnectionReuse: true,
}
```

### If Backend CPU Spikes
1. Add caching layer (Redis)
2. Optimize database queries (add indexes)
3. Use connection pooling
4. Consider horizontal scaling

### If Memory Usage Grows
1. Check for goroutine leaks in your code
2. Review database connection handling
3. Monitor for unbounded queues

## Advanced Features

### Custom Headers
```javascript
let headers = {
  'X-Custom-Header': 'value',
  'Authorization': 'Bearer token',
  'User-Agent': 'k6-test/1.0',
};
let res = http.get(url, { headers });
```

### Data File Input
Create `data.csv`:
```
user_id,email
1,user1@test.com
2,user2@test.com
```

```javascript
import { open } from 'k6/io';
let file = open('./data.csv');
```

### Custom Thresholds
```javascript
thresholds: {
  'http_req_duration{static:yes}': ['p(99)<250'],  // Only static assets
  'http_req_duration{api:yes}': ['p(99)<500'],     // Only API calls
  'errors': ['rate<0.1'],                          // Custom metric
}
```

## Troubleshooting

### "Connection refused"
- Ensure backend is running: `curl http://localhost:8080/health`
- Check firewall rules
- Verify BASE_URL environment variable

### "Too many open connections"
- Increase ulimit: `ulimit -n 10000`
- Reduce VUs or duration
- Check backend connection pool settings

### "Out of memory"
- Reduce VUs
- Reduce test duration
- Check for goroutine leaks with `pprof`

### "High error rate"
- Check backend logs for errors
- Verify database connectivity
- Check API endpoint validity
- Review auth token (if used)

## Integration with CI/CD

### GitHub Actions Example
```yaml
name: Load Tests
on: [push]
jobs:
  k6:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - uses: grafana/setup-k6-action@v1
      - run: k6 run k6/basic-load-test.js
```

### GitLab CI Example
```yaml
load_test:
  image: loadimpact/k6:latest
  script:
    - k6 run k6/basic-load-test.js
```

## Resources

- [k6 Official Docs](https://k6.io/docs/)
- [k6 GitHub](https://github.com/grafana/k6)
- [k6 Examples](https://github.com/grafana/k6/tree/master/samples)
- [Performance Testing Best Practices](https://k6.io/blog/)

## Next Steps

1. ✅ Install k6
2. ✅ Start your backend: `go run ./cmd/api/main.go`
3. ✅ Run basic test: `k6 run k6/basic-load-test.js`
4. ✅ Review results
5. ✅ Optimize based on findings
6. ✅ Run endurance test for validation
