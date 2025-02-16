package config

import (
	"context"
	"log"
	"sync"

	"cloud.google.com/go/firestore"
	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"github.com/ericflores108/spotify/auth"
	"github.com/ericflores108/spotify/genius"
	"github.com/ericflores108/spotify/logger"
	"github.com/ericflores108/spotify/spotify"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

type AppConfig struct {
	ClientID            string
	ClientSecret        string
	SecretManagerClient *secretmanager.Client
	FirestoreClient     *firestore.Client
	OpenAIClient        *openai.Client
	GeniusClient        *genius.GeniusClient
	SpotifyClient       *spotify.AuthClient
}

var (
	instance *AppConfig
	once     sync.Once
)

// GetConfig initializes and returns a singleton AppConfig instance
func GetConfig(ctx context.Context) *AppConfig {
	once.Do(func() {

		// Initialize Secret Manager client
		secretManagerClient, err := secretmanager.NewClient(ctx)
		if err != nil {
			logger.LogError("failed to create secret manager client: %v", err)
			log.Fatal(err) // Exit on failure
		}

		// Retrieve secrets
		clientID, err := auth.GetSecret(ctx, secretManagerClient, GoogleProjectID, SpotifyClientID)
		if err != nil {
			logger.LogError("failed to retrieve SpotifyClientID secret: %v", err)
			log.Fatal(err)
		}

		clientSecret, err := auth.GetSecret(ctx, secretManagerClient, GoogleProjectID, SpotifySecretID)
		if err != nil {
			logger.LogError("failed to retrieve SpotifySecretID secret: %v", err)
			log.Fatal(err)
		}

		openAISecret, err := auth.GetSecret(ctx, secretManagerClient, GoogleProjectID, OpenAIApiKeyID)
		if err != nil {
			logger.LogError("failed to retrieve OpenAIApiKeyID secret: %v", err)
			log.Fatal(err)
		}

		geniusClientSecret, err := auth.GetSecret(ctx, secretManagerClient, GoogleProjectID, GeniusClientSecret)
		if err != nil {
			logger.LogError("failed to retrieve GeniusClientSecret secret: %v", err)
			log.Fatal(err)
		}

		geniusClientID, err := auth.GetSecret(ctx, secretManagerClient, GoogleProjectID, GeniusClientID)
		if err != nil {
			logger.LogError("failed to retrieve GeniusClientID secret: %v", err)
			log.Fatal(err)
		}

		geniusClient, err := genius.NewClient(geniusClientID, geniusClientSecret)
		if err != nil {
			logger.LogError("failed to retrieve geniusClient: %v", err)
			log.Fatal(err)
		}

		// Initialize OpenAI client
		aiClient := openai.NewClient(
			option.WithAPIKey(openAISecret),
		)

		// Initialize Firestore client
		firestoreClient, err := firestore.NewClient(ctx, GoogleProjectID)
		if err != nil {
			logger.LogError("failed to create Firestore client: %v", err)
			log.Fatal(err)
		}

		// Initialize Client Credentials Flow https://developer.spotify.com/documentation/web-api/tutorials/client-credentials-flow
		spotifyClient, err := spotify.NewSpotifyClient(clientID, clientSecret)
		if err != nil {
			logger.LogError("failed to create Spotify client: %v", err)
			log.Fatal(err)
		}

		// Assign to singleton instance
		instance = &AppConfig{
			ClientID:            clientID,
			ClientSecret:        clientSecret,
			SecretManagerClient: secretManagerClient,
			FirestoreClient:     firestoreClient,
			OpenAIClient:        aiClient,
			GeniusClient:        geniusClient,
			SpotifyClient:       spotifyClient,
		}

		logger.LogInfo("Configuration initialized successfully.")
	})
	return instance
}
