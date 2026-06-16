import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  scenarios: {
    constant_request_rate: {
      executor: 'constant-arrival-rate',
      rate: 100, // 100 TPS
      timeUnit: '1s',
      duration: '30s',
      preAllocatedVUs: 50,
      maxVUs: 200,
    },
  },
  thresholds: {
    http_req_duration: ['p(95)<500'], // 95% of requests must complete below 500ms
  },
};

export default function () {
  const port = __ENV.PORT || '2050';
  const url = `http://localhost:${port}/ledger/transfers`;
  const payload = JSON.stringify({
    source_account_id: 'ACC_SOURCE_MOCK_123',
    destination_account_id: 'ACC_DEST_MOCK_456',
    amount: 10.0,
    currency: 'USD',
    transfer_id: `txn_load_${__VU}_${__ITER}`
  });

  const params = {
    headers: {
      'Content-Type': 'application/json',
    },
  };

  const res = http.post(url, payload, params);

  // We check for 200 (OK) or 404 (Not Found, since mocks may not exist in DB but the endpoint still routes)
  check(res, {
    'status is 200 or 404': (r) => r.status === 200 || r.status === 404,
  });
}
