import http from 'k6/http';
import { check, sleep, group } from 'k6';

// Configuration
const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';
const AUTH_TOKEN = __ENV.AUTH_TOKEN || ''; // Set via environment variable

export const options = {
  stages: [
    { duration: '1m', target: 50 },   // Ramp up to 50 VUs over 1 minute
    { duration: '3m', target: 50 },   // Stay at 50 VUs for 3 minutes
    { duration: '1m', target: 100 },  // Ramp up to 100 VUs over 1 minute
    { duration: '3m', target: 100 },  // Stay at 100 VUs for 3 minutes
    { duration: '1m', target: 0 },    // Ramp down to 0 VUs
  ],
  thresholds: {
    http_req_duration: ['p(95)<500', 'p(99)<1000'], // 95% under 500ms, 99% under 1s
    http_req_failed: ['rate<0.1'],                   // Error rate must be below 10%
  },
};

export default function () {
  // Test authentication endpoints
  group('Auth endpoints', () => {
    authTests();
  });

  sleep(1);

  // Test home services endpoints
  group('Home Services endpoints', () => {
    homeServicesTests();
  });

  sleep(1);

  // Test rider endpoints
  group('Rider endpoints', () => {
    riderTests();
  });

  sleep(1);

  // Test driver endpoints
  group('Driver endpoints', () => {
    driverTests();
  });

  sleep(1);
}

function authTests() {
  // Test health check
  let healthRes = http.get(`${BASE_URL}/health`, {
    tags: { name: 'Health Check' },
  });
  check(healthRes, {
    'health check status is 200': (r) => r.status === 200,
  });
}

function homeServicesTests() {
  // Get home services categories
  let servicesRes = http.get(`${BASE_URL}/api/v1/homeservices/categories`, {
    headers: {
      Authorization: `Bearer ${AUTH_TOKEN}`,
      'Content-Type': 'application/json',
    },
    tags: { name: 'Get Categories' },
  });

  check(servicesRes, {
    'categories endpoint status is 200 or 404': (r) => r.status === 200 || r.status === 404,
  });

  // Get service providers
  let providersRes = http.get(`${BASE_URL}/api/v1/serviceproviders`, {
    headers: {
      Authorization: `Bearer ${AUTH_TOKEN}`,
      'Content-Type': 'application/json',
    },
    tags: { name: 'Get Providers' },
  });

  check(providersRes, {
    'providers endpoint status is 200 or 404': (r) => r.status === 200 || r.status === 404,
  });
}

function riderTests() {
  // Get rider profile
  let riderRes = http.get(`${BASE_URL}/api/v1/riders/profile`, {
    headers: {
      Authorization: `Bearer ${AUTH_TOKEN}`,
      'Content-Type': 'application/json',
    },
    tags: { name: 'Get Rider Profile' },
  });

  check(riderRes, {
    'rider profile status is 200, 401, or 404': (r) =>
      r.status === 200 || r.status === 401 || r.status === 404,
  });
}

function driverTests() {
  // Get driver profile
  let driverRes = http.get(`${BASE_URL}/api/v1/drivers/profile`, {
    headers: {
      Authorization: `Bearer ${AUTH_TOKEN}`,
      'Content-Type': 'application/json',
    },
    tags: { name: 'Get Driver Profile' },
  });

  check(driverRes, {
    'driver profile status is 200, 401, or 404': (r) =>
      r.status === 200 || r.status === 401 || r.status === 404,
  });
}
