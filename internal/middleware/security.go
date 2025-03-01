package middleware

import (
	"log"
	"net/http"
)

func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Add security headers
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
		w.Header().Set("X-Frame-Options", "SAMEORIGIN")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "no-referrer-when-downgrade")

		log.Printf("Adding security headers to request: %s", r.URL.Path)

		// Call the next handler
		next.ServeHTTP(w, r)
	})
}
