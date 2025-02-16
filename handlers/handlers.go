package handlers

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"slices"
	"strings"
	"sync"
	"time"

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

	// check if album has been processed in the last week
	playlistTracks, err := db.GetTracks(ctx, s.Firestore, albumID)
	if err != nil {
		logger.LogDebug("Error occurred at db.GetTracks(ctx, s.Firestore, albumID): %v", err)
	}

	if len(playlistTracks) == 0 {
		// find tracks
		logger.LogDebug("Tracks not cached")

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

		filteredTracksMap := make(map[string]bool)
		filteredPlaylist := make([]string, 0, len(spotifyTracks))

		for _, uri := range spotifyTracks {
			if uri == "" {
				continue
			}

			// If the URI is already in the slice, remove it
			if filteredTracksMap[uri] {
				// Find and remove previous occurrence
				for i, existingURI := range playlistTracks {
					if existingURI == uri {
						playlistTracks = slices.Delete(filteredPlaylist, i, i+1)
						break // Remove only the first occurrence
					}
				}
			}

			// Append latest occurrence and mark it as seen
			filteredPlaylist = append(filteredPlaylist, uri)
			filteredTracksMap[uri] = true
		}
		playlistTracks = filteredPlaylist

		err = db.SetTracks(ctx, s.Firestore, albumID, filteredPlaylist)
		if err != nil {
			logger.LogError("Failed to set tracks: %v", err)
		}
	}

	if len(playlistTracks) == 0 {
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

	err = spotifyClient.AddToPlaylist(userPlaylist.ID, playlistTracks, nil)
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

func (s *Service) exchangeCodeForToken(code string) (*spotify.TokenResponse, error) {
	// Exchange code for tokens
	tokenURL := "https://accounts.spotify.com/api/token"
	data := fmt.Sprintf("grant_type=authorization_code&code=%s&redirect_uri=%s", code, s.URL+"/callback")
	req, _ := http.NewRequest("POST", tokenURL, strings.NewReader(data))
	req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(s.SpotifyClientID+":"+s.SpotifyClientSecret)))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logger.LogError("Failed to get tokens: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.LogError("Failed to get tokens from Spotify.")
		return nil, fmt.Errorf("failed to get tokens from Spotify")
	}

	var tokenResponse spotify.TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
		logger.LogError("Failed to decode token response: %v", err)
		return nil, err
	}

	return &tokenResponse, nil
}

func (s *Service) CallbackHandler(w http.ResponseWriter, ctx context.Context, r *http.Request) {
	state := r.URL.Query().Get("state")
	code := r.URL.Query().Get("code")

	cookie, err := r.Cookie(s.StateKey)
	if err != nil || cookie.Value != state {
		logger.LogError("State mismatch error: %v", err)
		logger.LogDebug("Expected state: %s, Received state: %s", cookie.Value, state)
		http.Error(w, "State mismatch", http.StatusBadRequest)
		return
	}

	token, err := s.exchangeCodeForToken(code)
	if err != nil || cookie.Value != state {
		logger.LogError("s.exchangeCodeForToken(code) error: %v", err)
		http.Error(w, "Failed to exchange code for token", http.StatusBadRequest)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "spotify_token",
		Value:    token.AccessToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(time.Hour - time.Minute),
	})

	spotifyClient := &spotify.AuthClient{
		Client:      &http.Client{},
		AccessToken: token.AccessToken,
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
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
	}

	docID, err := db.CreateUser(ctx, s.Firestore, user)
	if err != nil {
		logger.LogError("Failed to create Titled user: %v", err)
		http.Error(w, "Failed to create Titled user", http.StatusUnauthorized)
		return
	}
	logger.LogDebug("DOC ID: %s", docID)

	http.SetCookie(w, &http.Cookie{
		Name:     "_id",
		Value:    spotifyUser.UserID,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(time.Duration(token.ExpiresIn) * time.Second),
	})

	http.Redirect(w, r, "/home", http.StatusSeeOther)
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
		Name:     s.StateKey,
		Value:    state,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(10 * time.Minute), // Expire in 10 mins
	})
	authURL := fmt.Sprintf("https://accounts.spotify.com/authorize?response_type=code&client_id=%s&scope=%s&redirect_uri=%s&state=%s",
		s.SpotifyClientID, config.SpotifyScope, s.URL+"/callback", state)
	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}

func (s *Service) HomePageHandler(w http.ResponseWriter, ctx context.Context, r *http.Request) {
	if r.URL.Path == "/spotify" {
		eflorty := "31h2tegtv6vy7gkjsndegyk6hzgq"
		user, err := db.GetUserByID(ctx, s.Firestore, eflorty)
		if err != nil {
			logger.LogError("Failed to get default user: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		formData := struct {
			UserID      string
			AlbumURL    string
			AccessToken string
		}{
			UserID:      eflorty,
			AccessToken: user.AccessToken,
			AlbumURL:    "",
		}

		tmpl := template.Must(template.New("form").Parse(htmlpages.GeneratePlaylist))

		w.Header().Set("Content-Type", "text/html")
		if err := tmpl.Execute(w, formData); err != nil {
			logger.LogError("Failed to render template: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	} else {
		userIDCookie, err := r.Cookie("_id")
		if err != nil {
			logger.LogError("Failed to get user ID cookie: %v", err)
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		accessTokenCookie, err := r.Cookie("spotify_token")
		if err != nil {
			logger.LogError("Failed to get access token cookie: %v", err)
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		formData := struct {
			UserID      string
			AlbumURL    string
			AccessToken string
		}{
			UserID:      userIDCookie.Value,
			AccessToken: accessTokenCookie.Value,
			AlbumURL:    "",
		}

		tmpl := template.Must(template.New("form").Parse(htmlpages.GeneratePlaylist))

		w.Header().Set("Content-Type", "text/html")
		if err := tmpl.Execute(w, formData); err != nil {
			logger.LogError("Failed to render template: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}
}
