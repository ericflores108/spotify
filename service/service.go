package service

import (
	"context"
	"fmt"
	"net/http"
	"slices"

	"cloud.google.com/go/bigquery"
	"cloud.google.com/go/firestore"
	"github.com/ericflores108/spotify/auth"
	"github.com/ericflores108/spotify/config"
	"github.com/ericflores108/spotify/db"
	"github.com/ericflores108/spotify/logger"
	"github.com/ericflores108/spotify/spotify"
)

type Service struct {
	ClientID     string
	ClientSecret string
	Firestore    *firestore.Client
}

// Initialize dependencies
func NewService(clientID, clientSecret string, firestoreClient *firestore.Client) *Service {
	return &Service{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Firestore:    firestoreClient,
	}
}

func (s Service) StoreRecommendationsHandler(w http.ResponseWriter, ctx context.Context) {
	// Retrieve all users from the SpotifyUser collection
	users, err := db.GetAllUsers(ctx, s.Firestore)
	if err != nil {
		logger.LogError("failed to get users: %v", err)
		http.Error(w, "Failed to get users", http.StatusInternalServerError)
		return
	}

	for user := range slices.Values(users) {
		logger.LogDebug("User ID: %s, Display Name: %s", user.ID, user.DisplayName)

		accessToken, err := auth.GetUserAccessToken(user.RefreshToken, s.ClientID, s.ClientSecret)
		if err != nil {
			logger.LogError("failed to access token: %v", err)
			http.Error(w, "Failed to get access token", http.StatusInternalServerError)
			return
		}

		spotifyClient := &spotify.AuthClient{
			Client:      &http.Client{},
			AccessToken: accessToken,
		}

		recommendations, err := spotifyClient.Recommend()
		if err != nil {
			logger.LogError("failed to get recommendations: %v", err)
			http.Error(w, "Failed to get recommendations", http.StatusInternalServerError)
			return
		}

		bqClient, err := bigquery.NewClient(ctx, config.SpotifyProjectID)
		if err != nil {
			logger.LogError("failed to create BigQuery client: %v", err)
			http.Error(w, "Failed to create BigQuery client", http.StatusInternalServerError)
			return
		}
		defer bqClient.Close()

		err = db.StoreRecommendations(ctx, bqClient, user.ID, recommendations)
		if err != nil {
			logger.LogError("failed to store recommendations: %v", err)
			http.Error(w, "Failed to store recommendations", http.StatusInternalServerError)
			return
		}
	}

	logger.LogInfo("Recommendations stored successfully.")
	fmt.Fprintf(w, "Recommendations stored successfully.")
}

func (s Service) StoreTracksHandler(w http.ResponseWriter, ctx context.Context) {
	// Retrieve all users from the SpotifyUser collection
	users, err := db.GetAllUsers(ctx, s.Firestore)
	if err != nil {
		logger.LogError("failed to get users: %v", err)
		http.Error(w, "Failed to get users", http.StatusInternalServerError)
		return
	}

	for user := range slices.Values(users) {
		logger.LogDebug("User ID: %s, Display Name: %s", user.ID, user.DisplayName)

		accessToken, err := auth.GetUserAccessToken(user.RefreshToken, s.ClientID, s.ClientSecret)
		if err != nil {
			logger.LogError("failed to access token: %v", err)
			http.Error(w, "Failed to get access token", http.StatusInternalServerError)
			return
		}

		spotifyClient := &spotify.AuthClient{
			Client:      &http.Client{},
			AccessToken: accessToken,
			UserID:      user.ID,
		}

		topTracks, err := spotifyClient.TopTracks()
		if err != nil {
			logger.LogError("failed to get recommendations: %v", err)
			http.Error(w, "Failed to get recommendations", http.StatusInternalServerError)
			return
		}

		bqClient, err := bigquery.NewClient(ctx, config.SpotifyProjectID)
		if err != nil {
			logger.LogError("failed to create BigQuery client: %v", err)
			http.Error(w, "Failed to create BigQuery client", http.StatusInternalServerError)
			return
		}
		defer bqClient.Close()

		err = db.StoreTopTracks(ctx, bqClient, user.ID, user.DisplayName, topTracks)
		if err != nil {
			logger.LogError("failed to store top tracks: %v", err)
			http.Error(w, "Failed to store top tracks", http.StatusInternalServerError)
			return
		}
	}

	logger.LogInfo("Tracks stored successfully.")
	fmt.Fprintf(w, "Tracks stored successfully.")
}

