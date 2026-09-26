package routes

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthRoute(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/health", nil)
	recorder := httptest.NewRecorder()
	Router(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}
	if recorder.Body.String() != "{\"status\":\"ok\"}\n" {
		t.Fatalf("unexpected health response: %q", recorder.Body.String())
	}
}
