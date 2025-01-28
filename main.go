package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/ericflores108/spotify/config"
	"github.com/ericflores108/spotify/httpserver"
	"github.com/ericflores108/spotify/logger"
	"github.com/ericflores108/spotify/service"
)

func main() {
	ctx := context.Background()

	err := logger.InitializeLoggers(ctx, config.GoogleProjectID)
	if err != nil {
		log.Fatalf("Failed to initialize loggers: %v", err)
	}

	logger.LogInfo("starting app")

	appConfig := config.GetConfig(ctx)
	defer appConfig.SecretManagerClient.Close()
	defer appConfig.FirestoreClient.Close()

	// Define a flag for the redirect URL
	useLocalHost := flag.Bool("useLocalHost", false, "Use localhost as the redirect URL (default: production URL)")
	flag.Parse()

	// Determine the redirect URL
	redirectURL := "https://titled96.com/callback" // Default production URL
	if *useLocalHost {
		redirectURL = "http://localhost:8080/callback"
	}

	// Initialize the service
	svc := service.NewService(
		appConfig.ClientID,
		appConfig.ClientSecret,
		redirectURL,
		"spotify_auth_state",
		appConfig.FirestoreClient,
		appConfig.OpenAIClient,
		appConfig.GeniusClient,
	)

	// Initialize the server and register routes
	srv := httpserver.NewServer(ctx, svc)
	mux := srv.RegisterRoutes()

	// Determine port for HTTP service
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		logger.LogInfo("defaulting to port %s", port)
	}

	// Start HTTP server
	logger.LogInfo("listening on port %s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		logger.LogError("Failed to start server on port %s: %v", port, err)
	}
}
