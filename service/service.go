package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"cloud.google.com/go/firestore"
	"github.com/ericflores108/spotify/ai"
	"github.com/ericflores108/spotify/config"
	"github.com/ericflores108/spotify/db"
	"github.com/ericflores108/spotify/genius"
	"github.com/ericflores108/spotify/htmlpages"
	"github.com/ericflores108/spotify/logger"
	"github.com/ericflores108/spotify/sampled"
	"github.com/ericflores108/spotify/spotify"
	"github.com/openai/openai-go"
)

type Service struct {
	SpotifyClientID     string
	SpotifyClientSecret string
	Firestore           *firestore.Client
	AI                  *openai.Client
	URL                 string
	StateKey            string
	GeniusClient        *genius.GeniusClient
}

// Initialize dependencies
func NewService(spotifyClientID, spotifyClientSecret, URL, stateKey string, firestoreClient *firestore.Client, aiClient *openai.Client, geniusClient *genius.GeniusClient) *Service {
	return &Service{
		SpotifyClientID:     spotifyClientID,
		SpotifyClientSecret: spotifyClientSecret,
		Firestore:           firestoreClient,
		AI:                  aiClient,
		URL:                 URL,
		StateKey:            stateKey,
		GeniusClient:        geniusClient,
	}
}

func (s *Service) GeneratePlaylistHandler(w http.ResponseWriter, ctx context.Context, albumID, userID, accessToken string, r *http.Request) {
	spotifyClient := &spotify.AuthClient{
		Client:      &http.Client{},
		AccessToken: accessToken,
	}

	album, err := spotifyClient.GetAlbum(albumID)
	if err != nil {
		logger.LogError("Failed to get album: %v", err)
		http.Error(w, "Failed to get album", http.StatusInternalServerError)
		return
	}

	albumTracks, err := spotifyClient.GetAlbumTracks(albumID)
	if err != nil {
		logger.LogError("Failed to get album tracks: %v", err)
		http.Error(w, "Failed to get album tracks", http.StatusInternalServerError)
		return
	}

	ai := &ai.AIClient{
		Client: s.AI,
	}

	var excludedTracks []string
	var mu sync.Mutex // Protect shared state
	var wg sync.WaitGroup
	results := make(chan struct {
		index     int
		trackURIs []string
	}, len(albumTracks.Tracks.Items))

	for i, track := range albumTracks.Tracks.Items {
		wg.Add(1)

		go func(index int, track spotify.TrackDetails) {
			defer wg.Done()

			var trackURIsForTrack []string
			var artist string

			if len(track.Artists) > 0 {
				artist = track.Artists[0].Name
			} else {
				logger.LogDebug("Unknown artist for track %s", track.Name)
			}

			logger.LogDebug("Starting search for %s by %s", track.Name, artist)

			// Add to excluded tracks with mutex
			mu.Lock()
			trackURIsForTrack = append(trackURIsForTrack, track.URI)
			excludedTracks = append(excludedTracks, fmt.Sprintf("%s by %s", track.Name, artist))
			mu.Unlock()

			var sampledTrack *sampled.SampledTrack

			// Find using Genius
			geniusSearch, err := s.GeniusClient.Search(track.Name, artist)
			if err == nil && len(geniusSearch.Response.Hits) > 0 {
				geniusTrack, err := s.GeniusClient.Songs(strconv.Itoa(geniusSearch.Response.Hits[0].Result.ID))
				if err == nil && len(geniusTrack.Response.Song.SongRelationships) > 0 {
					for _, relation := range geniusTrack.Response.Song.SongRelationships {
						if relation.RelationshipType == "samples" && len(relation.Songs) > 0 {
							sampledTrack = &sampled.SampledTrack{
								Artist: relation.Songs[0].Artist,
								Name:   relation.Songs[0].Title,
							}
							break
						}
					}
				}
			} else {
				logger.LogError("Error occurred at geniusSearch: %v", err)
			}

			// Find using AI if Genius doesn't have results
			if sampledTrack == nil {
				sampledTrack, err = ai.FindTrackSamples(ctx, track.Name, artist, excludedTracks)
				if err != nil {
					logger.LogError("Error occurred at sampledTrack: %v", err)
					return
				}

				if sampledTrack == nil {
					logger.LogError("No sampledTrack found for userID - %s: %v", userID, err)
					return
				}
			}

			// Get Spotify URI
			trackURI, err := spotifyClient.GetTrackURI(sampledTrack.Name, sampledTrack.Artist)
			if err != nil {
				logger.LogError("Error occurred at trackURI: %v", err)
				return
			}

			if trackURI == "" {
				logger.LogDebug("No trackURI found for - TRACK - %s - ARTIST - %s: %v", sampledTrack.Name, sampledTrack.Artist, err)
				return
			}

			trackURIsForTrack = append(trackURIsForTrack, trackURI)

			mu.Lock()
			excludedTracks = append(excludedTracks, fmt.Sprintf("%s by %s", sampledTrack.Name, sampledTrack.Artist))
			mu.Unlock()

			results <- struct {
				index     int
				trackURIs []string
			}{index, trackURIsForTrack}
		}(i, track)
	}

	// Close results channel once all goroutines finish
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results in order
	orderedResults := make([][]string, len(albumTracks.Tracks.Items))
	for result := range results {
		orderedResults[result.index] = result.trackURIs
	}

	// Flatten results with uniqueness check
	uriSet := make(map[string]struct{})
	var trackURIs []string
	for _, uris := range orderedResults {
		for _, uri := range uris {
			if _, exists := uriSet[uri]; !exists {
				uriSet[uri] = struct{}{}           // Mark URI as seen
				trackURIs = append(trackURIs, uri) // Add unique URI
			}
		}
	}

	// Create Spotify playlist
	playlist := spotify.NewPlaylist{
		Name:        fmt.Sprintf("Titled - Inspired Songs from %s", album.Name),
		Description: "Generated playlist from Titled.",
		Public:      true,
	}

	userPlaylist, err := spotifyClient.CreatePlaylist(userID, playlist)
	if err != nil {
		logger.LogError("Failed to create playlist: %v", err)
		http.Error(w, "Failed to create playlist", http.StatusInternalServerError)
		return
	}

	err = spotifyClient.AddToPlaylist(userPlaylist.ID, trackURIs, nil)
	if err != nil {
		logger.LogError("Failed to add tracks to playlist: %v", err)
		http.Error(w, "Failed to add tracks to playlist", http.StatusInternalServerError)
		return
	}

	logger.LogInfo("Playlist created. URI: %s, ID: %s", userPlaylist.URI, userPlaylist.ID)

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, htmlpages.Playlist, userPlaylist.ExternalURLs.Spotify, userPlaylist.ID)
}