func (s Service) CreatePlaylistHandler(w http.ResponseWriter, ctx context.Context) {
	// Retrieve all users from the SpotifyUser collection
	users, err := db.GetAllUsers(ctx, s.Firestore)
	if err != nil {
		logger.LogError("failed to get users: %v", err)
		http.Error(w, "Failed to get users", http.StatusInternalServerError)
		return
	}

	for user := range slices.Values(users) {
		logger.LogDebug("User ID: %s, Display Name: %s", user.ID, user.DisplayName)

		accessToken, err := auth.GetUserAccessToken(user.RefreshToken, s.ClientID, s.ClientSecret)
		if err != nil {
			logger.LogError("failed to access token: %v", err)
			continue // Skip to the next user
		}

		spotifyClient := &spotify.AuthClient{
			Client:      &http.Client{},
			AccessToken: accessToken,
			UserID:      user.DisplayName,
		}
		playlist := spotify.NewPlaylist{
			Name:        "golang playlist",
			Description: "cool one",
			Public:      false,
		}

		me, err := spotifyClient.GetUser()
		fmt.Println(me)

		if err != nil {
			logger.LogError("failed to create playlist for user %s: %v", user.ID, err)
			continue // Skip to the next user
		}

		resp, err := spotifyClient.CreatePlaylist(me.UserID, playlist)
		if err != nil {
			logger.LogError("failed to create playlist for user %s: %v", user.ID, err)
			continue // Skip to the next user
		}

		// Log the URI of the created playlist
		logger.LogInfo("Playlist created for user %s: URI: %s", user.ID, resp.URI)
		fmt.Fprintf(w, "Playlist created for user %s: URI: %s\n", user.ID, resp.URI)
	}

	logger.LogInfo("Processed playlist creation for all users.")
	fmt.Fprintln(w, "Finished creating playlists for all users.")
}

func (s Service) GetAlbumDetailsHandler(w http.ResponseWriter, ctx context.Context) {
	// Retrieve all users from the SpotifyUser collection
	users, err := db.GetAllUsers(ctx, s.Firestore)
	if err != nil {
		logger.LogError("Failed to get users: %v", err)
		http.Error(w, "Failed to get users", http.StatusInternalServerError)
		return
	}

	for user := range slices.Values(users) {
		logger.LogDebug("User ID: %s, Display Name: %s", user.ID, user.DisplayName)

		// Get user access token
		accessToken, err := auth.GetUserAccessToken(user.RefreshToken, s.ClientID, s.ClientSecret)
		if err != nil {
			logger.LogError("Failed to get access token for user %s: %v", user.ID, err)
			continue // Skip to the next user
		}

		spotifyClient := &spotify.AuthClient{
			Client:      &http.Client{},
			AccessToken: accessToken,
			UserID:      user.ID,
		}

		albumID := "0hvT3yIEysuuvkK73vgdcW?si=qJ0OwJpkT_2ubOm2XhisRg"

		album, err := spotifyClient.GetSimplifiedAlbumDetails(albumID)
		if err != nil {
			logger.LogError("Failed to get tracks to playlist for user %s: %v", user.ID, err)
			continue // Skip to the next user
		}

		// Log success
		logger.LogInfo("Details %v", album)
		fmt.Fprintf(w, "Details for album %v\n", album)
	}

	logger.LogInfo("Retrieved tracks.")
	fmt.Fprintln(w, "Finished retrieving album tracks.")
}

func (s Service) AddToPlaylistHandler(w http.ResponseWriter, ctx context.Context) {
	// Retrieve all users from the SpotifyUser collection
	users, err := db.GetAllUsers(ctx, s.Firestore)
	if err != nil {
		logger.LogError("Failed to get users: %v", err)
		http.Error(w, "Failed to get users", http.StatusInternalServerError)
		return
	}

	for user := range slices.Values(users) {
		logger.LogDebug("User ID: %s, Display Name: %s", user.ID, user.DisplayName)

		// Get user access token
		accessToken, err := auth.GetUserAccessToken(user.RefreshToken, s.ClientID, s.ClientSecret)
		if err != nil {
			logger.LogError("Failed to get access token for user %s: %v", user.ID, err)
			continue // Skip to the next user
		}

		spotifyClient := &spotify.AuthClient{
			Client:      &http.Client{},
			AccessToken: accessToken,
			UserID:      user.ID,
		}

		// Define the playlist ID and tracks to add
		playlistID := "644l2DeNJdITOkkMeDmfFx" // Replace with your target playlist ID
		uris := []string{
			"spotify:track:3HFBqhotJeEKHJzMEW31jZ",
			"spotify:track:49FA0CCwP0GmIVbPzBqjD4",
			"spotify:track:44KWbTVZev3SWdv1t5UoYE",
			"spotify:track:4WhYHtwrNzjloBMdLOeK4o",
			"spotify:track:6FTSWKjJlM1LGZsoIwlD90",
		}
		position := 0

		// Add tracks to the playlist
		err = spotifyClient.AddToPlaylist(playlistID, uris, &position)
		if err != nil {
			logger.LogError("Failed to add tracks to playlist for user %s: %v", user.ID, err)
			continue // Skip to the next user
		}

		// Log success
		logger.LogInfo("Tracks added to playlist for user %s", user.ID)
		fmt.Fprintf(w, "Tracks added to playlist for user %s\n", user.ID)
	}

	logger.LogInfo("Processed adding tracks to playlists for all users.")
	fmt.Fprintln(w, "Finished adding tracks to playlists for all users.")
}
