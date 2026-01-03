import http from 'k6/http';
import { check, sleep } from 'k6';
import { Trend, Counter } from 'k6/metrics';

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

const requestDuration = new Trend('http_request_duration_ms');
const requestCount = new Counter('http_requests_total');

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
  },
};

export default function () {
  let endpoints = [
    '/api/v1/homeservices/categories',
    '/api/v1/serviceproviders',
    '/api/v1/health',
  ];

  let endpoint = endpoints[Math.floor(Math.random() * endpoints.length)];

  let res = http.get(`${BASE_URL}${endpoint}`, {
    tags: { name: endpoint },
  });

  requestCount.add(1);
  requestDuration.add(res.timings.duration);

  check(res, {
    'status ok': (r) => r.status < 400,
    'response time < 500ms': (r) => r.timings.duration < 500,
  });

  sleep(1);
}
