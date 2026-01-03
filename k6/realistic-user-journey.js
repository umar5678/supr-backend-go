import http from 'k6/http';
import { check, sleep, group } from 'k6';
import { Rate, Trend } from 'k6/metrics';

const BASE_URL = __ENV.BASE_URL || 'https://api.pittapizzahusrev.be/go';
const AUTH_TOKEN = __ENV.AUTH_TOKEN || '';

// Custom metrics
const errorRate = new Rate('errors');
const apiDuration = new Trend('api_duration');

export const options = {
  stages: [
    { duration: '2m', target: 20 },   // Ramp up to 20 VUs
    { duration: '5m', target: 20 },   // Stay at 20 VUs
    { duration: '2m', target: 50 },   // Ramp up to 50 VUs
    { duration: '5m', target: 50 },   // Stay at 50 VUs
    { duration: '2m', target: 0 },    // Ramp down
  ],
  thresholds: {
    'http_req_duration': ['p(95)<800', 'p(99)<2000'],
    'http_req_failed': ['rate<0.05'],
    'errors': ['rate<0.05'],
  },
};

export default function () {
  // Simulate a realistic user journey
  let hasAuth = AUTH_TOKEN.length > 0;

  group('Browse home services', () => {
    browseServices();
    sleep(2);
  });

  if (hasAuth) {
    group('Create service order', () => {
      createServiceOrder();
      sleep(2);
    });

    group('Track order', () => {
      trackOrder();
      sleep(2);
    });

    group('Rate service', () => {
      rateService();
      sleep(2);
    });
  }

  group('View wallet', () => {
    viewWallet();
    sleep(2);
  });

  sleep(Math.random() * 3); // Random think time between 0-3 seconds
}

function browseServices() {
  let res = http.get(`${BASE_URL}/api/v1/homeservices/categories`, {
    tags: { name: 'Browse Categories' },
  });

  let isError = res.status >= 400;
  errorRate.add(isError);
  apiDuration.add(res.timings.duration);

  check(res, {
    'browse services status ok': (r) => r.status < 400,
    'response time acceptable': (r) => r.timings.duration < 800,
  });
}

function createServiceOrder() {
  let orderPayload = {
    service_id: Math.floor(Math.random() * 100) + 1,
    provider_id: Math.floor(Math.random() * 50) + 1,
    location: {
      latitude: 24.8607 + (Math.random() - 0.5) * 0.1,
      longitude: 67.0011 + (Math.random() - 0.5) * 0.1,
    },
    scheduled_at: new Date(Date.now() + 3600000).toISOString(),
  };

  let res = http.post(`${BASE_URL}/api/v1/orders`, JSON.stringify(orderPayload), {
    headers: {
      Authorization: `Bearer ${AUTH_TOKEN}`,
      'Content-Type': 'application/json',
    },
    tags: { name: 'Create Order' },
  });

  let isError = res.status >= 400;
  errorRate.add(isError);
  apiDuration.add(res.timings.duration);

  check(res, {
    'create order status ok or client error': (r) => r.status < 500,
    'response time acceptable': (r) => r.timings.duration < 1000,
  });

  return res.status === 200 || res.status === 201 ? JSON.parse(res.body) : null;
}

function trackOrder() {
  let orderId = Math.floor(Math.random() * 1000) + 1;

  let res = http.get(`${BASE_URL}/api/v1/orders/${orderId}/tracking`, {
    headers: {
      Authorization: `Bearer ${AUTH_TOKEN}`,
      'Content-Type': 'application/json',
    },
    tags: { name: 'Track Order' },
  });

  let isError = res.status >= 400;
  errorRate.add(isError);
  apiDuration.add(res.timings.duration);

  check(res, {
    'track order endpoint responds': (r) => r.status < 500,
  });
}

function rateService() {
  let orderId = Math.floor(Math.random() * 1000) + 1;
  let ratingPayload = {
    order_id: orderId,
    rating: Math.floor(Math.random() * 5) + 1,
    comment: 'Good service!',
  };

  let res = http.post(`${BASE_URL}/api/v1/ratings`, JSON.stringify(ratingPayload), {
    headers: {
      Authorization: `Bearer ${AUTH_TOKEN}`,
      'Content-Type': 'application/json',
    },
    tags: { name: 'Rate Service' },
  });

  let isError = res.status >= 400;
  errorRate.add(isError);
  apiDuration.add(res.timings.duration);

  check(res, {
    'rate service endpoint responds': (r) => r.status < 500,
  });
}

function viewWallet() {
  let res = http.get(`${BASE_URL}/api/v1/wallet/balance`, {
    headers: {
      Authorization: `Bearer ${AUTH_TOKEN}`,
      'Content-Type': 'application/json',
    },
    tags: { name: 'View Wallet' },
  });

  let isError = res.status >= 400;
  errorRate.add(isError);
  apiDuration.add(res.timings.duration);

  check(res, {
    'wallet endpoint responds': (r) => r.status < 500,
  });
}
