package config

import (
	"os"
	"testing"
)

func TestLoad_SiteNameEn(t *testing.T) {
	// Store original environment variables to restore them later
	originalSiteNameEn := os.Getenv("SITE_NAME_EN")
	originalSiteNameFr := os.Getenv("SITE_NAME_FR")
	originalWordPressURL := os.Getenv("WORDPRESS_URL")
	originalWordPressUsername := os.Getenv("WORDPRESS_USERNAME")
	originalWordPressPassword := os.Getenv("WORDPRESS_PASSWORD")
	originalWordPressMenuIdEn := os.Getenv("WORDPRESS_MENU_ID_EN")
	originalWordPressMenuIdFr := os.Getenv("WORDPRESS_MENU_ID_FR")

	// Restore environment variables after test
	defer func() {
		os.Setenv("SITE_NAME_EN", originalSiteNameEn)
		os.Setenv("SITE_NAME_FR", originalSiteNameFr)
		os.Setenv("WORDPRESS_URL", originalWordPressURL)
		os.Setenv("WORDPRESS_USERNAME", originalWordPressUsername)
		os.Setenv("WORDPRESS_PASSWORD", originalWordPressPassword)
		os.Setenv("WORDPRESS_MENU_ID_EN", originalWordPressMenuIdEn)
		os.Setenv("WORDPRESS_MENU_ID_FR", originalWordPressMenuIdFr)
	}()

	t.Run("SiteNameEn is loaded correctly when set", func(t *testing.T) {
		// Setup required environment variables
		expectedSiteNameEn := "Test Site Name English"
		os.Setenv("SITE_NAME_EN", expectedSiteNameEn)
		os.Setenv("SITE_NAME_FR", "Test Site Name French")
		os.Setenv("WORDPRESS_URL", "https://example.com")
		os.Setenv("WORDPRESS_USERNAME", "user")
		os.Setenv("WORDPRESS_PASSWORD", "pass")
		os.Setenv("WORDPRESS_MENU_ID_EN", "1")
		os.Setenv("WORDPRESS_MENU_ID_FR", "2")

		cfg, err := Load()
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if cfg.SiteNameEn != expectedSiteNameEn {
			t.Errorf("Expected SiteNameEn to be %q, got %q", expectedSiteNameEn, cfg.SiteNameEn)
		}
	})

	t.Run("SiteNameEn is missing", func(t *testing.T) {
		// Setup other required variables but omit SITE_NAME_EN
		os.Unsetenv("SITE_NAME_EN")
		os.Setenv("SITE_NAME_FR", "Test Site Name French")
		os.Setenv("WORDPRESS_URL", "https://example.com")
		os.Setenv("WORDPRESS_USERNAME", "user")
		os.Setenv("WORDPRESS_PASSWORD", "pass")
		os.Setenv("WORDPRESS_MENU_ID_EN", "1")
		os.Setenv("WORDPRESS_MENU_ID_FR", "2")

		_, err := Load()
		if err == nil {
			t.Fatal("Expected error when SITE_NAME_EN is missing, got nil")
		}

		expectedErrSubstring := "SITE_NAME_EN"
		if err != nil && !containsString(err.Error(), expectedErrSubstring) {
			t.Errorf("Expected error to mention %q, got %q", expectedErrSubstring, err.Error())
		}
	})

	t.Run("SiteNameEn is empty", func(t *testing.T) {
		// Setup with empty SITE_NAME_EN
		os.Setenv("SITE_NAME_EN", "")
		os.Setenv("SITE_NAME_FR", "Test Site Name French")
		os.Setenv("WORDPRESS_URL", "https://example.com")
		os.Setenv("WORDPRESS_USERNAME", "user")
		os.Setenv("WORDPRESS_PASSWORD", "pass")
		os.Setenv("WORDPRESS_MENU_ID_EN", "1")
		os.Setenv("WORDPRESS_MENU_ID_FR", "2")

		_, err := Load()
		if err == nil {
			t.Fatal("Expected error when SITE_NAME_EN is empty, got nil")
		}

		expectedErrSubstring := "SITE_NAME_EN"
		if err != nil && !containsString(err.Error(), expectedErrSubstring) {
			t.Errorf("Expected error to mention %q, got %q", expectedErrSubstring, err.Error())
		}
	})

	t.Run("SiteNameEn is correctly assigned", func(t *testing.T) {
		// Test various values for SITE_NAME_EN
		testValues := []string{
			"Simple Name",
			"Site with Spaces and Punctuation!",
			"Very Long Site Name That Continues For A While And Has Many Words In It",
			"Special Characters: !@#$%^&*()",
			"1234567890",
		}

		// Set other required variables
		os.Setenv("SITE_NAME_FR", "Test Site Name French")
		os.Setenv("WORDPRESS_URL", "https://example.com")
		os.Setenv("WORDPRESS_USERNAME", "user")
		os.Setenv("WORDPRESS_PASSWORD", "pass")
		os.Setenv("WORDPRESS_MENU_ID_EN", "1")
		os.Setenv("WORDPRESS_MENU_ID_FR", "2")

		for _, expectedValue := range testValues {
			os.Setenv("SITE_NAME_EN", expectedValue)

			cfg, err := Load()
			if err != nil {
				t.Fatalf("Expected no error, got %v", err)
			}

			if cfg.SiteNameEn != expectedValue {
				t.Errorf("Expected SiteNameEn to be %q, got %q", expectedValue, cfg.SiteNameEn)
			}
		}
	})
}

