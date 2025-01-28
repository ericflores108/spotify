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
	"github.com/ericflores108/spotify/auth"
	"github.com/ericflores108/spotify/config"
	"github.com/ericflores108/spotify/db"
	"github.com/ericflores108/spotify/genius"
	"github.com/ericflores108/spotify/logger"
	"github.com/ericflores108/spotify/spotify"
	"github.com/openai/openai-go"
)

type Service struct {
	SpotifyClientID     string
	SpotifyClientSecret string
	Firestore           *firestore.Client
	AI                  *openai.Client
	RedirectURI         string
	StateKey            string
	GeniusClient        *genius.GeniusClient
}

// Initialize dependencies
func NewService(spotifyClientID, spotifyClientSecret, redirectURI, stateKey string, firestoreClient *firestore.Client, aiClient *openai.Client, geniusClient *genius.GeniusClient) *Service {
	return &Service{
		SpotifyClientID:     spotifyClientID,
		SpotifyClientSecret: spotifyClientSecret,
		Firestore:           firestoreClient,
		AI:                  aiClient,
		RedirectURI:         redirectURI,
		StateKey:            stateKey,
		GeniusClient:        geniusClient,
	}
}

func (s *Service) GeneratePlaylistHandler(w http.ResponseWriter, ctx context.Context, albumID, userID string, r *http.Request) {
	user, err := db.GetUserByID(ctx, s.Firestore, userID)
	if err != nil {
		logger.LogError("Failed to get user: %v", err)
		http.Error(w, "Failed to get user", http.StatusInternalServerError)
		return
	}

	logger.LogDebug("User ID: %s, Display Name: %s", user.ID, user.DisplayName)

	spotifyAccessToken, err := auth.GetUserAccessToken(user.RefreshToken, s.SpotifyClientID, s.SpotifyClientSecret)
	if err != nil {
		logger.LogError("Failed to access token: %v", err)
		http.Error(w, "Failed to get access token", http.StatusInternalServerError)
		return
	}

	spotifyClient := &spotify.AuthClient{
		Client:      &http.Client{},
		AccessToken: spotifyAccessToken,
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

	me, err := spotifyClient.GetUser()
	if err != nil {
		logger.LogError("Failed to get Spotify user: %v", err)
		http.Error(w, "Failed to get Spotify user", http.StatusInternalServerError)
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

			var sampledTrack *config.SampledTrack

			// Find using Genius
			geniusSearch, err := s.GeniusClient.Search(track.Name, artist)
			if err == nil && len(geniusSearch.Response.Hits) > 0 {
				geniusTrack, err := s.GeniusClient.Songs(strconv.Itoa(geniusSearch.Response.Hits[0].Result.ID))
				if err == nil && len(geniusTrack.Response.Song.SongRelationships) > 0 {
					for _, relation := range geniusTrack.Response.Song.SongRelationships {
						if relation.RelationshipType == "samples" && len(relation.Songs) > 0 {
							sampledTrack = &config.SampledTrack{
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
				logger.LogError("No trackURI found for - TRACK - %s - ARTIST - %s: %v", sampledTrack.Name, sampledTrack.Artist, err)
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

	// Flatten results
	var trackURIs []string
	for _, uris := range orderedResults {
		trackURIs = append(trackURIs, uris...)
	}

	// Create Spotify playlist
	playlist := spotify.NewPlaylist{
		Name:        fmt.Sprintf("Titled - Inspired Songs from %s", album.Name),
		Description: "Generated playlist from Titled.",
		Public:      true,
	}

	userPlaylist, err := spotifyClient.CreatePlaylist(me.UserID, playlist)
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
	fmt.Fprintf(w, `
	<!DOCTYPE html>
	<html>
	<head>
		<title>Titled - Playlist</title>
		<link rel="icon" href="/static/favicon.ico" type="image/x-icon">
		<link href="https://fonts.googleapis.com/css2?family=Raleway:wght@400;700&display=swap" rel="stylesheet">
		<meta name="viewport" content="width=device-width, initial-scale=1.0">
		<style>
			body {
				font-family: 'Raleway', Arial, sans-serif;
				margin: 0;
				padding: 0;
				background-color: #ffffff;
				color: #000000;
				display: flex;
				justify-content: center;
				align-items: center;
				height: 100vh;
				padding: 10px;
				overflow-x: hidden; /* Prevent horizontal scrolling */
			}
			.container {
				width: 100%%;
				max-width: 600px;
				background-color: #ffffff;
				border: 8px solid #000000;
				box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
				padding: 20px;
				display: flex;
				flex-direction: column;
				gap: 15px;
				border-radius: 10px;
				box-sizing: border-box; /* Include padding and border in width/height */
			}
			.container div {
				border: 3px solid #000000;
				padding: 15px;
				border-radius: 5px;
				box-sizing: border-box;
			}
			.container .red {
				background-color: #ff0000;
				text-align: center;
				color: #ffffff;
			}
			.container .white {
				background-color: #ffffff;
				text-align: center;
			}
			a {
				color: #ffffff;
				text-decoration: none;
				font-weight: bold;
				background-color: #000000;
				padding: 10px 20px;
				border-radius: 5px;
				display: inline-block;
			}
			a:hover {
				background-color: #555555;
			}
			iframe {
				border-radius: 12px;
				width: 100%%;
				height: 352px;
				border: 0;
			}
			/* Responsive Design */
			@media (max-width: 480px) {
				body {
					padding: 5px;
				}
				.container {
					border-width: 5px;
					padding: 15px;
				}
				.container div {
					border-width: 2px;
					padding: 10px;
				}
				a {
					padding: 8px 15px;
					font-size: 14px;
				}
				iframe {
					height: 280px;
				}
			}
		</style>
	</head>
	<body>
		<div class="container">
			<div class="red">
				<p>Your Spotify playlist is ready!</p>
				<a href="%s">Click here</a> to open it.
			</div>
			<div class="white">
				<iframe src="https://open.spotify.com/embed/playlist/%s?utm_source=generator" frameborder="0" allowfullscreen allow="autoplay; clipboard-write; encrypted-media; fullscreen; picture-in-picture" loading="lazy"></iframe>
			</div>
		</div>
	</body>
	</html>`, userPlaylist.ExternalURLs.Spotify, userPlaylist.ID)
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
	data := fmt.Sprintf("grant_type=authorization_code&code=%s&redirect_uri=%s", code, s.RedirectURI)
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

	tmpl := template.Must(template.New("form").Parse(`
	<!DOCTYPE html>
	<html>
	<head>
		<title>Titled - User Form</title>
		<link rel="icon" href="/static/favicon.ico" type="image/x-icon">
		<link href="https://fonts.googleapis.com/css2?family=Raleway:wght@400;700&display=swap" rel="stylesheet">
		<meta name="viewport" content="width=device-width, initial-scale=1.0">
		<style>
			body {
				font-family: 'Raleway', Arial, sans-serif;
				margin: 0;
				padding: 0;
				background-color: #ffffff;
				color: #000000;
				display: flex;
				justify-content: center;
				align-items: center;
				height: 100vh;
				padding: 10px;
			}
			.container {
				width: 100%;
				max-width: 600px;
				background-color: #ffffff;
				border: 8px solid #000000;
				box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
				padding: 20px;
				display: grid;
				grid-template-columns: 1fr;
				gap: 15px;
				border-radius: 10px;
			}
			.container div {
				border: 3px solid #000000;
				padding: 15px;
				border-radius: 5px;
			}
			.container .red {
				background-color: #ff0000;
				text-align: center;
				color: #ffffff;
			}
			.container .yellow {
				background-color: #ffff00;
				text-align: center;
			}
			.container .blue {
				background-color: #0000ff;
				text-align: center;
				color: #ffffff;
				display: none; /* Initially hidden */
				padding: 15px;
				border-radius: 5px;
			}
			form {
				display: flex;
				flex-direction: column;
				gap: 15px;
			}
			label {
				font-weight: bold;
				text-align: left;
			}
			input, button {
				width: 100%;
				padding: 12px;
				font-size: 16px;
				border: 2px solid #000000;
				border-radius: 5px;
				box-sizing: border-box;
			}
			button {
				background-color: #000000;
				color: #ffffff;
				cursor: pointer;
				font-weight: bold;
			}
			button:hover {
				background-color: #555555;
			}
			/* Responsive Design */
			@media (max-width: 480px) {
				body {
					padding: 5px;
				}
				.container {
					border-width: 5px;
					padding: 15px;
				}
				.container div {
					border-width: 2px;
					padding: 10px;
				}
				input, button {
					padding: 10px;
					font-size: 14px;
				}
			}
		</style>
		<script>
			// Add a script to handle form submission
			function showLoading(event) {
				// Prevent the default form submission
				event.preventDefault();

				// Show the blue box and loading message
				document.querySelector('.blue').style.display = 'block';

				// Allow form submission after the loading message is displayed
				setTimeout(() => {
					event.target.submit();
				}, 50);
			}
		</script>
	</head>
	<body>
		<div class="container">
			<div class="red">
				<h1>Generate Spotify Playlist</h1>
			</div>
			<div class="yellow">
				<form action="/generatePlaylist" method="post" onsubmit="showLoading(event)">
					<input type="hidden" id="userID" name="userID" value="{{.UserID}}">
					
					<label for="albumURL">Insert Spotify Album Link:</label>
					<input type="text" id="albumURL" name="albumURL" value="{{.AlbumURL}}">

					<button type="submit">Generate</button>
				</form>
			</div>
			<div class="blue">
				<p>Please wait... Generating your Spotify playlist.</p>
			</div>
		</div>
	</body>
	</html>`))

	formData := struct {
		UserID   string
		AlbumURL string
	}{
		UserID:   spotifyUser.UserID,
		AlbumURL: "", // Placeholder for album name input
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
		s.SpotifyClientID, config.SpotifyScope, s.RedirectURI, state)
	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}
