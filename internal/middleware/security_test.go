package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSecurityHeaders(t *testing.T) {
	// Define expected headers and their values
	expectedHeaders := map[string]string{
		"Strict-Transport-Security": "max-age=31536000; includeSubDomains; preload",
		"X-Frame-Options":           "SAMEORIGIN",
		"X-Content-Type-Options":    "nosniff",
		"Referrer-Policy":           "no-referrer-when-downgrade",
	}

	// Create a simple handler to wrap with our middleware
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Wrap our handler with the security middleware
	secureHandler := SecurityHeaders(nextHandler)

	// Create a test server
	ts := httptest.NewServer(secureHandler)
	defer ts.Close()

	// Make a request to the server
	resp, err := http.Get(ts.URL)
	if err != nil {
		t.Fatalf("Error making request to test server: %v", err)
	}
	defer resp.Body.Close()

	// Verify all expected headers are present with correct values
	for header, expectedValue := range expectedHeaders {
		value := resp.Header.Get(header)
		if value != expectedValue {
			t.Errorf("Expected header %s to be %q, got %q", header, expectedValue, value)
		}
	}

	// Verify response code is still OK (200)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}
}

func TestSecurityHeadersWithCustomHeaders(t *testing.T) {
	// Create a handler that sets its own headers
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Custom-Header", "custom-value")
		w.WriteHeader(http.StatusOK)
	})

	// Wrap with security middleware
	secureHandler := SecurityHeaders(nextHandler)

	// Create a test request and response recorder
	req := httptest.NewRequest("GET", "/test", nil)
	recorder := httptest.NewRecorder()

	// Execute the handler
	secureHandler.ServeHTTP(recorder, req)

	// Verify security headers are present
	expectedSecurityHeaders := []string{
		"Strict-Transport-Security",
		"X-Frame-Options",
		"X-Content-Type-Options",
		"Referrer-Policy",
	}

	for _, header := range expectedSecurityHeaders {
		if recorder.Header().Get(header) == "" {
			t.Errorf("Security header %s is missing", header)
		}
	}

	// Verify custom headers are still present
	if recorder.Header().Get("Content-Type") != "application/json" {
		t.Errorf("Expected Content-Type header to be 'application/json', got %q",
			recorder.Header().Get("Content-Type"))
	}

	if recorder.Header().Get("Custom-Header") != "custom-value" {
		t.Errorf("Expected Custom-Header to be 'custom-value', got %q",
			recorder.Header().Get("Custom-Header"))
	}
}
