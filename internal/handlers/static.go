package handlers

import (
	"log"
	"mime"
	"net/http"
	"path/filepath"
)

// StaticHandler handles static file requests
type StaticHandler struct {
	fileServer http.Handler
	staticDir  string
}

// NewStaticHandler creates a new static file handler
func NewStaticHandler(staticDir string) *StaticHandler {
	return &StaticHandler{
		fileServer: http.FileServer(http.Dir(staticDir)),
		staticDir:  staticDir,
	}
}

// ServeHTTP implements the http.Handler interface
func (h *StaticHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Get file extension
	ext := filepath.Ext(r.URL.Path)
	log.Printf("Serving static file: %s", r.URL.Path)

	// Set the content type based on file extension
	if ext != "" {
		mimeType := mime.TypeByExtension(ext)
		if mimeType != "" {
			w.Header().Set("Content-Type", mimeType)
		}
	}

	// Set cache control headers for static assets
	w.Header().Set("Cache-Control", "public, max-age=604800") // 7 days

	h.fileServer.ServeHTTP(w, r)
}
