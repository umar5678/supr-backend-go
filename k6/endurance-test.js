import http from 'k6/http';
import { check, sleep, group } from 'k6';
import { Rate } from 'k6/metrics';

const BASE_URL = __ENV.BASE_URL || 'https://api.pittapizzahusrev.be/go';

const errorRate = new Rate('errors');

export const options = {
  stages: [
    // Ramp up
    { duration: '5m', target: 50 },
    // Long steady load to test for memory leaks and stability
    { duration: '30m', target: 50 },
    // Ramp down
    { duration: '5m', target: 0 },
  ],
  thresholds: {
    'http_req_duration': ['p(95)<500'],
    'http_req_failed': ['rate<0.01'],
    'errors': ['rate<0.01'],
  },
};

export default function () {
  group('multiple endpoint calls', () => {
    // Simulate multiple API calls in a session
    for (let i = 0; i < 5; i++) {
      let res = http.get(`${BASE_URL}/api/v1/homeservices/categories`, {
        tags: { name: `Call ${i + 1}` },
      });

      let isError = res.status >= 400;
      errorRate.add(isError);

      check(res, {
        'status ok': (r) => r.status === 200,
      });

      sleep(Math.random() * 2); // Random sleep between requests
    }
  });

  sleep(5); // Longer think time between iterations
}
