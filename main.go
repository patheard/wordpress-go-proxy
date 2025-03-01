package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// WordPressPage represents the WordPress API response structure
type WordPressPage struct {
	ID       int    `json:"id"`
	Lang     string `json:"lang"`
	Modified string `json:"modified"`
	Content  struct {
		Rendered string `json:"rendered"`
	} `json:"content"`
	Title struct {
		Rendered string `json:"rendered"`
	} `json:"title"`
}

// PageData holds the data to be passed to our HTML template
type PageData struct {
	Lang     string
	Modified string
	Title    string
	Content  template.HTML
}

var (
	wpBaseURL = "http://articles.alpha.canada.ca/pcs-superset/wp-json/wp/v2"
	templates *template.Template
)

func init() {
	// Load templates
	var err error
	templates, err = template.ParseFiles("templates/layout.html")
	if err != nil {
		log.Fatal("Error parsing template:", err)
	}

	mime.AddExtensionType(".js", "application/javascript")
	mime.AddExtensionType(".css", "text/css")
}

func main() {
	// Set up routes
	staticFS := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", staticContentHandler(staticFS)))
	http.HandleFunc("/", wordpressPageHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Server starting on port %s...\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func addDefaultHeaders(w http.ResponseWriter) http.ResponseWriter {
	w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
	w.Header().Set("X-Frame-Options", "SAMEORIGIN")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Referrer-Policy", "no-referrer-when-downgrade")
	return w
}

func wordpressPageHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	log.Printf("Path: %s", path)

	// Check if path has the correct format
	segments := strings.Split(strings.Trim(path, "/"), "/")
	if len(segments) > 0 {
		lastSegment := segments[len(segments)-1]
		if ext := filepath.Ext(lastSegment); ext != "" {
			log.Printf("Invalid path: Last segment contains file extension: %s", path)
			http.NotFound(w, r)
			return
		}

		w = addDefaultHeaders(w)
		handlePage(w, r)
		return
	}

	// If path doesn't match our required pattern, return 404
	log.Printf("Invalid path format: %s", path)
	http.NotFound(w, r)
}

func handlePage(w http.ResponseWriter, r *http.Request) {
	// Get the path from the request
	path := r.URL.Path
	if path == "/" {
		path = "home" // Default to home page
	}

	// Fetch content from WordPress API
	wpPage, err := fetchWordPressPage(path)
	if err != nil {
		http.Error(w, "Error fetching page content", http.StatusInternalServerError)
		log.Printf("Error fetching page: %v", err)
		return
	}

	// Prepare data for template
	data := PageData{
		Lang:     wpPage.Lang,
		Modified: strings.Split(wpPage.Modified, "T")[0],
		Title:    wpPage.Title.Rendered,
		Content:  template.HTML(wpPage.Content.Rendered),
	}

	// Render template
	err = templates.ExecuteTemplate(w, "layout.html", data)
	if err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
		log.Printf("Error rendering template: %v", err)
		return
	}
}

func staticContentHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get file extension
		ext := filepath.Ext(r.URL.Path)
		log.Printf("Serving file: %s", r.URL.Path)

		// Set the content type based on file extension
		if ext != "" {
			mimeType := mime.TypeByExtension(ext)
			if mimeType != "" {
				w.Header().Set("Content-Type", mimeType)
			}
		}

		// Let the original handler serve the file
		h.ServeHTTP(w, r)
	})
}

func fetchWordPressPage(path string) (*WordPressPage, error) {
	// Get the last segment of the path
	path = strings.TrimSuffix(path, "/")
	slug := path[strings.LastIndex(path, "/")+1:]

	// Make request to WordPress API
	log.Printf("Fetching page with slug: %s", slug)
	resp, err := http.Get(fmt.Sprintf("%s/pages?slug=%s", wpBaseURL, slug))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("WordPress API returned status: %d", resp.StatusCode)
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Parse JSON response
	var pages []WordPressPage
	err = json.Unmarshal(body, &pages)
	if err != nil {
		return nil, err
	}

	if len(pages) == 0 {
		return nil, fmt.Errorf("page not found")
	}

	return &pages[0], nil
}