func (s *Service) CallbackHandler(w http.ResponseWriter, ctx context.Context, r *http.Request) {
	state := r.URL.Query().Get("state")
	code := r.URL.Query().Get("code")

	cookie, err := r.Cookie(s.StateKey)
	if err != nil || cookie.Value != state {
		http.Error(w, "State mismatch", http.StatusBadRequest)
		return
	}

	// Exchange code for tokens
	tokenURL := "https://accounts.spotify.com/api/token"
	data := fmt.Sprintf("grant_type=authorization_code&code=%s&redirect_uri=%s", code, s.URL+"/callback")
	req, _ := http.NewRequest("POST", tokenURL, strings.NewReader(data))
	req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(s.SpotifyClientID+":"+s.SpotifyClientSecret)))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "Failed to get tokens", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		http.Error(w, "Failed to get tokens from Spotify", http.StatusUnauthorized)
		return
	}

	var tokenResponse struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
		logger.LogError("Failed to decode token response: %v", err)
		http.Error(w, "Failed to decode token response", http.StatusInternalServerError)
		return
	}

	spotifyClient := &spotify.AuthClient{
		Client:      &http.Client{},
		AccessToken: tokenResponse.AccessToken,
	}

	spotifyUser, err := spotifyClient.GetUser()
	if err != nil {
		logger.LogError("Failed to get user from Spotify: %v", err)
		http.Error(w, "Failed to get user from Spotify", http.StatusUnauthorized)
		return
	}

	user := db.User{
		ID:           spotifyUser.UserID,
		DisplayName:  spotifyUser.DisplayName,
		AccessToken:  tokenResponse.AccessToken,
		RefreshToken: tokenResponse.RefreshToken,
	}

	docID, err := db.CreateUser(ctx, s.Firestore, user)
	if err != nil {
		logger.LogError("Failed to create Titled user: %v", err)
		http.Error(w, "Failed to create Titled user", http.StatusUnauthorized)
		return
	}
	logger.LogDebug("DOC ID: %s", docID)

	tmpl := template.Must(template.New("form").Parse(htmlpages.GeneratePlaylist))

	formData := struct {
		UserID      string
		AlbumURL    string
		AccessToken string
	}{
		UserID:      spotifyUser.UserID,
		AccessToken: tokenResponse.AccessToken,
		AlbumURL:    "", // Placeholder for album name input
	}

	// Set the content type for the response
	w.Header().Set("Content-Type", "text/html")

	// Render the template
	if err := tmpl.Execute(w, formData); err != nil {
		logger.LogError("Failed to render template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func generateRandomString(length int) string {
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		log.Fatal(err)
	}
	return base64.URLEncoding.EncodeToString(b)[:length]
}

func (s *Service) LoginHandler(w http.ResponseWriter, r *http.Request) {
	state := generateRandomString(16)
	http.SetCookie(w, &http.Cookie{
		Name:  s.StateKey,
		Value: state,
		Path:  "/",
	})
	authURL := fmt.Sprintf("https://accounts.spotify.com/authorize?response_type=code&client_id=%s&scope=%s&redirect_uri=%s&state=%s",
		s.SpotifyClientID, config.SpotifyScope, s.URL+"/callback", state)
	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}
