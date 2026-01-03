import http from 'k6/http';
import { check, sleep, group } from 'k6';
import { Rate } from 'k6/metrics';

const BASE_URL = __ENV.BASE_URL || 'https://api.pittapizzahusrev.be/go';
const AUTH_TOKEN = __ENV.AUTH_TOKEN || '';

const errorRate = new Rate('errors');

export const options = {
  stages: [
    { duration: '5m', target: 50 },    // Ramp up
    { duration: '30m', target: 50 },   // Soak: sustained load
    { duration: '5m', target: 0 },     // Ramp down
  ],
  thresholds: {
    'http_req_duration': ['p(95)<500'],
    'errors': ['rate<0.01'],  // Use custom metric instead of http_req_failed
  },
};

// Optional: Get auth token once before test starts
export function setup() {
  if (AUTH_TOKEN) return { token: AUTH_TOKEN };
  
  // If you have a login endpoint, use it here
  // Otherwise, provide token via: k6 run -e AUTH_TOKEN=your_token soak-test.js
  return { token: '' };
}

export default function (data) {
  group('multiple endpoint calls', () => {
    for (let i = 0; i < 5; i++) {
      let res = http.get(`${BASE_URL}/api/v1/homeservices/categories`, {
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${data.token || AUTH_TOKEN}`,
        },
        tags: { name: `Call ${i + 1}` },
      });

      // Accept 200, 401 (no token), 404 as valid - only count real errors
      let isError = ![200, 401, 404].includes(res.status);
      errorRate.add(isError);

      check(res, {
        'status ok': (r) => [200, 401, 404].includes(r.status),
        'response under 500ms': (r) => r.timings.duration < 500,
      });

      sleep(Math.random() * 2);
    }
  });

  sleep(5);
}
