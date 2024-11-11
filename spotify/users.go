package spotify

import (
	"encoding/json"
	"fmt"
	"io"
)

// GetTopItems retrieves the user's top artists or tracks from Spotify, based on the specified TopType.
// It makes an authenticated request to the "me/top/{type}" endpoint and returns a pointer to ArtistsResponse or an error.
func (a *AuthClient) GetTopItems(top TopType) (*ArtistsResponse, error) {
	resp, err := a.Get("/me/top/" + string(top))
	if err != nil {
		return nil, fmt.Errorf("failed to get response: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var artistsResponse ArtistsResponse
	if err := json.Unmarshal(body, &artistsResponse); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Return a pointer to artistsResponse and a nil error
	return &artistsResponse, nil
}
