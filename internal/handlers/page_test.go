package handlers

import (
	"bytes"
	"encoding/json"
	"html/template"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"wordpress-go-proxy/internal/api"
	"wordpress-go-proxy/pkg/models"
)

// setupTestTemplates creates a mock template for testing
func setupTestTemplates() *template.Template {
	tmpl := template.New("layout.html")
	tmpl, err := tmpl.Parse(`<!DOCTYPE html>
<html lang="{{.Lang}}">
<head><title>{{.Title}}</title></head>
<body>{{.Content}}</body>
</html>`)
	if err != nil {
		panic(err)
	}
	return tmpl
}

// setupTestServer creates a test HTTP server that mimics WordPress API responses
func setupTestServer(t *testing.T, responses map[string]interface{}) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set default content type
		w.Header().Set("Content-Type", "application/json")

		// Check for custom error status codes
		if statusCode, ok := responses["statusCode"]; ok {
			if code, ok := statusCode.(int); ok && code != http.StatusOK {
				http.Error(w, "API Error", code)
				return
			}
		}

		// Handle WordPress API paths
		switch {
		case strings.Contains(r.URL.Path, "/wp-json/wp/v2/pages"):
			// Page endpoint
			slug := r.URL.Query().Get("slug")
			key := "pages/" + slug

			if response, ok := responses[key]; ok {
				json.NewEncoder(w).Encode(response)
				return
			}

			// Default page response
			if val, ok := responses["defaultPage"]; ok {
				json.NewEncoder(w).Encode(val)
				return
			}

			// Empty response if not found
			json.NewEncoder(w).Encode([]models.WordPressPage{})

		case strings.Contains(r.URL.Path, "/wp-json/wp/v2/menu-items"):
			// Menu endpoint
			lang := "en"
			if r.URL.Query().Get("menus") == "menu-fr" {
				lang = "fr"
			}

			key := "menu/" + lang
			if response, ok := responses[key]; ok {
				json.NewEncoder(w).Encode(response)
				return
			}

			// Default empty menu
			json.NewEncoder(w).Encode([]models.WordPressMenuItem{})
		}
	}))
}

// TestNewPageHandler tests the creation of a new page handler
func TestNewPageHandler(t *testing.T) {
	// Save the original template parsing function and restore it after the test
	originalParseFiles := parseTemplateFiles
	parseTemplateFiles = func(filenames ...string) (*template.Template, error) {
		return setupTestTemplates(), nil
	}
	defer func() { parseTemplateFiles = originalParseFiles }()

	// Setup test server and client
	server := setupTestServer(t, map[string]interface{}{
		"menu/en": []models.WordPressMenuItem{},
		"menu/fr": []models.WordPressMenuItem{},
	})
	defer server.Close()

	client := api.NewWordPressClient(
		server.URL,
		"testuser",
		"testpass",
		"menu-en",
		"menu-fr",
	)

	// Create site names
	siteNames := map[string]string{
		"en": "English Site",
		"fr": "French Site",
	}

	// Create the handler
	handler := NewPageHandler(siteNames, client)

	// Verify handler was created correctly
	if handler == nil {
		t.Fatal("Expected handler to be created, got nil")
	}

	if handler.SiteNames["en"] != "English Site" {
		t.Errorf("Expected English site name, got %s", handler.SiteNames["en"])
	}

	if handler.WordPressClient == nil {
		t.Error("Expected client to be assigned")
	}

	if handler.Templates == nil {
		t.Error("Expected templates to be initialized")
	}
}

// TestServeHTTP tests the HTTP request handling logic
func TestServeHTTP(t *testing.T) {
	// Setup the test responses
	testResponses := map[string]interface{}{
		"defaultPage": []models.WordPressPage{{
			ID:   1,
			Slug: "test-page",
			Lang: "en",
			Title: struct {
				Rendered string `json:"rendered"`
			}{Rendered: "Test Page"},
			Content: struct {
				Rendered string `json:"rendered"`
				Raw      string `json:"raw,omitempty"`
			}{Rendered: "<p>Test content</p>"},
		}},
	}

	// Setup test server
	server := setupTestServer(t, testResponses)
	defer server.Close()

	// Create real client pointing to test server
	client := api.NewWordPressClient(
		server.URL,
		"testuser",
		"testpass",
		"menu-en",
		"menu-fr",
	)

	// Create handler with the real client and mocked templates
	siteNames := map[string]string{
		"en": "English Site",
		"fr": "French Site",
	}

	handler := &PageHandler{
		SiteNames:       siteNames,
		WordPressClient: client,
		Templates:       setupTestTemplates(),
	}

	tests := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
		expectError    bool
	}{
		{
			name:           "Valid GET request",
			method:         "GET",
			path:           "/about-us",
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "Valid HEAD request",
			method:         "HEAD",
			path:           "/about-us",
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "Valid OPTIONS request",
			method:         "OPTIONS",
			path:           "/about-us",
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "Invalid POST method",
			method:         "POST",
			path:           "/about-us",
			expectedStatus: http.StatusMethodNotAllowed,
			expectError:    true,
		},
		{
			name:           "Path with file extension",
			method:         "GET",
			path:           "/about-us.html",
			expectedStatus: http.StatusNotFound,
			expectError:    true,
		},
		{
			name:           "Path with invalid characters",
			method:         "GET",
			path:           "/about<us>",
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name:           "Path too long",
			method:         "GET",
			path:           "/" + strings.Repeat("a", 255),
			expectedStatus: http.StatusRequestURITooLong,
			expectError:    true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			// Check status code
			if resp.StatusCode != tc.expectedStatus {
				t.Errorf("Expected status %d, got %d", tc.expectedStatus, resp.StatusCode)
			}

			// For error cases, body should contain error message
			if tc.expectError {
				if resp.StatusCode == http.StatusOK {
					t.Errorf("Expected an error status code, got %d", resp.StatusCode)
				}
			} else {
				// For success cases
				if resp.StatusCode != http.StatusOK {
					t.Errorf("Expected HTTP 200 OK, got %d", resp.StatusCode)
				}

				// For GET requests, verify body contains expected content (not for HEAD)
				if tc.method == "GET" {
					body, _ := io.ReadAll(resp.Body)
					if !bytes.Contains(body, []byte("Test Page")) {
						t.Errorf("Expected body to contain page title, got: %s", string(body))
					}
				}
			}
		})
	}
}

