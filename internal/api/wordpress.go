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

type WordPressClient struct {
	BaseURL       string
	WordPressAuth string
	Menus         map[string]*models.MenuData
	MenuIdEn      string
	MenuIdFr      string
}

type menuResult struct {
	lang      string
	menuItems *[]models.WordPressMenuItem
	err       error
}

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
	results := make(chan menuResult, len(languages))
	for _, lang := range languages {
		go func(language string) {
			menuItems, err := client.FetchMenu(language)
			results <- menuResult{lang: language, menuItems: menuItems, err: err}
		}(lang)
	}

	// Wait for both requests to complete
	for range languages {
		result := <-results
		if result.err != nil {
			log.Fatalf("Error fetching menu items for %s: %v", result.lang, result.err)
		}
		log.Printf("Fetched %d menu items for %s", len(*result.menuItems), result.lang)
		client.Menus[result.lang] = models.NewMenuData(result.menuItems, baseURL)
	}

	return client
}

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

func (c *WordPressClient) FetchPage(path string) (*models.WordPressPage, error) {
	// Get the last segment of the path
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
