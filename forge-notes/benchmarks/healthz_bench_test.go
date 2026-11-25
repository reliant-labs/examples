package benchmarks

// Package benchmarks hosts the project's microbenchmarks. Each file
// follows the standard Go convention (`func BenchmarkXxx(b *testing.B)`)
// so they run under `go test -bench` / `task bench` without special flags.
//
// Benchmarks live in their own package so they can evolve independently
// of unit tests in pkg/middleware — no risk of accidentally sharing
// state or setup across the two.

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

// healthzHandler mirrors cmd/server.go's liveness endpoint closely
// enough to give a realistic baseline for the request/response path:
// a tiny handler with no dependencies. If the real handler grows
// (metrics, checks), update this to match so the benchmark stays
// representative.
func healthzHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

// BenchmarkHealthz measures the cost of serving a single /healthz
// request via net/http/httptest. Reports ns/op and allocs/op so a CI
// regression gate (e.g. benchstat) can detect unexpected growth.
func BenchmarkHealthz(b *testing.B) {
	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/healthz", nil)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rr := httptest.NewRecorder()
		healthzHandler(rr, req)
		if rr.Code != http.StatusOK {
			b.Fatalf("unexpected status: %d", rr.Code)
		}
	}
}
