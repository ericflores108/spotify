package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"cloud.google.com/go/firestore"
	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"github.com/ericflores108/spotify/auth"
	"github.com/ericflores108/spotify/config"
	"github.com/ericflores108/spotify/db"
	"github.com/ericflores108/spotify/logger"
	"github.com/ericflores108/spotify/spotify"
)

func main() {
	log := logger.NewLogger()
	log.Info("starting app")

	// Initialize context and Secret Manager client
	ctx := context.Background()
	secretManagerClient, err := secretmanager.NewClient(ctx)
	if err != nil {
		log.Error("failed to create secret manager client")
	}
	defer secretManagerClient.Close()

	// Retrieve the ClientID and ClientSecret secrets
	clientID, err := auth.GetSecret(ctx, secretManagerClient, config.SpotifyProjectID, config.SpotifyClientID)
	if err != nil {
		log.InfoLogger.Fatalf("Secret Error: %v", err)
		return
	}

	clientSecret, err := auth.GetSecret(ctx, secretManagerClient, config.SpotifyProjectID, config.SpotifySecretID)
	if err != nil {
		log.InfoLogger.Fatalf("Secret Error: %v", err)
		return
	}

	firestoreClient, err := firestore.NewClient(ctx, config.SpotifyProjectID)
	if err != nil {
		log.ErrorLogger.Fatalf("failed to create Firestore client: %v", err)
	}

	// Retrieve all users from the SpotifyUser collection
	users, err := db.GetAllUsers(ctx, firestoreClient)
	if err != nil {
		log.ErrorLogger.Fatalf("failed to get users: %v", err)
	}

	for _, user := range users {
		log.InfoLogger.Printf("User ID: %s, Display Name: %s", user.ID, user.DisplayName)

		accessToken, err := auth.GetUserAccessToken(user.RefreshToken, clientID, clientSecret)
		if err != nil {
			log.ErrorLogger.Fatalf("failed to access token: %v", err)
		}

		spotifyClient := *&spotify.AuthClient{
			Client:      &http.Client{},
			AccessToken: accessToken,
		}

		spotifyClient.GetTopItems("artists")
	}

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
