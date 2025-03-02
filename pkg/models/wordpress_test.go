package models

import (
	"strings"
	"testing"
)

// TestNewPageData tests the NewPageData function which creates page rendering data
func TestNewPageData(t *testing.T) {
	testCases := []struct {
		name         string
		page         WordPressPage
		menu         *MenuData
		siteNames    map[string]string
		baseUrl      string
		expectedData PageData
	}{
		{
			name: "English page",
			page: WordPressPage{
				ID:       1,
				Slug:     "about",
				SlugEn:   "about",
				SlugFr:   "a-propos",
				Lang:     "en",
				Modified: "2023-05-15T10:30:45",
				Title: struct {
					Rendered string `json:"rendered"`
				}{Rendered: "About Us"},
				Content: struct {
					Rendered string `json:"rendered"`
					Raw      string `json:"raw,omitempty"`
				}{Rendered: "<p>This is content with https://example.com/image.jpg</p>"},
			},
			menu: &MenuData{
				Items: []*MenuItemData{},
			},
			siteNames: map[string]string{
				"en": "English Site Name",
				"fr": "French Site Name",
			},
			baseUrl: "https://example.com",
			expectedData: PageData{
				Lang:           "en",
				LangSwapPath:   "/fr/",
				LangSwapSlug:   "a-propos",
				Home:           "/",
				Modified:       "2023-05-15",
				Title:          "About Us",
				Content:        "<p>This is content with /image.jpg</p>",
				ShowBreadcrumb: true,
				SiteName:       "English Site Name",
			},
		},
		{
			name: "French page",
			page: WordPressPage{
				ID:       2,
				Slug:     "a-propos",
				SlugEn:   "about",
				SlugFr:   "a-propos",
				Lang:     "fr",
				Modified: "2023-05-15T10:30:45",
				Title: struct {
					Rendered string `json:"rendered"`
				}{Rendered: "À propos"},
				Content: struct {
					Rendered string `json:"rendered"`
					Raw      string `json:"raw,omitempty"`
				}{Rendered: "<p>C'est du contenu avec https://example.com/image.jpg</p>"},
			},
			menu: &MenuData{
				Items: []*MenuItemData{},
			},
			siteNames: map[string]string{
				"en": "English Site Name",
				"fr": "French Site Name",
			},
			baseUrl: "https://example.com",
			expectedData: PageData{
				Lang:           "fr",
				LangSwapPath:   "/",
				LangSwapSlug:   "about",
				Home:           "/fr/",
				Modified:       "2023-05-15",
				Title:          "À propos",
				Content:        "<p>C'est du contenu avec /image.jpg</p>",
				ShowBreadcrumb: true,
				SiteName:       "French Site Name",
			},
		},
		{
			name: "Invalid language defaulting to English",
			page: WordPressPage{
				ID:       3,
				Slug:     "about",
				SlugEn:   "about",
				SlugFr:   "a-propos",
				Lang:     "es", // Invalid language
				Modified: "2023-05-15T10:30:45",
				Title: struct {
					Rendered string `json:"rendered"`
				}{Rendered: "About Us"},
				Content: struct {
					Rendered string `json:"rendered"`
					Raw      string `json:"raw,omitempty"`
				}{Rendered: "<p>Content</p>"},
			},
			menu: &MenuData{
				Items: []*MenuItemData{},
			},
			siteNames: map[string]string{
				"en": "English Site Name",
				"fr": "French Site Name",
			},
			baseUrl: "https://example.com",
			expectedData: PageData{
				Lang:           "en",
				LangSwapPath:   "/fr/",
				LangSwapSlug:   "a-propos",
				Home:           "/",
				Modified:       "2023-05-15",
				Title:          "About Us",
				Content:        "<p>Content</p>",
				ShowBreadcrumb: true,
				SiteName:       "English Site Name",
			},
		},
		{
			name: "Home page (no breadcrumb)",
			page: WordPressPage{
				ID:       4,
				Slug:     "home",
				SlugEn:   "home",
				SlugFr:   "accueil",
				Lang:     "en",
				Modified: "2023-05-15T10:30:45",
				Title: struct {
					Rendered string `json:"rendered"`
				}{Rendered: "Home Page"},
				Content: struct {
					Rendered string `json:"rendered"`
					Raw      string `json:"raw,omitempty"`
				}{Rendered: "<p>Welcome home</p>"},
			},
			menu: &MenuData{
				Items: []*MenuItemData{},
			},
			siteNames: map[string]string{
				"en": "English Site Name",
				"fr": "French Site Name",
			},
			baseUrl: "https://example.com",
			expectedData: PageData{
				Lang:           "en",
				LangSwapPath:   "/fr/",
				LangSwapSlug:   "accueil",
				Home:           "/",
				Modified:       "2023-05-15",
				Title:          "Home Page",
				Content:        "<p>Welcome home</p>",
				ShowBreadcrumb: false, // Home page, no breadcrumb
				SiteName:       "English Site Name",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a copy of the page as we're passing a pointer
			page := tc.page

			// Call the function being tested
			result := NewPageData(&page, tc.menu, tc.siteNames, tc.baseUrl)

			// Verify results
			if result.Lang != tc.expectedData.Lang {
				t.Errorf("Expected Lang %q, got %q", tc.expectedData.Lang, result.Lang)
			}

			if result.LangSwapPath != tc.expectedData.LangSwapPath {
				t.Errorf("Expected LangSwapPath %q, got %q", tc.expectedData.LangSwapPath, result.LangSwapPath)
			}

			if result.LangSwapSlug != tc.expectedData.LangSwapSlug {
				t.Errorf("Expected LangSwapSlug %q, got %q", tc.expectedData.LangSwapSlug, result.LangSwapSlug)
			}

			if result.Home != tc.expectedData.Home {
				t.Errorf("Expected Home %q, got %q", tc.expectedData.Home, result.Home)
			}

			if result.Modified != tc.expectedData.Modified {
				t.Errorf("Expected Modified %q, got %q", tc.expectedData.Modified, result.Modified)
			}

			if string(result.Title) != string(tc.expectedData.Title) {
				t.Errorf("Expected Title %q, got %q", tc.expectedData.Title, result.Title)
			}

			if string(result.Content) != string(tc.expectedData.Content) {
				t.Errorf("Expected Content %q, got %q", tc.expectedData.Content, result.Content)
			}

			if result.ShowBreadcrumb != tc.expectedData.ShowBreadcrumb {
				t.Errorf("Expected ShowBreadcrumb %v, got %v", tc.expectedData.ShowBreadcrumb, result.ShowBreadcrumb)
			}

			if result.SiteName != tc.expectedData.SiteName {
				t.Errorf("Expected SiteName %q, got %q", tc.expectedData.SiteName, result.SiteName)
			}

			// Menu is passed by reference, so it should be the same object
			if result.Menu != tc.menu {
				t.Errorf("Expected Menu to be the same object that was passed in")
			}
		})
	}
}

