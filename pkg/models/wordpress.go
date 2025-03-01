package models

import (
	"encoding/json"
	"fmt"
	"html/template"
	"time"
)

// WordPressPage represents the WordPress API response structure
type WordPressPage struct {
	ID       int    `json:"id"`
	Slug     string `json:"slug"`
	Lang     string `json:"lang"`
	Modified string `json:"modified"`
	Content  struct {
		Rendered string `json:"rendered"`
		Raw      string `json:"raw,omitempty"`
	} `json:"content"`
	Title struct {
		Rendered string `json:"rendered"`
	} `json:"title"`
	Excerpt struct {
		Rendered string `json:"rendered,omitempty"`
	} `json:"excerpt,omitempty"`
	FeaturedMedia int   `json:"featured_media,omitempty"`
	Categories    []int `json:"categories,omitempty"`
}

type WordPressMenuItem struct {
	ID    int `json:"id"`
	Title struct {
		Rendered string `json:"rendered"`
	} `json:"title"`
	Parent int    `json:"parent"`
	Url    string `json:"url"`
}

// PageData holds the data to be passed to our HTML template
type PageData struct {
	Lang     string
	Modified string
	Title    string
	Content  template.HTML
	Excerpt  template.HTML
	Path     string
	Metadata map[string]string
}

// Breadcrumb represents a single item in a breadcrumb trail
type Breadcrumb struct {
	Label  string
	URL    string
	Active bool
}

// NewPageData creates a new PageData instance from a WordPress page
func NewPageData(page *WordPressPage, path string) PageData {
	// Format the modified date
	formattedDate := page.Modified
	if parsedDate, err := time.Parse("2006-01-02T15:04:05", page.Modified); err == nil {
		formattedDate = parsedDate.Format("2006-01-02")
	} else if parsedDate, err := time.Parse(time.RFC3339, page.Modified); err == nil {
		formattedDate = parsedDate.Format("2006-01-02")
	}

	// Initialize metadata map
	metadata := make(map[string]string)
	metadata["page_id"] = fmt.Sprintf("%d", page.ID)

	return PageData{
		Lang:     page.Lang,
		Modified: formattedDate,
		Title:    page.Title.Rendered,
		Content:  template.HTML(page.Content.Rendered),
		Excerpt:  template.HTML(page.Excerpt.Rendered),
		Path:     path,
		Metadata: metadata,
	}
}

// ToJSON converts PageData to JSON
func (pd *PageData) ToJSON() ([]byte, error) {
	return json.Marshal(pd)
}
