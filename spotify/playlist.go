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
