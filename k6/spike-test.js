import http from 'k6/http';
import { check, sleep } from 'k6';

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

export const options = {
  stages: [
    // Normal load
    { duration: '2m', target: 30 },
    // Spike to high load
    { duration: '1m', target: 200 },
    // Back to normal
    { duration: '2m', target: 30 },
    // Another spike
    { duration: '1m', target: 150 },
    // Back to normal and ramp down
    { duration: '2m', target: 0 },
  ],
  thresholds: {
    http_req_duration: ['p(95)<1000'],
    http_req_failed: ['rate<0.2'], // Allow higher error rate during spikes
  },
};

export default function () {
  let res = http.get(`${BASE_URL}/api/v1/homeservices/categories`, {
    tags: { name: 'Spike Test' },
  });

  check(res, {
    'status ok': (r) => r.status < 500,
    'response time under 1s': (r) => r.timings.duration < 1000,
  });

  sleep(0.5);
}
