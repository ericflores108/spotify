package spotify

import (
	"encoding/json"
	"fmt"
	"io"
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
	payload := map[string]interface{}{
		"uris": uris,
	}

	if position != nil {
		payload["position"] = *position
	}

	endpoint := fmt.Sprintf("/playlists/%s/tracks", playlistID)
	resp, err := c.Post(endpoint, payload)
	if err != nil {
		return fmt.Errorf("failed to get response: %w", err)
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
	endpoint := fmt.Sprintf("/users/%s/playlists", userID)

	resp, err := c.Get(endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to get response: %w", err)
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
	endpoint := fmt.Sprintf("/playlists/%s/tracks", playlistID)

	resp, err := c.Get(endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to get response: %w", err)
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
