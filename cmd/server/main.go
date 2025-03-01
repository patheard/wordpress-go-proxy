package main

import (
	"log"
	"net/http"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"

	"wordpress-go-proxy/internal/config"
	"wordpress-go-proxy/internal/handlers"
	"wordpress-go-proxy/internal/middleware"
)

func main() {
	// Initialize config
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Error loading config: ", err)
	}

	// Set up routes
	http.Handle("/static/", http.StripPrefix("/static/", handlers.NewStaticHandler("static")))
	http.Handle("/", middleware.SecurityHeaders(handlers.NewPageHandler(cfg.WordPressBaseURL, cfg.WordPressUsername, cfg.WordPressPassword, cfg.WordPressMenuIdEn, cfg.WordPressMenuIdFr)))

	lambda.Start(httpadapter.NewV2(http.DefaultServeMux).ProxyWithContext)

	// Start server
	// fmt.Printf("Server starting on port %s...\n", cfg.Port)
	// log.Fatal(http.ListenAndServe(":"+cfg.Port, nil))
}
