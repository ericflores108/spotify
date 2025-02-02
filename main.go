package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/ericflores108/spotify/ai"
	"github.com/ericflores108/spotify/config"
	"github.com/ericflores108/spotify/httpserver"
	"github.com/ericflores108/spotify/logger"
	"github.com/ericflores108/spotify/sampled"
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

	// Define a flag for the URL
	useLocalHost := flag.Bool("useLocalHost", false, "Use localhost as the URL (default: production URL)")
	flag.Parse()

	// Determine the URL
	titledURL := config.ProductionURL
	if *useLocalHost {
		titledURL = config.DevURL
	}

	aiClient := &ai.AIClient{
		Client: appConfig.OpenAIClient,
	}
	aiService := &sampled.AIService{
		Spotify: appConfig.SpotifyClient,
		AI:      aiClient,
	}

	geniusService := &sampled.GeniusService{
		Spotify: appConfig.SpotifyClient,
		Genius:  appConfig.GeniusClient,
	}

	sampledManager := sampled.NewSampledManager(geniusService, aiService)

	// Initialize the service
	svc := &service.Service{
		SampledManager:      sampledManager,
		Firestore:           appConfig.FirestoreClient,
		SpotifyClientID:     appConfig.ClientID,
		SpotifyClientSecret: appConfig.ClientSecret,
		URL:                 titledURL,
		StateKey:            config.StateKey,
	}

	// Initialize the server and register routes
	srv := httpserver.NewServer(svc)
	mux := srv.RegisterRoutes(ctx)

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
