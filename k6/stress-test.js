import http from 'k6/http';
import { check, sleep } from 'k6';

const BASE_URL = __ENV.BASE_URL || 'https://api.pittapizzahusrev.be/go';

export const options = {
  stages: [
    // Gradually increase load until system breaks
    { duration: '5m', target: 100 },
    { duration: '5m', target: 200 },
    { duration: '5m', target: 300 },
    { duration: '5m', target: 400 },
    { duration: '5m', target: 500 },
    // Ramp down
    { duration: '5m', target: 0 },
  ],
  thresholds: {
    http_req_failed: ['rate<0.5'], // Expect higher failure rates during stress
  },
  gracefulStop: '30s',
};

export default function () {
  let res = http.get(`${BASE_URL}/api/v1/homeservices/categories`, {
    tags: { name: 'Stress Test' },
  });

  check(res, {
    'status ok': (r) => r.status < 500,
  });

  sleep(0.1); // Minimal think time for maximum stress
}
