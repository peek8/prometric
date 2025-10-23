import http from 'k6/http';
import { sleep } from 'k6';
import { randomItem } from 'https://jslib.k6.io/k6-utils/1.4.0/index.js';

const BASE_URL = 'http://localhost:8080';

const firstNames = [
     "Alice", "Bob", "Charlie", "David", "Eve", "Frank", "Grace", "Heidi",
    "Ivy", "Jack", "Karen", "Leo", "Mona", "Nina", "Oscar", "Paul", "Quinn",
    "Rita", "Sam", "Tina", "Uma", "Vik", "Walt", "Xena", "Yuri", "Zane"
];
const lastNames  = [
    "Anderson", "Brown", "Clark", "Davis", "Evans", "Fisher", "Garcia", "Hill",
    "Irwin", "Johnson", "Keller", "Lopez", "Miller", "Nelson", "Owens", "Perez",
    "Quinn", "Roberts", "Smith", "Taylor", "Upton", "Vargas", "White", "Young", "Zimmer"
];

function randomPerson() {
  return {
    first_name: randomItem(firstNames),
    last_name: randomItem(lastNames),
    email: `${Math.random().toString(36).substring(7)}@example.com`,
  };
}

// --- Scenarios ---
export const options = {
  scenarios: {
    create_users: {
      executor: 'constant-arrival-rate',
      rate: 50, // requests per second
      timeUnit: '1s',
      duration: '30s',
      preAllocatedVUs: 10,
      exec: 'createPerson',
    },
    create_users_more: {
      executor: 'shared-iterations',
      vus: 50,
      iterations: 50000,
      startTime: '0s',
      exec: 'createPerson',
    },
    update_users: {
      executor: 'constant-arrival-rate',
      rate: 5, // 5 req/s
      timeUnit: '1s',
      duration: '30s',
      preAllocatedVUs: 5,
      exec: 'updatePerson',
      startTime: '2s', // slight delay
    },
    delete_users: {
      executor: 'shared-iterations',
      vus: 10,
      iterations: 100,
      exec: 'deletePerson',
      startTime: '10s',
    },
    get_users: {
      executor: 'shared-iterations',
      vus: 20,
      iterations: 5000,
      exec: 'getPerson',
      startTime: '1s',
    },
  },
};

// --- Scenario Functions ---

export function createPerson() {
  const person = randomPerson();
  //console.log("Json person:: ", JSON.stringify(person))
  const res = http.post(`${BASE_URL}/person`, JSON.stringify(person), {
    headers: { 'Content-Type': 'application/json' },
  });
  sleep(0.5);
}

export function updatePerson() {
  const id = Math.floor(Math.random() * 1000);
  const person = randomPerson();
  http.put(`${BASE_URL}/person/${id}`, JSON.stringify(person), {
    headers: { 'Content-Type': 'application/json' },
  });
  sleep(0.5);
}

export function deletePerson() {
  const id = Math.floor(Math.random() * 1000);
  http.del(`${BASE_URL}/person/${id}`);
  sleep(0.5);
}

export function getPerson() {
  const id = Math.floor(Math.random() * 1000);
  http.get(`${BASE_URL}/person/${id}`);
  sleep(0.5);
}