// TestNewMenuData tests the NewMenuData function which creates hierarchical menu data
func TestNewMenuData(t *testing.T) {
	testCases := []struct {
		name             string
		menuItems        []WordPressMenuItem
		baseUrl          string
		expectedTopItems int
		expectedChildren map[string]int // Map of parent title to number of children
	}{
		{
			name: "Simple menu with no children",
			menuItems: []WordPressMenuItem{
				{
					ID: 1,
					Title: struct {
						Rendered string `json:"rendered"`
					}{Rendered: "Home"},
					Parent: 0,
					Url:    "https://example.com/",
				},
				{
					ID: 2,
					Title: struct {
						Rendered string `json:"rendered"`
					}{Rendered: "About"},
					Parent: 0,
					Url:    "https://example.com/about",
				},
			},
			baseUrl:          "https://example.com",
			expectedTopItems: 2,
			expectedChildren: map[string]int{
				"Home":  0,
				"About": 0,
			},
		},
		{
			name: "Menu with parent-child relationships",
			menuItems: []WordPressMenuItem{
				{
					ID: 1,
					Title: struct {
						Rendered string `json:"rendered"`
					}{Rendered: "Home"},
					Parent: 0,
					Url:    "https://example.com/",
				},
				{
					ID: 2,
					Title: struct {
						Rendered string `json:"rendered"`
					}{Rendered: "Products"},
					Parent: 0,
					Url:    "https://example.com/products",
				},
				{
					ID: 3,
					Title: struct {
						Rendered string `json:"rendered"`
					}{Rendered: "Product A"},
					Parent: 2, // Child of Products
					Url:    "https://example.com/products/a",
				},
				{
					ID: 4,
					Title: struct {
						Rendered string `json:"rendered"`
					}{Rendered: "Product B"},
					Parent: 2, // Child of Products
					Url:    "https://example.com/products/b",
				},
				{
					ID: 5,
					Title: struct {
						Rendered string `json:"rendered"`
					}{Rendered: "About"},
					Parent: 0,
					Url:    "https://example.com/about",
				},
			},
			baseUrl:          "https://example.com",
			expectedTopItems: 3,
			expectedChildren: map[string]int{
				"Home":     0,
				"Products": 2,
				"About":    0,
			},
		},
		{
			name: "Menu with URL base replacement",
			menuItems: []WordPressMenuItem{
				{
					ID: 1,
					Title: struct {
						Rendered string `json:"rendered"`
					}{Rendered: "Home"},
					Parent: 0,
					Url:    "https://other-domain.com/",
				},
				{
					ID: 2,
					Title: struct {
						Rendered string `json:"rendered"`
					}{Rendered: "About"},
					Parent: 0,
					Url:    "https://other-domain.com/about",
				},
			},
			baseUrl:          "https://other-domain.com",
			expectedTopItems: 2,
			expectedChildren: map[string]int{
				"Home":  0,
				"About": 0,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Copy the menu items as we're passing a pointer
			menuItems := tc.menuItems

			// Call the function being tested
			result := NewMenuData(&menuItems, tc.baseUrl)

			// Verify results
			if len(result.Items) != tc.expectedTopItems {
				t.Errorf("Expected %d top-level menu items, got %d", tc.expectedTopItems, len(result.Items))
			}

			// Create a map to easily look up items by title
			itemMap := make(map[string]*MenuItemData)
			for _, item := range result.Items {
				itemMap[item.Title] = item

				// Check if baseUrl is properly replaced in URLs
				if strings.Contains(item.Url, tc.baseUrl) {
					t.Errorf("URL should not contain base URL %q, got %q", tc.baseUrl, item.Url)
				}
			}

			// Check children counts
			for title, expectedCount := range tc.expectedChildren {
				item, exists := itemMap[title]
				if !exists {
					t.Errorf("Expected menu item with title %q not found", title)
					continue
				}

				actualCount := len(item.Children)
				if actualCount != expectedCount {
					t.Errorf("Expected item %q to have %d children, got %d", title, expectedCount, actualCount)
				}

				// For items with children, verify they don't contain the base URL
				for _, child := range item.Children {
					if strings.Contains(child.Url, tc.baseUrl) {
						t.Errorf("Child URL should not contain base URL %q, got %q", tc.baseUrl, child.Url)
					}
				}
			}

			// For nested relationships, check parent-child connections
			if tc.name == "Menu with multiple levels of nesting" {
				productsItem := itemMap["Products"]
				if len(productsItem.Children) != 1 {
					t.Fatalf("Expected Products to have 1 child, got %d", len(productsItem.Children))
				}

				categoryA := productsItem.Children[0]
				if categoryA.Title != "Category A" {
					t.Errorf("Expected child of Products to be 'Category A', got %q", categoryA.Title)
				}

				if len(categoryA.Children) != 1 {
					t.Fatalf("Expected Category A to have 1 child, got %d", len(categoryA.Children))
				}

				productA1 := categoryA.Children[0]
				if productA1.Title != "Product A1" {
					t.Errorf("Expected child of Category A to be 'Product A1', got %q", productA1.Title)
				}
			}
		})
	}
}
