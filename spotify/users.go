package spotify

import (
	"encoding/json"
	"fmt"
	"io"
)

// GetTopItems retrieves the user's top artists or tracks from Spotify, based on the specified TopType.
// It makes an authenticated request to the "me/top/{type}" endpoint and returns a pointer to TopResponse or an error.
func (c *AuthClient) GetTopItems(top TopType) (*TopResponse, error) {
	resp, err := c.Get("/me/top/" + string(top))
	if err != nil {
		return nil, fmt.Errorf("failed to get response: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var topResponse TopResponse
	if err := json.Unmarshal(body, &topResponse); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Return a pointer to artistsResponse and a nil error
	return &topResponse, nil
}

func (c *AuthClient) GetUser() (*MeResponse, error) {
	resp, err := c.Get("/me")
	if err != nil {
		return nil, fmt.Errorf("failed to get response: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var me MeResponse
	if err := json.Unmarshal(body, &me); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Return a pointer to artistsResponse and a nil error
	return &me, nil
}
