package main

import (
	"context"
	"flag"
	"net/http"
	"os"

	"github.com/ericflores108/spotify/config"
	"github.com/ericflores108/spotify/httpserver"
	"github.com/ericflores108/spotify/logger"
	"github.com/ericflores108/spotify/service"
)

func main() {
	logger.LogInfo("starting app")

	ctx := context.Background()

	appConfig := config.GetConfig(ctx)
	defer appConfig.SecretManagerClient.Close()
	defer appConfig.FirestoreClient.Close()

	// Define a flag for the redirect URL
	useLocalHost := flag.Bool("useLocalHost", false, "Use localhost as the redirect URL (default: production URL)")
	flag.Parse()

	// Determine the redirect URL
	redirectURL := "https://spotify-123259034538.us-west1.run.app/callback" // Default production URL
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
