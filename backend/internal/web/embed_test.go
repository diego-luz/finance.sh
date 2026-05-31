package web

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestHandlerServesIndex verifies the embedded app shell is served at root.
func TestHandlerServesIndex(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GET / status = %d, want 200", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "finance.sh") {
		t.Errorf("GET / body does not look like the app shell: %q", rec.Body.String())
	}
	if cc := rec.Header().Get("Cache-Control"); cc != "no-cache" {
		t.Errorf("index Cache-Control = %q, want no-cache", cc)
	}
}

// TestHandlerSPAFallback verifies client-side deep links fall back to index.html
// instead of 404, so the SPA router can resolve them.
func TestHandlerSPAFallback(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/transactions/123/edit", nil)
	Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("deep-link status = %d, want 200 (SPA fallback)", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "<!doctype html>") &&
		!strings.Contains(rec.Body.String(), "<html") {
		t.Errorf("deep-link did not fall back to the HTML shell: %q", rec.Body.String())
	}
}
