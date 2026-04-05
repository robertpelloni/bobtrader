package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewHandlerHealthAndReady(t *testing.T) {
	h := NewHandler(Status{Name: "ultratrader-go", Ready: true, AccountCount: 1})

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("healthz expected 200, got %d", w.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/readyz", nil)
	w = httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("readyz expected 200, got %d", w.Code)
	}
}
