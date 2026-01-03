import http from 'k6/http';
import { check, sleep } from 'k6';
import { Trend, Counter, Rate } from 'k6/metrics';

const BASE_URL = __ENV.BASE_URL || 'https://api.pittapizzahusrev.be/go';
const AUTH_TOKEN = __ENV.AUTH_TOKEN || ''; // Set via -e AUTH_TOKEN=yourtoken for auth-protected endpoints

const requestDuration = new Trend('http_request_duration_ms');
const requestCount = new Counter('http_requests_total');
const errorRate = new Rate('errors'); // Custom metric for unexpected failures

export const options = {
  stages: [
    { duration: '1m', target: 10 },
    { duration: '1m', target: 25 },
    { duration: '1m', target: 50 },
    { duration: '2m', target: 100 },
    { duration: '1m', target: 0 },
  ],
  thresholds: {
    'http_request_duration_ms': ['p(95)<500', 'p(99)<1000'],
    'http_requests_total': ['count>100'],
    'errors': ['rate<0.1'],  // Only unexpected errors <10%
  },
};

export default function () {
  let endpoints = [
    '/api/v1/homeservices/categories',  // Requires auth; expect 401 without token
    '/api/v1/serviceproviders',         // May return 404 or similar
    '/api/v1/health',                   // Should be 200
  ];

  let endpoint = endpoints[Math.floor(Math.random() * endpoints.length)];

  let res = http.get(`${BASE_URL}${endpoint}`, {
    headers: {
      'Content-Type': 'application/json',
      Authorization: `Bearer ${AUTH_TOKEN}`,  // Add for auth if available
    },
    tags: { name: endpoint },
  });

  requestCount.add(1);
  requestDuration.add(res.timings.duration);

  // Flexible status check: Allow expected 4xx for certain endpoints
  const isStatusOk = (r) => {
    if (endpoint === '/api/v1/health') {
      return r.status === 200;
    } else {
      return [200, 401, 404].includes(r.status);  // Accept 401/404 as "ok" for protected/missing
    }
  };

  const statusCheck = check(res, {
    'status ok': (r) => isStatusOk(r),
    'response time < 500ms': (r) => r.timings.duration < 500,
  });

  // Only count as error if status check fails (unexpected status)
  errorRate.add(!statusCheck);

  sleep(1);
}