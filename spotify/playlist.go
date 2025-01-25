package spotify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func (c *AuthClient) CreatePlaylist(userID string, playlist NewPlaylist) (*NewPlaylistResponse, error) {
	// Use the Post method with the playlist payload
	resp, err := c.Post(fmt.Sprintf("/users/%s/playlists", userID), playlist)
	if err != nil {
		return nil, fmt.Errorf("failed to get response: %w", err)
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Parse the JSON response
	var newPlaylistResponse NewPlaylistResponse
	if err := json.Unmarshal(body, &newPlaylistResponse); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Return the parsed response
	return &newPlaylistResponse, nil
}

func (c *AuthClient) AddToPlaylist(playlistID string, uris []string, position *int) error {
	// Define the payload structure
	payload := map[string]interface{}{
		"uris": uris,
	}

	// Add the position if provided
	if position != nil {
		payload["position"] = *position
	}

	// Convert the payload to JSON
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Create the POST request
	url := fmt.Sprintf("/playlists/%s/tracks", playlistID)
	req, err := http.NewRequest("POST", BaseURL+url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set the necessary headers
	req.Header.Set("Authorization", "Bearer "+c.AccessToken)
	req.Header.Set("Content-Type", "application/json")

	// Send the request
	resp, err := c.Client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Handle non-2xx responses
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("request to add tracks failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

func (c *AuthClient) GetUserPlaylists(userID string) ([]Playlist, error) {
	// Construct the URL for the request
	url := fmt.Sprintf("/users/%s/playlists", userID)

	// Create the GET request
	req, err := http.NewRequest("GET", BaseURL+url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set the Authorization header
	req.Header.Set("Authorization", "Bearer "+c.AccessToken)

	// Send the request
	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Handle non-2xx responses
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get playlists: status %d, response %s", resp.StatusCode, string(body))
	}

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Parse the JSON response
	var playlistsResponse PlaylistsResponse
	if err := json.Unmarshal(body, &playlistsResponse); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Return the list of playlists
	return playlistsResponse.Items, nil
}

func (c *AuthClient) GetPlaylistTracks(playlistID string) ([]Track, error) {
	// Construct the URL for the request
	url := fmt.Sprintf("/playlists/%s/tracks", playlistID)

	// Create the GET request
	req, err := http.NewRequest("GET", BaseURL+url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set the Authorization header
	req.Header.Set("Authorization", "Bearer "+c.AccessToken)

	// Send the request
	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Handle non-2xx responses
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get playlist tracks: status %d, response %s", resp.StatusCode, string(body))
	}

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Parse the JSON response
	var playlistTracksResponse PlaylistTracksResponse
	if err := json.Unmarshal(body, &playlistTracksResponse); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Extract track details
	var tracks []Track
	for _, item := range playlistTracksResponse.Items {
		track := item.Track
		artistNames := make([]string, len(track.Artists))
		for i, artist := range track.Artists {
			artistNames[i] = artist.Name
		}
		tracks = append(tracks, Track{
			ID:   track.ID,
			Name: track.Name,
			URI:  track.URI,
		})
	}

	return tracks, nil
}
