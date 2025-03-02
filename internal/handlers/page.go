package handlers

import (
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"strings"

	"wordpress-go-proxy/internal/api"
	"wordpress-go-proxy/pkg/models"
)

// PageHandler handles requests for WordPress pages.  It is responsible for
// fetching the page content from the WordPress API and rendering it using
// an HTML template.
type PageHandler struct {
	SiteNames       map[string]string
	WordPressClient *api.WordPressClient
	Templates       *template.Template
}

var parseTemplateFiles = template.ParseFiles

// NewPageHandler creates a new page handler that will be used
// to retrieve and render WordPress pages.
func NewPageHandler(siteNames map[string]string, wordPressClient *api.WordPressClient) *PageHandler {
	// Load templates
	tmpl, err := parseTemplateFiles("templates/layout.html")
	if err != nil {
		log.Fatal("Error parsing template:", err)
	}

	return &PageHandler{
		SiteNames:       siteNames,
		WordPressClient: wordPressClient,
		Templates:       tmpl,
	}
}

// ServeHTTP implements the http.Handler interface. It processes incoming
// requests for WordPress pages and renders them using an HTML template.
func (h *PageHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	log.Printf("Page request: %s", path)

	// Only allow GET, HEAD and OPTIONS methods
	if r.Method != http.MethodGet && r.Method != http.MethodHead && r.Method != http.MethodOptions {
		log.Printf("Invalid HTTP method: %s", r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Do not allow paths with file extensions
	if ext := filepath.Ext(path); ext != "" {
		log.Printf("Invalid path: contains file extension: %s", path)
		http.NotFound(w, r)
		return
	}

	// Check for invalid URL characters
	if strings.ContainsAny(path, "<>\"'%\\`^{}|") {
		log.Printf("URL contains invalid characters: %s", path)
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	// Prevent DoS via long URLs
	if len(path) > 255 {
		log.Printf("URL path too long: %d characters", len(path))
		http.Error(w, "URI too long", http.StatusRequestURITooLong)
		return
	}

	h.handlePage(w, r, path)
}

// handlePage processes a page request by retrieving the page content
// from the WordPress API and rendering it using an HTML template.
func (h *PageHandler) handlePage(w http.ResponseWriter, _ *http.Request, path string) {
	page, err := h.WordPressClient.FetchPage(path)
	if err != nil {
		http.Error(w, "Error fetching page content", http.StatusInternalServerError)
		log.Printf("Error fetching page: %v", err)
		return
	}

	menu, ok := h.WordPressClient.Menus[page.Lang]
	if !ok {
		log.Printf("Warning: No menu found for language %s defaulting to 'en'", page.Lang)
		menu = h.WordPressClient.Menus["en"]
	}

	data := models.NewPageData(page, menu, h.SiteNames, h.WordPressClient.BaseURL)

	log.Printf("Rendering page template")
	err = h.Templates.ExecuteTemplate(w, "layout.html", data)
	if err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
		log.Printf("Error rendering template: %v", err)
		return
	}
	log.Printf("Rendering page template complete")
}
