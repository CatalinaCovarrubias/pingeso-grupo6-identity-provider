package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestHealth checks that the health endpoint returns 200 and the expected body.
func TestHealth(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	rec := httptest.NewRecorder()

	Health(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `"estado":"ok"`) {
		t.Fatalf("unexpected body: %s", rec.Body.String())
	}
}