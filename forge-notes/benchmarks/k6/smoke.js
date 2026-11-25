// smoke.js — k6 smoke test.
// Ramps a handful of VUs against the running service and fails the run
// if p95 latency or error rate blow past the thresholds below. Keep the
// load modest; this is a smoke test, not a capacity test.
//
// Run with:
//     k6 run --env BASE_URL=http://localhost:8080 benchmarks/k6/smoke.js

import http from "k6/http";
import { check, sleep } from "k6";

export const options = {
  stages: [
    { duration: "10s", target: 5 },
    { duration: "20s", target: 10 },
    { duration: "10s", target: 0 },
  ],
  thresholds: {
    http_req_failed: ["rate<0.01"],
    http_req_duration: ["p(95)<300"],
  },
};

const BASE_URL = __ENV.BASE_URL || "http://localhost:8080";

export default function () {
  const healthz = http.get(`${BASE_URL}/healthz`);
  check(healthz, {
    "healthz 200": (r) => r.status === 200,
  });

  sleep(1);
}
