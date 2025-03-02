package main

import (
	"log"
	"net/http"

	"wordpress-go-proxy/internal/api"
	"wordpress-go-proxy/internal/config"
	"wordpress-go-proxy/internal/handlers"
	"wordpress-go-proxy/internal/middleware"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"
	_ "golang.org/x/crypto/x509roots/fallback"
)

func main() {

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Error loading config: ", err)
	}

	// Create WordPress client.  This will fetch menus asynchronously.
	wordPressClient := api.NewWordPressClient(
		cfg.WordPressBaseURL,
		cfg.WordPressUsername,
		cfg.WordPressPassword,
		cfg.WordPressMenuIdEn,
		cfg.WordPressMenuIdFr)

	siteNames := map[string]string{
		"en": cfg.SiteNameEn,
		"fr": cfg.SiteNameFr,
	}

	// Set up routes
	http.Handle("/static/", http.StripPrefix("/static/", handlers.NewStaticHandler("static")))
	http.Handle("/", middleware.SecurityHeaders(handlers.NewPageHandler(siteNames, wordPressClient)))

	// Start Lambda proxy handler
	lambda.Start(httpadapter.NewV2(http.DefaultServeMux).ProxyWithContext)
}
