import http from 'k6/http';
import { check, sleep, group } from 'k6';
import { Rate } from 'k6/metrics';

// Configuration
const BASE_URL = __ENV.BASE_URL || 'https://api.pittapizzahusrev.be/go';
let AUTH_TOKEN = __ENV.AUTH_TOKEN || ''; // Override via env if no setup login

// Custom metric for true errors (unexpected statuses)
export const errorRate = new Rate('errors');

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
    errors: ['rate<0.1'],                   // Custom error rate below 10%
  },
};

// Setup: Login to get auth token (run once)
export function setup() {
  if (AUTH_TOKEN) return { token: AUTH_TOKEN };  // Use env if provided

  const loginUrl = `${BASE_URL}/api/v1/auth/login`;  // Adjust if path is different (e.g., /auth/login)
  const payload = JSON.stringify({
    email: 'test@example.com',  // Replace with valid test credentials
    password: 'testpassword',
  });
  const params = { headers: { 'Content-Type': 'application/json' } };

  const res = http.post(loginUrl, payload, params);
  check(res, { 'login status is 200': (r) => r.status === 200 });

  const authData = res.json();
  const token = authData.token || authData.access_token;  // Adjust based on response structure (check docs for field)

  if (!token) {
    throw new Error('Login failed: No token received');
  }

  return { token };
}

export default function (data) {
  AUTH_TOKEN = data.token;  // Use token from setup

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
  const healthCheck = check(healthRes, {
    'health check status is 200': (r) => r.status === 200,
  });
  errorRate.add(!healthCheck);
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
  const categoriesCheck = check(categoriesRes, {
    'categories endpoint status is 200 or 404': (r) => r.status === 200 || r.status === 404,  // Remove 401 now that we have token
  });
  errorRate.add(!categoriesCheck);

  // Get all services - assuming public, but add token if needed
  let allServicesRes = http.get(`${BASE_URL}/api/v1/homeservices/services`, {
    headers: {
      'Content-Type': 'application/json',
    },
    tags: { name: 'Get Services' },
  });
  const servicesCheck = check(allServicesRes, {
    'services endpoint status is 200 or 404': (r) => r.status === 200 || r.status === 404,
  });
  errorRate.add(!servicesCheck);
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
  const riderCheck = check(riderRes, {
    'rider profile status is 200 or 404': (r) => r.status === 200 || r.status === 404,  // Assume 404 if not found, remove 401
  });
  errorRate.add(!riderCheck);
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
  const driverCheck = check(driverRes, {
    'driver profile status is 200 or 404': (r) => r.status === 200 || r.status === 404,
  });
  errorRate.add(!driverCheck);
}