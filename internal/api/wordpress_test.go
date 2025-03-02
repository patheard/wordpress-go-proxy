package api

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"wordpress-go-proxy/pkg/models"
)

type Rendered struct {
	Rendered string `json:"rendered"`
}

func TestFetchPage(t *testing.T) {
	// Test cases for different page paths
	testCases := []struct {
		name           string
		path           string
		expectedSlug   string
		expectedLang   string
		mockedResponse []models.WordPressPage
		shouldError    bool
		errorMessage   string
	}{
		{
			name:         "Regular English page",
			path:         "/about-us",
			expectedSlug: "about-us",
			expectedLang: "en",
			mockedResponse: []models.WordPressPage{
				{
					ID: 123,
					Title: Rendered{
						Rendered: "About Us",
					},
					Slug: "about-us",
				},
			},
		},
		{
			name:         "Regular French page",
			path:         "/fr/a-propos",
			expectedSlug: "a-propos",
			expectedLang: "fr",
			mockedResponse: []models.WordPressPage{
				{
					ID: 124,
					Title: Rendered{
						Rendered: "À propos",
					},
					Slug: "a-propos",
				},
			},
		},
		{
			name:         "English home page",
			path:         "/",
			expectedSlug: "home",
			expectedLang: "en",
			mockedResponse: []models.WordPressPage{
				{
					ID: 125,
					Title: Rendered{
						Rendered: "Home",
					},
					Slug: "home",
				},
			},
		},
		{
			name:         "French home page",
			path:         "/fr",
			expectedSlug: "home-fr",
			expectedLang: "fr",
			mockedResponse: []models.WordPressPage{
				{
					ID: 126,
					Title: Rendered{
						Rendered: "About Us",
					},
					Slug: "home-fr",
				},
			},
		},
		{
			name:           "Page not found",
			path:           "/non-existent",
			expectedSlug:   "non-existent",
			expectedLang:   "en",
			mockedResponse: []models.WordPressPage{},
			shouldError:    true,
			errorMessage:   "page not found",
		},
		{
			name:         "API error response",
			path:         "/error-page",
			expectedSlug: "error-page",
			expectedLang: "en",
			shouldError:  true,
			errorMessage: "WordPress API returned status: 500",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a test server that returns predefined responses
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Check request URL for expected params
				if r.URL.Path != "/wp-json/wp/v2/pages" {
					t.Errorf("Expected path /wp-json/wp/v2/pages, got %s", r.URL.Path)
				}

				q := r.URL.Query()
				slug := q.Get("slug")
				lang := q.Get("lang")

				if slug != tc.expectedSlug {
					t.Errorf("Expected slug %s, got %s", tc.expectedSlug, slug)
				}

				if lang != tc.expectedLang {
					t.Errorf("Expected lang %s, got %s", tc.expectedLang, lang)
				}

				// Special case for error testing
				if slug == "error-page" {
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte("Internal server error"))
					return
				}

				// Return mocked response
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(tc.mockedResponse)
			}))
			defer server.Close()

			// Create WordPress client pointing to our test server
			client := &WordPressClient{
				BaseURL:  server.URL,
				MenuIdEn: "1",
				MenuIdFr: "2",
			}

			// Call the method being tested
			page, err := client.FetchPage(tc.path)

			// Verify results
			if tc.shouldError {
				if err == nil {
					t.Errorf("Expected error, got nil")
				} else if !strings.Contains(err.Error(), tc.errorMessage) {
					t.Errorf("Expected error to contain %q, got %q", tc.errorMessage, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
				if page == nil {
					t.Fatalf("Expected page, got nil")
				}
				if page.Title.Rendered != tc.mockedResponse[0].Title.Rendered {
					t.Errorf("Expected title %q, got %q", tc.mockedResponse[0].Title.Rendered, page.Title.Rendered)
				}
			}
		})
	}
}