// TestHandlePage tests the page handling logic
func TestHandlePage(t *testing.T) {
	tests := []struct {
		name           string
		path           string
		testResponses  map[string]interface{}
		expectedStatus int
		expectedTitle  string
	}{
		{
			name: "Successful page fetch - English",
			path: "/about-us",
			testResponses: map[string]interface{}{
				"pages/about-us": []models.WordPressPage{{
					ID:   1,
					Slug: "about-us",
					Lang: "en",
					Title: struct {
						Rendered string `json:"rendered"`
					}{Rendered: "About Us"},
					Content: struct {
						Rendered string `json:"rendered"`
						Raw      string `json:"raw,omitempty"`
					}{Rendered: "<p>About us content</p>"},
				}},
			},
			expectedStatus: http.StatusOK,
			expectedTitle:  "About Us",
		},
		{
			name: "Successful page fetch - French",
			path: "/fr/a-propos",
			testResponses: map[string]interface{}{
				"pages/a-propos": []models.WordPressPage{{
					ID:   2,
					Slug: "a-propos",
					Lang: "fr",
					Title: struct {
						Rendered string `json:"rendered"`
					}{Rendered: "À propos"},
					Content: struct {
						Rendered string `json:"rendered"`
						Raw      string `json:"raw,omitempty"`
					}{Rendered: "<p>Contenu à propos</p>"},
				}},
			},
			expectedStatus: http.StatusOK,
			expectedTitle:  "À propos",
		},
		{
			name: "Page not found",
			path: "/not-found",
			testResponses: map[string]interface{}{
				"pages/not-found": []models.WordPressPage{},
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Setup test server with specific responses
			server := setupTestServer(t, tc.testResponses)
			defer server.Close()

			// Create real client pointing to test server
			client := api.NewWordPressClient(
				server.URL,
				"testuser",
				"testpass",
				"menu-en",
				"menu-fr",
			)

			// Create handler
			handler := &PageHandler{
				SiteNames:       map[string]string{"en": "English Site", "fr": "French Site"},
				WordPressClient: client,
				Templates:       setupTestTemplates(),
			}

			// Create request and response recorder
			req := httptest.NewRequest("GET", tc.path, nil)
			w := httptest.NewRecorder()

			// Call the handler method directly
			handler.handlePage(w, req, tc.path)

			resp := w.Result()
			defer resp.Body.Close()

			// Verify status code
			if resp.StatusCode != tc.expectedStatus {
				t.Errorf("Expected status %d, got %d", tc.expectedStatus, resp.StatusCode)
			}

			// For successful cases, verify content
			if tc.expectedStatus == http.StatusOK {
				body, _ := io.ReadAll(resp.Body)
				if !bytes.Contains(body, []byte(tc.expectedTitle)) {
					t.Errorf("Expected body to contain page title %q, got: %s", tc.expectedTitle, string(body))
				}
			}
		})
	}
}

// TestTemplateRenderingError tests handling of template rendering errors
func TestTemplateRenderingError(t *testing.T) {
	// Create a template that will generate an error
	errorTemplate := template.New("layout.html")
	errorTemplate, _ = errorTemplate.Parse(`{{ .NonExistentField.WillCauseError }}`)

	// Setup test server with a valid page response
	testResponses := map[string]interface{}{
		"pages/test-page": []models.WordPressPage{{
			ID:   1,
			Slug: "test-page",
			Lang: "en",
			Title: struct {
				Rendered string `json:"rendered"`
			}{Rendered: "Test Page"},
			Content: struct {
				Rendered string `json:"rendered"`
				Raw      string `json:"raw,omitempty"`
			}{Rendered: "<p>Test content</p>"},
		}},
	}

	server := setupTestServer(t, testResponses)
	defer server.Close()

	// Create real client pointing to test server
	client := api.NewWordPressClient(
		server.URL,
		"testuser",
		"testpass",
		"menu-en",
		"menu-fr",
	)

	// Create handler with the error-generating template
	handler := &PageHandler{
		SiteNames:       map[string]string{"en": "English Site"},
		WordPressClient: client,
		Templates:       errorTemplate,
	}

	// Create request and response recorder
	req := httptest.NewRequest("GET", "/test-page", nil)
	w := httptest.NewRecorder()

	// Call the handler method
	handler.handlePage(w, req, "/test-page")

	resp := w.Result()
	defer resp.Body.Close()

	// Verify status code indicates error
	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, resp.StatusCode)
	}

	// Verify error message
	body, _ := io.ReadAll(resp.Body)
	expectedError := "Error rendering template"
	if !bytes.Contains(body, []byte(expectedError)) {
		t.Errorf("Expected error message containing %q, got: %s", expectedError, string(body))
	}
}
