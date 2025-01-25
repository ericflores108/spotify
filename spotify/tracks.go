package spotify

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// GetRecommendations retrieves recommendations based on seed artists, genres, and tracks.
// It makes an authenticated request to the "recommendations" endpoint and returns a pointer to RecommendationsResponse or an error.
func (c *AuthClient) GetRecommendations(seedArtists, seedGenres, seedTracks string) (*RecommendationsResponse, error) {
	// Start building the endpoint with the base URL
	endpoint := "/recommendations?"

	// Append query parameters conditionally based on which are non-empty
	if seedArtists != "" {
		endpoint += fmt.Sprintf("seed_artists=%s&", seedArtists)
	}
	if seedGenres != "" {
		endpoint += fmt.Sprintf("seed_genres=%s&", seedGenres)
	}
	if seedTracks != "" {
		endpoint += fmt.Sprintf("seed_tracks=%s&", seedTracks)
	}

	// Remove the trailing "&" if any parameters were added
	endpoint = strings.TrimSuffix(endpoint, "&")

	// Make the GET request using the AuthClient's Get method
	resp, err := c.Get(endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to get response: %w", err)
	}
	defer resp.Body.Close()

	// Read and parse the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var recommendationsResponse RecommendationsResponse

	// Unmarshal the JSON data into the struct
	if err := json.Unmarshal(body, &recommendationsResponse); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Return the parsed response and nil error on success
	return &recommendationsResponse, nil
}