// TestFetchPageWithTrailingSlash ensures that paths with trailing slashes work correctly
func TestFetchPageWithTrailingSlash(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		slug := q.Get("slug")

		// Verify trailing slash was removed
		if slug != "about-us" {
			t.Errorf("Expected slug 'about-us', got %s", slug)
		}

		response := []models.WordPressPage{
			{
				ID: 123,
				Title: Rendered{
					Rendered: "About Us",
				},
				Slug: "about-us",
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := &WordPressClient{BaseURL: server.URL}
	page, err := client.FetchPage("/about-us/")

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if page == nil {
		t.Fatalf("Expected page, got nil")
	}
	if page.Title.Rendered != "About Us" {
		t.Errorf("Expected title 'About Us', got %q", page.Title.Rendered)
	}
}

// TestFetchPageNetworkError tests handling of network errors
func TestFetchPageNetworkError(t *testing.T) {
	// Create client with invalid URL to trigger network error
	client := &WordPressClient{BaseURL: "http://invalid-domain-that-does-not-exist.example"}

	_, err := client.FetchPage("/any-page")

	if err == nil {
		t.Errorf("Expected network error, got nil")
	}
}

// TestFetchMenu tests the FetchMenu method which retrieves menu items for a specific language
func TestFetchMenu(t *testing.T) {
	testCases := []struct {
		name           string
		language       string
		expectedMenuId string
		mockedResponse []models.WordPressMenuItem
		shouldError    bool
		errorMessage   string
	}{
		{
			name:           "English menu",
			language:       "en",
			expectedMenuId: "123",
			mockedResponse: []models.WordPressMenuItem{
				{
					ID: 1,
					Title: Rendered{
						Rendered: "Home",
					},
					Url: "https://example.com/",
				},
				{
					ID: 2,
					Title: Rendered{
						Rendered: "About",
					},
					Url: "https://example.com/about",
				},
			},
		},
		{
			name:           "French menu",
			language:       "fr",
			expectedMenuId: "456",
			mockedResponse: []models.WordPressMenuItem{
				{
					ID: 3,
					Title: Rendered{
						Rendered: "Accueil",
					},
					Url: "https://example.com/fr",
				},
				{
					ID: 4,
					Title: Rendered{
						Rendered: "À propos",
					},
					Url: "https://example.com/fr/a-propos",
				},
			},
		},
		{
			name:           "API error response",
			language:       "en",
			expectedMenuId: "123",
			shouldError:    true,
			errorMessage:   "WordPress API returned status: 500",
		},
		{
			name:           "Invalid JSON response",
			language:       "en",
			expectedMenuId: "123",
			shouldError:    true,
			errorMessage:   "invalid character",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify the request path
				if r.URL.Path != "/wp-json/wp/v2/menu-items" {
					t.Errorf("Expected path /wp-json/wp/v2/menu-items, got %s", r.URL.Path)
				}

				// Check query parameters
				q := r.URL.Query()
				if q.Get("menus") != tc.expectedMenuId {
					t.Errorf("Expected menus=%s, got %s", tc.expectedMenuId, q.Get("menus"))
				}

				// Verify authorization header is present
				authHeader := r.Header.Get("Authorization")
				if !strings.HasPrefix(authHeader, "Basic ") {
					t.Errorf("Expected Authorization header with Basic auth, got: %s", authHeader)
				}

				// Handle error cases
				if tc.name == "API error response" {
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte("Internal server error"))
					return
				}

				// Handle invalid JSON case
				if tc.name == "Invalid JSON response" {
					w.Header().Set("Content-Type", "application/json")
					w.Write([]byte("This is not valid JSON"))
					return
				}

				// Return mocked response for success cases
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(tc.mockedResponse)
			}))
			defer server.Close()

			// Create WordPress client pointing to test server
			client := &WordPressClient{
				BaseURL:       server.URL,
				WordPressAuth: "dGVzdHVzZXI6dGVzdHBhc3N3b3Jk", // Base64 of "testuser:testpassword"
				MenuIdEn:      "123",
				MenuIdFr:      "456",
			}

			// Call the method being tested
			menuItems, err := client.FetchMenu(tc.language)

			// Verify results
			if tc.shouldError {
				if err == nil {
					t.Errorf("Expected error, got nil")
				} else if !strings.Contains(err.Error(), tc.errorMessage) {
					t.Errorf("Expected error containing %q, got %q", tc.errorMessage, err.Error())
				}
				return
			}

			// Check success cases
			if err != nil {
				t.Errorf("Expected no error, got %v", err)
				return
			}

			if menuItems == nil {
				t.Fatal("Expected menu items, got nil")
			}

			if len(*menuItems) != len(tc.mockedResponse) {
				t.Errorf("Expected %d menu items, got %d", len(tc.mockedResponse), len(*menuItems))
			}

			// Verify content of menu items
			for i, item := range *menuItems {
				if item.Title.Rendered != tc.mockedResponse[i].Title.Rendered {
					t.Errorf("Expected menu item title %q, got %q", tc.mockedResponse[i].Title.Rendered, item.Title.Rendered)
				}
				if item.Url != tc.mockedResponse[i].Url {
					t.Errorf("Expected menu item URL %q, got %q", tc.mockedResponse[i].Url, item.Url)
				}
			}
		})
	}
}

