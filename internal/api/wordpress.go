package api

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"wordpress-go-proxy/pkg/models"
)

// WordPressClient handles communication with the WordPress REST API
// It manages authentication, caching of menus, and provides methods
// to fetch content from WordPress.
type WordPressClient struct {
	BaseURL       string
	WordPressAuth string
	Menus         map[string]*models.MenuData
	MenuIdEn      string
	MenuIdFr      string
}

// MenuResult represents the result of an asynchronous menu fetch operation
type MenuResult struct {
	Lang      string
	MenuItems *[]models.WordPressMenuItem
	Err       error
}

// NewWordPressClient creates and initializes a new WordPress API client.
// It performs authentication and fetches menus concurrently during initialization.
func NewWordPressClient(baseURL string, username string, password string, menuIdEn string, menuIdFr string) *WordPressClient {
	auth := base64.StdEncoding.EncodeToString([]byte(username + ":" + password))
	client := &WordPressClient{
		BaseURL:       baseURL,
		WordPressAuth: auth,
		MenuIdEn:      menuIdEn,
		MenuIdFr:      menuIdFr,
		Menus:         make(map[string]*models.MenuData),
	}

	// Launch concurrent requests to retrieve the menus
	languages := []string{"en", "fr"}
	results := make(chan MenuResult, len(languages))
	for _, lang := range languages {
		go func(language string) {
			menuItems, err := client.FetchMenu(language)
			results <- MenuResult{
				Lang:      language,
				MenuItems: menuItems,
				Err:       err}
		}(lang)
	}

	// Wait for both requests to complete
	for range languages {
		result := <-results
		if result.Err != nil {
			log.Fatalf("Error fetching menu items for %s: %v", result.Lang, result.Err)
		}
		log.Printf("Fetched %d menu items for %s", len(*result.MenuItems), result.Lang)
		client.Menus[result.Lang] = models.NewMenuData(result.MenuItems, baseURL)
	}

	return client
}

// FetchMenu retrieves the menu items for a given language.
func (c *WordPressClient) FetchMenu(lang string) (*[]models.WordPressMenuItem, error) {
	menuId := c.MenuIdEn
	if lang == "fr" {
		menuId = c.MenuIdFr
	}

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/wp-json/wp/v2/menu-items?menus=%s", c.BaseURL, menuId), nil)
	req.Header.Add("Authorization", "Basic "+c.WordPressAuth)
	if err != nil {
		return nil, err
	}

	// Execute the request
	client := &http.Client{
		Timeout: 3 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("WordPress API returned status: %d, body: %s", resp.StatusCode, string(body))
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Parse JSON response
	var menuItems []models.WordPressMenuItem
	err = json.Unmarshal(body, &menuItems)
	if err != nil {
		return nil, err
	}

	return &menuItems, nil
}

// FetchPage retrieves a page from WordPress by its path.
// The path is split and the last segment is the slug used to fetch the page.
// The language is determined by the second segment of the path.
func (c *WordPressClient) FetchPage(path string) (*models.WordPressPage, error) {
	path = strings.TrimSuffix(path, "/")
	slug := path[strings.LastIndex(path, "/")+1:]
	segments := strings.Split(path, "/")

	lang := "en"
	if len(segments) > 1 && segments[1] == "fr" {
		lang = "fr"
	}

	homePages := map[string]string{
		"":   "home",
		"fr": "home-fr",
	}
	if homeSlug, isHome := homePages[slug]; isHome {
		slug = homeSlug
	}

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/wp-json/wp/v2/pages?slug=%s&lang=%s", c.BaseURL, slug, lang), nil)
	if err != nil {
		return nil, err
	}

	log.Printf("Fetching page: %s", req.URL.String())
	client := &http.Client{
		Timeout: 3 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("WordPress API returned status: %d, body: %s", resp.StatusCode, string(body))
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Parse JSON response
	var pages []models.WordPressPage
	err = json.Unmarshal(body, &pages)
	if err != nil {
		return nil, err
	}

	if len(pages) == 0 {
		return nil, fmt.Errorf("page not found")
	}

	return &pages[0], nil
}
