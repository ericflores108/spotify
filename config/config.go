package config

import (
	"context"
	"log"
	"sync"

	"cloud.google.com/go/firestore"
	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"github.com/ericflores108/spotify/auth"
	"github.com/ericflores108/spotify/logger"
)

type AppConfig struct {
	ClientID            string
	ClientSecret        string
	SecretManagerClient *secretmanager.Client
	FirestoreClient     *firestore.Client
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
		clientID, err := auth.GetSecret(ctx, secretManagerClient, SpotifyProjectID, SpotifyClientID)
		if err != nil {
			logger.LogError("failed to retrieve ClientID secret: %v", err)
			log.Fatal(err)
		}

		clientSecret, err := auth.GetSecret(ctx, secretManagerClient, SpotifyProjectID, SpotifySecretID)
		if err != nil {
			logger.LogError("failed to retrieve ClientSecret secret: %v", err)
			log.Fatal(err)
		}

		// Initialize Firestore client
		firestoreClient, err := firestore.NewClient(ctx, SpotifyProjectID)
		if err != nil {
			logger.LogError("failed to create Firestore client: %v", err)
			log.Fatal(err)
		}

		// Assign to singleton instance
		instance = &AppConfig{
			ClientID:            clientID,
			ClientSecret:        clientSecret,
			SecretManagerClient: secretManagerClient,
			FirestoreClient:     firestoreClient,
		}

		logger.LogInfo("Configuration initialized successfully.")
	})
	return instance
}
