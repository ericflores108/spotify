package main

import (
	"context"
	"net/http"
	"os"

	"github.com/ericflores108/spotify/api"
	"github.com/ericflores108/spotify/config"
	"github.com/ericflores108/spotify/logger"
	"github.com/ericflores108/spotify/service"
)

func main() {
	logger.LogInfo("starting app")

	ctx := context.Background()

	appConfig := config.GetConfig(ctx)
	defer appConfig.SecretManagerClient.Close()
	defer appConfig.FirestoreClient.Close()

	// Initialize the service
	svc := service.NewService(
		appConfig.ClientID,
		appConfig.ClientSecret,
		appConfig.FirestoreClient,
		appConfig.OpenAIClient,
	)

	// Initialize the server and register routes
	srv := api.NewServer(ctx, svc)
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
