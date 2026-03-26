package aso

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCallbackHandlerCorrectState(t *testing.T) {
	state := "test-state-abc123"
	resultCh := make(chan callbackResult, 1)
	handler := callbackHandler(state, resultCh)

	req := httptest.NewRequest(http.MethodGet, "/?state=test-state-abc123&key=my-api-key", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}

	result := <-resultCh
	if result.err != nil {
		t.Fatalf("unexpected error: %v", result.err)
	}
	if result.key != "my-api-key" {
		t.Errorf("key = %q, want %q", result.key, "my-api-key")
	}
}

func TestCallbackHandlerWrongState(t *testing.T) {
	state := "correct-state"
	resultCh := make(chan callbackResult, 1)
	handler := callbackHandler(state, resultCh)

	req := httptest.NewRequest(http.MethodGet, "/?state=wrong-state&key=my-api-key", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}

	result := <-resultCh
	if result.err == nil {
		t.Fatal("expected error for wrong state, got nil")
	}
}

func TestCallbackHandlerMissingKey(t *testing.T) {
	state := "test-state"
	resultCh := make(chan callbackResult, 1)
	handler := callbackHandler(state, resultCh)

	req := httptest.NewRequest(http.MethodGet, "/?state=test-state", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}

	result := <-resultCh
	if result.err == nil {
		t.Fatal("expected error for missing key, got nil")
	}
}

func TestGenerateState(t *testing.T) {
	s1, err := generateState()
	if err != nil {
		t.Fatalf("generateState: %v", err)
	}
	if len(s1) != 32 { // 16 bytes = 32 hex chars
		t.Errorf("state length = %d, want 32", len(s1))
	}

	s2, err := generateState()
	if err != nil {
		t.Fatalf("generateState: %v", err)
	}
	if s1 == s2 {
		t.Error("two generated states should not be equal")
	}
}
