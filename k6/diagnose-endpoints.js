import http from 'k6/http';
import { sleep } from 'k6';

// Configuration
const BASE_URL = __ENV.BASE_URL || 'https://api.pittapizzahusrev.be/go';
const AUTH_TOKEN = __ENV.AUTH_TOKEN || '';

// Run with 1 VU to see status codes clearly
export const options = {
  vus: 1,
  duration: '10s',
  thresholds: {},
};

export default function () {
  console.log('\n=== DIAGNOSTIC: Testing all endpoints ===\n');

  // Test 1: Health Check
  let res = http.get(`${BASE_URL}/health`);
  console.log(`✓ Health: ${res.status} ${res.url}`);
  console.log(`  Response: ${res.body.substring(0, 100)}\n`);

  // Test 2: Categories
  res = http.get(`${BASE_URL}/api/v1/homeservices/categories`, {
    headers: {
      'Authorization': `Bearer ${AUTH_TOKEN}`,
      'Content-Type': 'application/json',
    },
  });
  console.log(`✓ Categories: ${res.status} ${res.url}`);
  console.log(`  Response: ${res.body.substring(0, 100)}\n`);

  // Test 3: Service Providers
  res = http.get(`${BASE_URL}/api/v1/serviceproviders`, {
    headers: {
      'Authorization': `Bearer ${AUTH_TOKEN}`,
      'Content-Type': 'application/json',
    },
  });
  console.log(`✓ Providers: ${res.status} ${res.url}`);
  console.log(`  Response: ${res.body.substring(0, 100)}\n`);

  // Test 4: Rider Profile
  res = http.get(`${BASE_URL}/api/v1/riders/profile`, {
    headers: {
      'Authorization': `Bearer ${AUTH_TOKEN}`,
      'Content-Type': 'application/json',
    },
  });
  console.log(`✓ Rider Profile: ${res.status} ${res.url}`);
  console.log(`  Response: ${res.body.substring(0, 100)}\n`);

  // Test 5: Driver Profile
  res = http.get(`${BASE_URL}/api/v1/drivers/profile`, {
    headers: {
      'Authorization': `Bearer ${AUTH_TOKEN}`,
      'Content-Type': 'application/json',
    },
  });
  console.log(`✓ Driver Profile: ${res.status} ${res.url}`);
  console.log(`  Response: ${res.body.substring(0, 100)}\n`);

  // Test 6: Check if /api/v1/homeservices/categories exists (maybe it's /categories?)
  res = http.get(`${BASE_URL}/api/v1/homeservices/list`, {
    headers: {
      'Authorization': `Bearer ${AUTH_TOKEN}`,
      'Content-Type': 'application/json',
    },
  });
  console.log(`✓ Homeservices List: ${res.status} ${res.url}`);
  console.log(`  Response: ${res.body.substring(0, 100)}\n`);

  console.log('=== END DIAGNOSTIC ===\n');

  sleep(1);
}
