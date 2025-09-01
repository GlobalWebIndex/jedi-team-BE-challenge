import http from 'k6/http';
import { check, sleep } from 'k6';

const baseUrl = 'http://gateway:8080';

export const options = {
    scenarios: {
      post_chat_test: {
        executor: 'constant-vus',
        vus: 5,
        duration: '30s',
        exec: 'testPostChat',
      },
      get_chat_stress_test: {
        executor: 'ramping-vus',
        startVUs: 5,
        stages: [
          { duration: '10s', target: 50 },
          { duration: '30s', target: 50 },
          { duration: '10s', target: 100 },
          { duration: '20s', target: 100 },
          { duration: '10s', target: 0 },
        ],
        exec: 'testGetChat',
      },
      get_chat_users_stress_test: {
        executor: 'ramping-vus',
        startVUs: 5,
        stages: [
          { duration: '10s', target: 60 },
          { duration: '40s', target: 60 },
          { duration: '20s', target: 120 },
          { duration: '20s', target: 120 },
          { duration: '10s', target: 0 },
        ],
        exec: 'testGetChatUsers',
      },
    },
    thresholds: {
      http_req_duration: ['p(95)<5000'],
      http_req_failed: ['rate<0.05'],

      'http_req_duration{scenario:post_chat_test}': ['p(95)<30000'],
      'http_req_duration{scenario:get_chat_stress_test}': ['p(95)<2000'],
      'http_req_duration{scenario:get_chat_users_stress_test}': ['p(95)<2000'],
      
      'http_req_failed{scenario:post_chat_test}': ['rate<0.01'],
      'http_req_failed{scenario:get_chat_stress_test}': ['rate<0.1'],
      'http_req_failed{scenario:get_chat_users_stress_test}': ['rate<0.1'],
    },
  };

export default function () {  
  testPostChat();
  
  testGetChat();
  
  testGetChatUsers();
  
  sleep(1);
}

export function testPostChat() {
  const url = `${baseUrl}/chat`;
  const payload = JSON.stringify({
    user_id: Math.floor(Math.random() * 1000),
    message: "Tell me a fun fact about space."
  });

  const params = {
    headers: {
      'Content-Type': 'application/json',
    },
    timeout: '32s',
  };

  let res = http.post(url, payload, params);

  check(res, {
    'POST /chat - status is 200': (r) => r.status === 200,
    'POST /chat - response contains message': (r) => r.json('response') !== undefined,
    'POST /chat - response time under 30s': (r) => r.timings.duration < 32000
  });
}

export function testGetChat() {
  const chatId = Math.floor(Math.random() * 100) + 1;
  const url = `${baseUrl}/chat/${chatId}/history`;
  
  const params = {
    timeout: '30s',
  };

  let res = http.get(url, params);

  check(res, {
    'GET /chat/{id} - status is 200 or 404': (r) => r.status === 200 || r.status === 404,
    'GET /chat/{id} - response time under 15s': (r) => r.timings.duration < 15000,
    'GET /chat/{id} - valid JSON response': (r) => {
      try {
        JSON.parse(r.body);
        return true;
      } catch (e) {
        return false;
      }
    }
  });
}

export function testGetChatUsers() {
  const userId = Math.floor(Math.random() * 100) + 1;
  const url = `${baseUrl}/chat/users/${userId}`;
  
  const params = {
    timeout: '30s',
  };

  let res = http.get(url, params);

  check(res, {
    'GET /chat/users/{id} - status is 200 or 404': (r) => r.status === 200 || r.status === 404,
    'GET /chat/users/{id} - response time under 15s': (r) => r.timings.duration < 15000,
    'GET /chat/users/{id} - valid JSON response': (r) => {
      try {
        JSON.parse(r.body);
        return true;
      } catch (e) {
        return false;
      }
    }
  });
}