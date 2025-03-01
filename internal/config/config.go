package config

import (
	"fmt"
	"os"
)

// Config holds all application configuration
type Config struct {
	// Server settings
	Port       string
	SiteNameEn string
	SiteNameFr string

	// WordPress API settings
	WordPressBaseURL  string
	WordPressUsername string
	WordPressPassword string
	WordPressMenuIdEn string
	WordPressMenuIdFr string
}

// Load reads configuration from environment variables and sets defaults
func Load() (*Config, error) {
	cfg := &Config{}

	requiredVars := map[string]*string{
		"PORT":                 &cfg.Port,
		"SITE_NAME_EN":         &cfg.SiteNameEn,
		"SITE_NAME_FR":         &cfg.SiteNameFr,
		"WORDPRESS_URL":        &cfg.WordPressBaseURL,
		"WORDPRESS_USERNAME":   &cfg.WordPressUsername,
		"WORDPRESS_PASSWORD":   &cfg.WordPressPassword,
		"WORDPRESS_MENU_ID_EN": &cfg.WordPressMenuIdEn,
		"WORDPRESS_MENU_ID_FR": &cfg.WordPressMenuIdFr,
	}

	// Check all required variables
	var missingVars []string
	for name, ptr := range requiredVars {
		val := os.Getenv(name)
		if val == "" {
			missingVars = append(missingVars, name)
		} else {
			*ptr = val
		}
	}

	// Return error if any required variables are missing
	if len(missingVars) > 0 {
		return nil, fmt.Errorf("missing required environment variables: %v", missingVars)
	}

	return cfg, nil
}
