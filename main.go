package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"github.com/ericflores108/spotify/auth"
	"github.com/ericflores108/spotify/config"
	"github.com/ericflores108/spotify/logger"
)

func main() {
	log := logger.NewLogger()
	log.Info("starting app")

	// Initialize context and Secret Manager client
	ctx := context.Background()
	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		log.Error("failed to create secret manager client")
	}
	defer client.Close()

	// Retrieve the ClientID and ClientSecret secrets
	clientID, err := auth.GetSecret(ctx, client, config.SpotifyProjectID, config.SpotifyClientID)
	if err != nil {
		log.InfoLogger.Fatalf("Secret Error: %v", err)
		return
	}

	clientSecret, err := auth.GetSecret(ctx, client, config.SpotifyProjectID, config.SpotifySecretID)
	if err != nil {
		log.InfoLogger.Fatalf("Secret Error: %v", err)
		return
	}

	// Populate the Config struct with the retrieved secrets
	config := auth.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
	}

	_, err = auth.GetSpotifyToken(config)

	if err != nil {
		log.Error("error getting token")
		return
	}

	log.Info("access token created")

	log.Info("starting server...")
	http.HandleFunc("/", handler)

	// Determine port for HTTP service.
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.InfoLogger.Printf("defaulting to port %s", port)
	}

	// Start HTTP server.
	log.InfoLogger.Printf("listening on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.InfoLogger.Fatal(err)
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	name := os.Getenv("NAME")
	if name == "" {
		name = "World"
	}
	fmt.Fprintf(w, "Hello %s!\n", name)
}
