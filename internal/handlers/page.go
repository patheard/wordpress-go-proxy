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

// PageHandler handles requests for WordPress pages
type PageHandler struct {
	wpClient  *api.WordPressClient
	templates *template.Template
}

// NewPageHandler creates a new page handler with the given WordPress base URL
func NewPageHandler(wordpressBaseURL string, wordpressUsername string, wordpressPassword string, wordpressMenuIdEn string, wordpressMeuIdFr string) *PageHandler {
	// Load templates
	tmpl, err := template.ParseFiles("templates/layout.html")
	if err != nil {
		log.Fatal("Error parsing template:", err)
	}

	return &PageHandler{
		wpClient:  api.NewWordPressClient(wordpressBaseURL, wordpressUsername, wordpressPassword, wordpressMenuIdEn, wordpressMeuIdFr),
		templates: tmpl,
	}
}

// ServeHTTP implements the http.Handler interface
func (h *PageHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	log.Printf("Page request: %s", path)

	// Check if path has the correct format
	segments := strings.Split(strings.Trim(path, "/"), "/")
	if path == "/" || len(segments) == 0 {
		path = "home" // Default to home page
	} else {
		// Check that the last segment doesn't have a file extension
		lastSegment := segments[len(segments)-1]
		if ext := filepath.Ext(lastSegment); ext != "" {
			log.Printf("Invalid path: Last segment contains file extension: %s", path)
			http.NotFound(w, r)
			return
		}
	}

	// Handle the page request
	h.handlePage(w, r, path)
}

// handlePage processes a page request
func (h *PageHandler) handlePage(w http.ResponseWriter, _ *http.Request, path string) {
	wpPage, err := h.wpClient.FetchPage(path)
	if err != nil {
		http.Error(w, "Error fetching page content", http.StatusInternalServerError)
		log.Printf("Error fetching page: %v", err)
		return
	}

	wpMenu, ok := h.wpClient.Menus[wpPage.Lang]
	if !ok {
		log.Printf("Warning: No menu found for language %s defaulting to 'en'", wpPage.Lang)
		wpMenu = h.wpClient.Menus["en"]
	}
	data := models.NewPageData(wpPage, wpMenu)

	log.Printf("Rendering page template")
	err = h.templates.ExecuteTemplate(w, "layout.html", data)
	if err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
		log.Printf("Error rendering template: %v", err)
		return
	}
}
