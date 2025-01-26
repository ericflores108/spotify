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
	"slices"
	"strings"

	"cloud.google.com/go/bigquery"
	"cloud.google.com/go/firestore"
	"github.com/ericflores108/spotify/ai"
	"github.com/ericflores108/spotify/auth"
	"github.com/ericflores108/spotify/config"
	"github.com/ericflores108/spotify/db"
	"github.com/ericflores108/spotify/logger"
	"github.com/ericflores108/spotify/spotify"
	"github.com/openai/openai-go"
)

type Service struct {
	ClientID     string
	ClientSecret string
	Firestore    *firestore.Client
	AI           *openai.Client
	RedirectURI  string
	StateKey     string
}

// Initialize dependencies
func NewService(clientID, clientSecret, redirectURI, stateKey string, firestoreClient *firestore.Client, aiClient *openai.Client) *Service {
	return &Service{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Firestore:    firestoreClient,
		AI:           aiClient,
		RedirectURI:  redirectURI,
		StateKey:     stateKey,
	}
}

func (s *Service) CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	var newUser db.User
	if err := json.NewDecoder(r.Body).Decode(&newUser); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if newUser.ID == "" || newUser.DisplayName == "" || newUser.AccessToken == "" || newUser.RefreshToken == "" {
		http.Error(w, "Missing required fields: ID, DisplayName, RefreshToken, or AccessToken", http.StatusBadRequest)
		return
	}

	query := s.Firestore.Collection("SpotifyUser").Where("id", "==", newUser.ID).Limit(1)
	iter := query.Documents(r.Context())
	defer iter.Stop()

	doc, err := iter.Next()
	if err == nil {
		_, err := doc.Ref.Set(r.Context(), newUser)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to update user: %v", err), http.StatusInternalServerError)
			return
		}

		response := map[string]string{"message": "User updated successfully", "documentID": doc.Ref.ID}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	docRef, _, err := s.Firestore.Collection("SpotifyUser").Add(r.Context(), newUser)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create user: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]string{"message": "User created successfully", "documentID": docRef.ID}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *Service) GeneratePlaylistHandler(w http.ResponseWriter, ctx context.Context, albumID, userID string, r *http.Request) {
	user, err := db.GetUserByID(ctx, s.Firestore, userID)
	if err != nil {
		logger.LogError("failed to get user: %v", err)
		http.Error(w, "Failed to get user", http.StatusInternalServerError)
		return
	}

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

	album, err := spotifyClient.GetAlbum(albumID)
	if err != nil {
		logger.LogError("failed to get album: %v", err)
		http.Error(w, "Failed to get album", http.StatusInternalServerError)
		return
	}

	me, err := spotifyClient.GetUser()
	if err != nil {
		logger.LogError("failed to get spotify user id %s: %v", user.ID, err)
		http.Error(w, "Failed to get spotify user id", http.StatusInternalServerError)
		return
	}

	albumTracks, err := spotifyClient.GetAlbumTracks(albumID)
	if err != nil {
		logger.LogError("Failed to get album tracks for user %s: %v", user.ID, err)
		http.Error(w, "Failed to get album tracks", http.StatusInternalServerError)
		return
	}

	ai := &ai.AIClient{
		Client: s.AI,
	}

	var trackUris []string
	for track := range slices.Values(albumTracks.Tracks.Items) {
		trackUris = append(trackUris, track.URI)

		var artist string

		if len(track.Artists) > 0 {
			artist = track.Artists[0].Name
		} else {
			artist = "Unknown Artist"
		}

		sampledTrack, err := ai.FindTrackSamples(ctx, track.Name, artist)
		if err != nil {
			logger.LogError("Failed to get tracks to playlist for user %s: %v", user.ID, err)
			continue
		}

		if sampledTrack == nil {
			logger.LogDebug("No valid sampled track found.")
			continue
		}

		logger.LogDebug("Found sampled track: Artist: %s, Name: %s", sampledTrack.Artist, sampledTrack.Name)

		trackUri, err := spotifyClient.GetTrackURI(sampledTrack.Name, sampledTrack.Artist)
		if err != nil {
			logger.LogError("Failed to get Spotify URI for sampled track '%s' by '%s': %v", sampledTrack.Name, sampledTrack.Artist, err)
			continue
		}

		if trackUri == "" {
			logger.LogDebug("No Spotify URI found for sampled track '%s' by '%s'.", sampledTrack.Name, sampledTrack.Artist)
			continue
		}

		logger.LogDebug("Found Spotify URI: %s for track '%s' by '%s'", trackUri, sampledTrack.Name, sampledTrack.Artist)

		trackUris = append(trackUris, trackUri)
	}

	playlist := spotify.NewPlaylist{
		Name:        fmt.Sprintf("Titled - Inspired Songs from %s", album.Name),
		Description: "Generated playlist from Titled.",
		Public:      true,
	}

	userPlaylist, err := spotifyClient.CreatePlaylist(me.UserID, playlist)
	if err != nil {
		logger.LogError("failed to create playlist %s: %v", user.ID, err)
		http.Error(w, "Failed to create playlist", http.StatusInternalServerError)
		return
	}

	err = spotifyClient.AddToPlaylist(userPlaylist.ID, trackUris, nil)
	if err != nil {
		logger.LogError("Failed to get add %v to playlist %s", trackUris, userPlaylist.ID)
		http.Error(w, "Failed to add tracks to playlist", http.StatusInternalServerError)
		return
	}

	logger.LogInfo("URI: %s - ID: %s", userPlaylist.URI, userPlaylist.ID)

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `
		<!DOCTYPE html>
		<html>
		<head>
			<title>Spotify Playlist</title>
		</head>
		<body>
			<p>Your Spotify playlist is ready! <a href="%s">Click here</a> to open it.</p>
		</body>
		</html>`, userPlaylist.ExternalURLs.Spotify)
}

func (s *Service) StoreTracksHandler(w http.ResponseWriter, ctx context.Context) {
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

func (s *Service) CreatePlaylistHandler(w http.ResponseWriter, ctx context.Context) {
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
	req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(s.ClientID+":"+s.ClientSecret)))
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
		<title>User Form</title>
		<script>
			// Add a script to handle form submission
			function showLoading(event) {
				// Prevent the default form submission
				event.preventDefault();

				// Show the loading message
				document.getElementById('loadingMessage').style.display = 'block';

				// Submit the form programmatically
				event.target.submit();
			}
		</script>
	</head>
	<body>
		<form action="/generatePlaylist" method="post" onsubmit="showLoading(event)">
			<input type="hidden" id="userID" name="userID" value="{{.UserID}}">
			
			<label for="albumURL">Insert Spotify Album Link:</label>
			<input type="text" id="albumURL" name="albumURL" value="{{.AlbumURL}}"><br><br>

			<button type="submit">Generate Spotify Playlist</button>
		</form>

		<!-- Loading message hidden by default -->
		<div id="loadingMessage" style="display:none; color:blue; margin-top:20px;">
			Please wait... Generating your Spotify playlist.
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
		s.ClientID, config.SpotifyScope, s.RedirectURI, state)
	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}
