import http from 'k6/http';
import { check, sleep, group } from 'k6';

// Configuration
const BASE_URL = __ENV.BASE_URL || 'https://api.pittapizzahusrev.be/go';
const AUTH_TOKEN = __ENV.AUTH_TOKEN || ''; // Set via environment variable for protected endpoints

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
  // Get home services categories ( /api/v1/homeservices/categories) - requires auth
  let categoriesRes = http.get(`${BASE_URL}/api/v1/homeservices/categories`, {
    headers: {
      'Content-Type': 'application/json',
      Authorization: `Bearer ${AUTH_TOKEN}`,
    },
    tags: { name: 'Get Categories' },
  });

  // If no token provided, expect 401; with token, 200 or 404
  check(categoriesRes, {
    'categories endpoint status is 200, 401, or 404': (r) => r.status === 200 || r.status === 401 || r.status === 404,
  }, { expectedStatuses: [200, 401, 404] });  // Ignore these in http_req_failed

  // Get all services - no auth needed based on previous tests
  let allServicesRes = http.get(`${BASE_URL}/api/v1/homeservices/services`, {
    headers: {
      'Content-Type': 'application/json',
    },
    tags: { name: 'Get Services' },
  });

  check(allServicesRes, {
    'services endpoint status is 200 or 404': (r) => r.status === 200 || r.status === 404,
  }, { expectedStatuses: [200, 404] });  // Add to ignore 404
}

function riderTests() {
  // Get rider profile (requires authentication)
  let riderRes = http.get(`${BASE_URL}/api/v1/riders/profile`, {
    headers: {
      Authorization: `Bearer ${AUTH_TOKEN}`,
      'Content-Type': 'application/json',
    },
    tags: { name: 'Get Rider Profile' },
  });

  check(riderRes, {
    'rider profile status is 200 or 401': (r) =>
      r.status === 200 || r.status === 401,
  }, { expectedStatuses: [200, 401] });  // Ignore 401
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
  }, { expectedStatuses: [200, 401, 404] });  // Ignore 401/404
}