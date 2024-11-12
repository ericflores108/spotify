package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"slices"

	"cloud.google.com/go/bigquery"
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

	// Define HTTP route handlers
	http.HandleFunc("/", helloHandler)
	http.HandleFunc("/recommendations", func(w http.ResponseWriter, r *http.Request) {
		storeRecommendationsHandler(w, ctx, log, clientID, clientSecret, firestoreClient)
	})
	http.HandleFunc("/topTracks", func(w http.ResponseWriter, r *http.Request) {
		storeTracksHandler(w, ctx, log, clientID, clientSecret, firestoreClient)
	})

	// Determine port for HTTP service
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.InfoLogger.Printf("defaulting to port %s", port)
	}

	// Start HTTP server
	log.InfoLogger.Printf("listening on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.InfoLogger.Fatal(err)
	}
}

// helloHandler responds with a simple "Hello" message
func helloHandler(w http.ResponseWriter, r *http.Request) {
	name := os.Getenv("NAME")
	if name == "" {
		name = "Spotify"
	}
	fmt.Fprintf(w, "Hello %s!\n", name)
}

// storeRecommendationsHandler retrieves recommendations for each user and stores them in BigQuery
func storeRecommendationsHandler(w http.ResponseWriter, ctx context.Context, log *logger.Logger, clientID, clientSecret string, firestoreClient *firestore.Client) {
	// Retrieve all users from the SpotifyUser collection
	users, err := db.GetAllUsers(ctx, firestoreClient)
	if err != nil {
		log.ErrorLogger.Fatalf("failed to get users: %v", err)
		http.Error(w, "Failed to get users", http.StatusInternalServerError)
		return
	}

	for user := range slices.Values(users) {
		log.InfoLogger.Printf("User ID: %s, Display Name: %s", user.ID, user.DisplayName)

		accessToken, err := auth.GetUserAccessToken(user.RefreshToken, clientID, clientSecret)
		if err != nil {
			log.ErrorLogger.Fatalf("failed to access token: %v", err)
			http.Error(w, "Failed to get access token", http.StatusInternalServerError)
			return
		}

		spotifyClient := &spotify.AuthClient{
			Client:      &http.Client{},
			AccessToken: accessToken,
		}

		recommendations, err := spotifyClient.Recommend()
		if err != nil {
			log.ErrorLogger.Fatalf("failed to get recommendations: %v", err)
			http.Error(w, "Failed to get recommendations", http.StatusInternalServerError)
			return
		}

		bqClient, err := bigquery.NewClient(ctx, config.SpotifyProjectID)
		if err != nil {
			log.ErrorLogger.Fatalf("failed to create BigQuery client: %v", err)
			http.Error(w, "Failed to create BigQuery client", http.StatusInternalServerError)
			return
		}
		defer bqClient.Close()

		err = db.StoreRecommendations(ctx, bqClient, user.ID, recommendations)
		if err != nil {
			log.ErrorLogger.Fatalf("failed to store recommendations: %v", err)
			http.Error(w, "Failed to store recommendations", http.StatusInternalServerError)
			return
		}
	}

	log.Info("Recommendations stored successfully.")
	fmt.Fprintf(w, "Recommendations stored successfully.")
}

func storeTracksHandler(w http.ResponseWriter, ctx context.Context, log *logger.Logger, clientID, clientSecret string, firestoreClient *firestore.Client) {
	// Retrieve all users from the SpotifyUser collection
	users, err := db.GetAllUsers(ctx, firestoreClient)
	if err != nil {
		log.ErrorLogger.Fatalf("failed to get users: %v", err)
		http.Error(w, "Failed to get users", http.StatusInternalServerError)
		return
	}

	for user := range slices.Values(users) {
		log.InfoLogger.Printf("User ID: %s, Display Name: %s", user.ID, user.DisplayName)

		accessToken, err := auth.GetUserAccessToken(user.RefreshToken, clientID, clientSecret)
		if err != nil {
			log.ErrorLogger.Fatalf("failed to access token: %v", err)
			http.Error(w, "Failed to get access token", http.StatusInternalServerError)
			return
		}

		spotifyClient := &spotify.AuthClient{
			Client:      &http.Client{},
			AccessToken: accessToken,
		}

		topTracks, err := spotifyClient.TopTracks()
		if err != nil {
			log.ErrorLogger.Fatalf("failed to get recommendations: %v", err)
			http.Error(w, "Failed to get recommendations", http.StatusInternalServerError)
			return
		}

		bqClient, err := bigquery.NewClient(ctx, config.SpotifyProjectID)
		if err != nil {
			log.ErrorLogger.Fatalf("failed to create BigQuery client: %v", err)
			http.Error(w, "Failed to create BigQuery client", http.StatusInternalServerError)
			return
		}
		defer bqClient.Close()

		err = db.StoreTopTracks(ctx, bqClient, user.ID, user.DisplayName, topTracks)
		if err != nil {
			log.ErrorLogger.Fatalf("failed to store top tracks: %v", err)
			http.Error(w, "Failed to store top tracks", http.StatusInternalServerError)
			return
		}
	}

	log.Info("Tracks stored successfully.")
	fmt.Fprintf(w, "Tracks stored successfully.")
}