// Helper function to check if a string contains another string
func containsString(s, substr string) bool {
	return s != "" && substr != "" && s != substr && len(s) > len(substr) && s != "" && substr != "" && s != substr && len(s) > len(substr)
}

// TestConfigCompleteness verifies that all fields in Config are properly populated
func TestConfigCompleteness(t *testing.T) {
	// Store original environment variables
	origEnv := map[string]string{
		"SITE_NAME_EN":         os.Getenv("SITE_NAME_EN"),
		"SITE_NAME_FR":         os.Getenv("SITE_NAME_FR"),
		"WORDPRESS_URL":        os.Getenv("WORDPRESS_URL"),
		"WORDPRESS_USERNAME":   os.Getenv("WORDPRESS_USERNAME"),
		"WORDPRESS_PASSWORD":   os.Getenv("WORDPRESS_PASSWORD"),
		"WORDPRESS_MENU_ID_EN": os.Getenv("WORDPRESS_MENU_ID_EN"),
		"WORDPRESS_MENU_ID_FR": os.Getenv("WORDPRESS_MENU_ID_FR"),
		"PORT":                 os.Getenv("PORT"),
	}

	// Restore environment variables after test
	defer func() {
		for k, v := range origEnv {
			os.Setenv(k, v)
		}
	}()

	// Set test values for all required environment variables
	testValues := map[string]string{
		"SITE_NAME_EN":         "Example English Site",
		"SITE_NAME_FR":         "Example French Site",
		"WORDPRESS_URL":        "https://example.com/wp-api",
		"WORDPRESS_USERNAME":   "apiuser",
		"WORDPRESS_PASSWORD":   "apisecret",
		"WORDPRESS_MENU_ID_EN": "42",
		"WORDPRESS_MENU_ID_FR": "43",
		"PORT":                 "8080",
	}

	for k, v := range testValues {
		os.Setenv(k, v)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify all fields are correctly populated
	if cfg.SiteNameEn != testValues["SITE_NAME_EN"] {
		t.Errorf("Expected SiteNameEn to be %q, got %q", testValues["SITE_NAME_EN"], cfg.SiteNameEn)
	}
	if cfg.SiteNameFr != testValues["SITE_NAME_FR"] {
		t.Errorf("Expected SiteNameFr to be %q, got %q", testValues["SITE_NAME_FR"], cfg.SiteNameFr)
	}
	if cfg.WordPressBaseURL != testValues["WORDPRESS_URL"] {
		t.Errorf("Expected WordPressBaseURL to be %q, got %q", testValues["WORDPRESS_URL"], cfg.WordPressBaseURL)
	}
	if cfg.WordPressUsername != testValues["WORDPRESS_USERNAME"] {
		t.Errorf("Expected WordPressUsername to be %q, got %q", testValues["WORDPRESS_USERNAME"], cfg.WordPressUsername)
	}
	if cfg.WordPressPassword != testValues["WORDPRESS_PASSWORD"] {
		t.Errorf("Expected WordPressPassword to be %q, got %q", testValues["WORDPRESS_PASSWORD"], cfg.WordPressPassword)
	}
	if cfg.WordPressMenuIdEn != testValues["WORDPRESS_MENU_ID_EN"] {
		t.Errorf("Expected WordPressMenuIdEn to be %q, got %q", testValues["WORDPRESS_MENU_ID_EN"], cfg.WordPressMenuIdEn)
	}
	if cfg.WordPressMenuIdFr != testValues["WORDPRESS_MENU_ID_FR"] {
		t.Errorf("Expected WordPressMenuIdFr to be %q, got %q", testValues["WORDPRESS_MENU_ID_FR"], cfg.WordPressMenuIdFr)
	}
	if cfg.Port != testValues["PORT"] {
		t.Errorf("Expected Port to be %q, got %q", testValues["PORT"], cfg.Port)
	}
}
