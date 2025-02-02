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
	"strings"
	"sync"

	"cloud.google.com/go/firestore"
	"github.com/ericflores108/spotify/config"
	"github.com/ericflores108/spotify/db"
	"github.com/ericflores108/spotify/htmlpages"
	"github.com/ericflores108/spotify/logger"
	"github.com/ericflores108/spotify/sampled"
	"github.com/ericflores108/spotify/spotify"
)

type Service struct {
	SampledManager      *sampled.SampledManager
	Firestore           *firestore.Client
	SpotifyClientID     string
	SpotifyClientSecret string
	URL                 string
	StateKey            string
}

func (s *Service) GeneratePlaylistHandler(w http.ResponseWriter, ctx context.Context, albumID, userID, accessToken string, r *http.Request) {
	spotifyClient := &spotify.AuthClient{
		Client:      &http.Client{},
		AccessToken: accessToken,
	}

	album, err := spotifyClient.GetAlbum(albumID)
	if err != nil {
		logger.LogError("Failed to get album: %v", err)
		htmlpages.RenderErrorPage(w, fmt.Sprintf("Failed to get album: %v", err.Error()))
		return
	}

	if album == nil {
		logger.LogError("Failed to get album ID: %s", albumID)
		htmlpages.RenderErrorPage(w, "Failed to get album.")
		return
	}

	albumTracks, err := spotifyClient.GetAlbumTracks(albumID)
	if err != nil {
		logger.LogError("Failed to get album tracks: %v", err)
		htmlpages.RenderErrorPage(w, fmt.Sprintf("Failed to get album tracks: %v", err.Error()))
		return
	}

	if albumTracks == nil {
		logger.LogError("Failed to get album tracks for ID: %s", albumID)
		htmlpages.RenderErrorPage(w, "Failed to get album tracks")
		return
	}

	var (
		spotifyTracks = make([]string, len(albumTracks.Tracks.Items)*2)
		mu            sync.Mutex
		wg            sync.WaitGroup
	)

	for index, track := range albumTracks.Tracks.Items {

		var artist string

		if len(track.Artists) > 0 {
			artist = track.Artists[0].Name
		} else {
			logger.LogDebug("Unknown artist for track %s", track.Name)
		}

		mu.Lock()
		spotifyTracks[index*2] = track.URI
		mu.Unlock()

		// this can be genius, openai, etc. order matters when set in main
		wg.Add(1)
		go func(index int, trackName, artist string) {
			defer wg.Done()
			for _, source := range s.SampledManager.Sources {
				spotifyTrack, err := source.GetSample(ctx, track.Name, artist)
				if err != nil {
					logger.LogError("Error getting %s by %s sample: %v", track.Name, artist, err)
					continue
				}

				if spotifyTrack == nil {
					continue
				}

				mu.Lock()
				spotifyTracks[index*2+1] = spotifyTrack.URI
				mu.Unlock()

				break
			}
		}(index, track.Name, artist)
	}

	wg.Wait()

	filteredTracksMap := make(map[string]struct{})
	for _, uri := range spotifyTracks {
		if uri != "" { // Only keep valid URIs
			filteredTracksMap[uri] = struct{}{}
		}
	}

	filteredTracks := make([]string, 0, len(filteredTracksMap))
	for uri := range filteredTracksMap {
		filteredTracks = append(filteredTracks, uri)
	}

	if len(filteredTracks) == 0 {
		logger.LogError("Failed to retrieve tracks")
		htmlpages.RenderErrorPage(w, "Failed to retrieve tracks")
		return
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
		htmlpages.RenderErrorPage(w, fmt.Sprintf("Failed to create playlist: %v", err.Error()))
		return
	}

	err = spotifyClient.AddToPlaylist(userPlaylist.ID, filteredTracks, nil)
	if err != nil {
		logger.LogError("Failed to add tracks to playlist: %v", err)
		htmlpages.RenderErrorPage(w, fmt.Sprintf("Failed to add tracks to playlist: %v", err.Error()))
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
