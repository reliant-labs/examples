//go:build e2e
// +build e2e

package notes_test

import (
	"context"
	"net/http"
	"testing"
)

// TestHealthCheck verifies the service binary starts and responds to health checks.
func TestHealthCheck(t *testing.T) {
	t.Parallel()

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, baseURL+"/healthz", nil)
	if err != nil {
		t.Fatalf("failed to create health check request: %v", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("health check request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 from /healthz, got %d", resp.StatusCode)
	}
}