// TestNewWordPressClient tests the client initialization and concurrent menu fetching
func TestNewWordPressClient(t *testing.T) {
	// Mock server to respond to menu requests
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract path and query parameters
		if !strings.HasPrefix(r.URL.Path, "/wp-json/wp/v2/menu-items") {
			t.Errorf("Unexpected URL path: %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			return
		}

		// Verify authorization header is present
		authHeader := r.Header.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Basic ") {
			t.Errorf("Expected Authorization header with Basic auth, got: %s", authHeader)
		}

		// Check language-specific menus
		q := r.URL.Query()
		menuId := q.Get("menus")

		var menuItems []models.WordPressMenuItem
		switch menuId {
		case "123": // English menu
			menuItems = []models.WordPressMenuItem{
				{
					ID: 1,
					Title: Rendered{
						Rendered: "Home",
					},
					Url: "https://example.com/",
				},
				{
					ID: 2,
					Title: Rendered{
						Rendered: "About",
					},
					Url: "https://example.com/about",
				},
			}
		case "456": // French menu
			menuItems = []models.WordPressMenuItem{
				{
					ID: 3,
					Title: Rendered{
						Rendered: "Accueil",
					},
					Url: "https://example.com/fr",
				},
				{
					ID: 4,
					Title: Rendered{
						Rendered: "À propos",
					},
					Url: "https://example.com/fr/a-propos",
				},
			}
		default:
			t.Errorf("Unexpected menu ID: %s", menuId)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(menuItems)
	}))
	defer server.Close()

	// Test parameters
	baseURL := server.URL
	username := "testuser"
	password := "testpassword"
	menuIdEn := "123"
	menuIdFr := "456"

	// Create client - this will trigger concurrent menu fetches
	client := NewWordPressClient(baseURL, username, password, menuIdEn, menuIdFr)

	// Verify client initialization
	expectedAuth := base64.StdEncoding.EncodeToString([]byte(username + ":" + password))
	if client.BaseURL != baseURL {
		t.Errorf("Expected BaseURL %s, got %s", baseURL, client.BaseURL)
	}
	if client.WordPressAuth != expectedAuth {
		t.Errorf("Expected WordPressAuth %s, got %s", expectedAuth, client.WordPressAuth)
	}
	if client.MenuIdEn != menuIdEn {
		t.Errorf("Expected MenuIdEn %s, got %s", menuIdEn, client.MenuIdEn)
	}
	if client.MenuIdFr != menuIdFr {
		t.Errorf("Expected MenuIdFr %s, got %s", menuIdFr, client.MenuIdFr)
	}

	// Verify menus were fetched and processed
	expectedLanguages := []string{"en", "fr"}
	for _, lang := range expectedLanguages {
		menu, exists := client.Menus[lang]
		if !exists {
			t.Errorf("Expected menu for language %s to be present", lang)
			continue
		}

		// Verify menu items were processed correctly
		if menu == nil {
			t.Errorf("Menu for language %s is nil", lang)
			continue
		}

		// Verify menu structure (top-level items and their children)
		expectedItemCount := 2 // Both English and French menus have 2 items
		if len(menu.Items) != expectedItemCount {
			t.Errorf("Expected %d top-level menu items for %s, got %d",
				expectedItemCount, lang, len(menu.Items))
		}
	}
}
