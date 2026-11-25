# benchmarks

Two flavours of perf harness:

- `k6/` — [k6](https://k6.io) scripts for HTTP-level load tests.
  Good for smoke-testing a running service end-to-end.
- `*_bench_test.go` — Go `testing.B` microbenchmarks. Good for
  measuring specific handlers or hot paths in isolation.

## Running k6

Install k6: <https://k6.io/docs/get-started/installation/>.

Start the service locally (`task dev` or `docker compose up`), then:

```bash
k6 run benchmarks/k6/smoke.js
# Or point at a staging host:
k6 run --env BASE_URL=https://staging.example.com benchmarks/k6/smoke.js
```

### What to watch

- `http_req_duration` — p(50), p(95), p(99). Thresholds in the script
  will fail the run if p(95) regresses past the baseline.
- `http_req_failed` — should stay below 1%. Any non-2xx response that
  isn't an intentional auth/404 test counts.
- `vus` / `iterations` — sanity check that ramping actually happened.

## Running Go benchmarks

```bash
task bench
# or, for a specific package:
go test -bench=. -benchmem ./benchmarks/...
```

Follow up with `benchstat` when comparing two revisions:

```bash
go install golang.org/x/perf/cmd/benchstat@latest
go test -bench=. -count=10 ./benchmarks/... > old.txt
# ...make change, then...
go test -bench=. -count=10 ./benchmarks/... > new.txt
benchstat old.txt new.txt
```
