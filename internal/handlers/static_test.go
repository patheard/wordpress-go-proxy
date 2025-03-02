package handlers

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewStaticHandler(t *testing.T) {
	staticDir := "/test/static"
	handler := NewStaticHandler(staticDir)

	if handler == nil {
		t.Fatal("Expected handler to be non-nil")
	}

	if handler.staticDir != staticDir {
		t.Errorf("Expected staticDir to be %q, got %q", staticDir, handler.staticDir)
	}

	// Can't directly compare the file server, but it should be initialized
	if handler.fileServer == nil {
		t.Error("Expected fileServer to be initialized")
	}
}

func TestStaticHandlerServeHTTP(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "static_test")
	if err != nil {
		t.Fatalf("Could not create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test files with different extensions
	testFiles := map[string]string{
		"test.css":    "body { color: red; }",
		"test.js":     "function test() { return 'test'; }",
		"test.html":   "<html><body>Test</body></html>",
		"test.txt":    "This is a text file",
		"test.json":   `{"key": "value"}`,
		"noextension": "File with no extension",
	}

	for filename, content := range testFiles {
		err := os.WriteFile(filepath.Join(tmpDir, filename), []byte(content), 0644)
		if err != nil {
			t.Fatalf("Could not create test file %s: %v", filename, err)
		}
	}

	// Create the static handler
	handler := NewStaticHandler(tmpDir)

	// Test cases
	testCases := []struct {
		name           string
		path           string
		expectedStatus int
		expectedType   string
		checkBody      bool
	}{
		{
			name:           "CSS file",
			path:           "/test.css",
			expectedStatus: http.StatusOK,
			expectedType:   "text/css; charset=utf-8",
			checkBody:      true,
		},
		{
			name:           "JavaScript file",
			path:           "/test.js",
			expectedStatus: http.StatusOK,
			expectedType:   "application/javascript",
			checkBody:      true,
		},
		{
			name:           "HTML file",
			path:           "/test.html",
			expectedStatus: http.StatusOK,
			expectedType:   "text/html; charset=utf-8",
			checkBody:      true,
		},
		{
			name:           "Text file",
			path:           "/test.txt",
			expectedStatus: http.StatusOK,
			expectedType:   "text/plain; charset=utf-8",
			checkBody:      true,
		},
		{
			name:           "JSON file",
			path:           "/test.json",
			expectedStatus: http.StatusOK,
			expectedType:   "application/json",
			checkBody:      true,
		},
		{
			name:           "File with no extension",
			path:           "/noextension",
			expectedStatus: http.StatusOK,
			expectedType:   "", // No content type should be set
			checkBody:      true,
		},
		{
			name:           "Directory listing",
			path:           "/",
			expectedStatus: http.StatusOK,
			expectedType:   "text/html; charset=utf-8",
			checkBody:      false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tc.path, nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			// Check status code
			if resp.StatusCode != tc.expectedStatus {
				t.Errorf("Expected status code %d, got %d", tc.expectedStatus, resp.StatusCode)
			}

			// Check Content-Type header
			contentType := resp.Header.Get("Content-Type")
			if tc.expectedType != "" && contentType != tc.expectedType {
				t.Errorf("Expected Content-Type %q, got %q", tc.expectedType, contentType)
			}

			// Check Cache-Control header
			cacheControl := resp.Header.Get("Cache-Control")
			expectedCache := "public, max-age=604800"
			if cacheControl != expectedCache {
				t.Errorf("Expected Cache-Control %q, got %q", expectedCache, cacheControl)
			}

			// For files, verify content if requested
			if tc.checkBody && tc.expectedStatus == http.StatusOK {
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					t.Errorf("Error reading response body: %v", err)
				}

				filename := strings.TrimPrefix(tc.path, "/")
				expectedContent := testFiles[filename]
				if string(body) != expectedContent {
					t.Errorf("Expected body %q, got %q", expectedContent, string(body))
				}
			}
		})
	}
}

func TestStaticHandlerWithDifferentMethods(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "static_test")
	if err != nil {
		t.Fatalf("Could not create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a test file
	testFile := "test.txt"
	err = os.WriteFile(filepath.Join(tmpDir, testFile), []byte("Test content"), 0644)
	if err != nil {
		t.Fatalf("Could not create test file: %v", err)
	}

	// Create the static handler
	handler := NewStaticHandler(tmpDir)

	// Test different HTTP methods
	methods := []string{"GET", "HEAD", "POST", "PUT", "DELETE", "PATCH"}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/"+testFile, nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			// GET and HEAD are allowed for static files
			if method == "GET" || method == "HEAD" {
				if resp.StatusCode != http.StatusOK {
					t.Errorf("Expected status 200 OK for %s request, got %d", method, resp.StatusCode)
				}
			} else {
				// FileServer typically returns 200 for non-standard methods too, but the body is empty
				// Just make sure it doesn't crash
				if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusMethodNotAllowed {
					t.Errorf("Unexpected status %d for %s request", resp.StatusCode, method)
				}
			}

			// Check Cache-Control header is always set
			cacheControl := resp.Header.Get("Cache-Control")
			expectedCache := "public, max-age=604800"
			if cacheControl != expectedCache {
				t.Errorf("Expected Cache-Control %q, got %q", expectedCache, cacheControl)
			}
		})
	}
}

func TestStaticHandlerWithNestedPaths(t *testing.T) {
	// Create a temporary directory structure
	rootDir, err := os.MkdirTemp("", "static_test_nested")
	if err != nil {
		t.Fatalf("Could not create temp directory: %v", err)
	}
	defer os.RemoveAll(rootDir)

	// Create nested directories
	nestedDir := filepath.Join(rootDir, "css")
	if err := os.Mkdir(nestedDir, 0755); err != nil {
		t.Fatalf("Could not create nested directory: %v", err)
	}

	// Create files in both root and nested directories
	rootFile := filepath.Join(rootDir, "index.html")
	nestedFile := filepath.Join(nestedDir, "style.css")

	if err := os.WriteFile(rootFile, []byte("<html>Root file</html>"), 0644); err != nil {
		t.Fatalf("Could not create root file: %v", err)
	}

	if err := os.WriteFile(nestedFile, []byte("body { color: blue; }"), 0644); err != nil {
		t.Fatalf("Could not create nested file: %v", err)
	}

	// Create the static handler
	handler := NewStaticHandler(rootDir)

	// Test accessing the nested file
	req := httptest.NewRequest("GET", "/css/style.css", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200 for nested file, got %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("Error reading response body: %v", err)
	}

	expectedContent := "body { color: blue; }"
	if string(body) != expectedContent {
		t.Errorf("Expected body %q, got %q", expectedContent, string(body))
	}

	contentType := resp.Header.Get("Content-Type")
	expectedType := "text/css; charset=utf-8"
	if contentType != expectedType {
		t.Errorf("Expected Content-Type %q, got %q", expectedType, contentType)
	}

	// Test directory traversal attempt
	req = httptest.NewRequest("GET", "/../outside.txt", nil)
	w = httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	resp = w.Result()
	defer resp.Body.Close()

	// Should not allow accessing files outside the static directory
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status 404 for directory traversal attempt, got %d", resp.StatusCode)
	}
}
