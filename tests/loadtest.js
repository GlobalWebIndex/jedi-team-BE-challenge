import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  vus: 10, // number of virtual users
  duration: '30s', // how long test runs
};

export default function () {
  const url = 'http://gateway:8080/chat';
  const payload = JSON.stringify({
    user_id: Math.floor(Math.random() * 1000), // random user
    message: "Tell me a fun fact about space."
  });

  const params = {
    headers: {
      'Content-Type': 'application/json',
    },
    timeout: '30s',
  };

  let res = http.post(url, payload, params);

  check(res, {
    'status is 200': (r) => r.status === 200,
    'response contains message': (r) => r.json('response') !== undefined,
    'response time under 15s': (r) => r.timings.duration < 30000
  });

  sleep(1);
}
