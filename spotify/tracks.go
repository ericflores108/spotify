package spotify

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"
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

func (c *AuthClient) GetTrackURI(trackName, artistName string) (string, error) {
	query := url.Values{}
	query.Set("q", fmt.Sprintf("track:%s artist:%s", trackName, artistName))
	query.Set("type", "track")
	query.Set("limit", "1")

	endpoint := fmt.Sprintf("/search?%s", query.Encode())

	resp, err := c.Get(endpoint)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	var searchResponse SearchResponse
	if err := json.Unmarshal(body, &searchResponse); err != nil {
		return "", fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Check if any tracks were found
	if len(searchResponse.Tracks.Items) == 0 {
		return "", fmt.Errorf("no tracks found for %s by %s", trackName, artistName)
	}

	// Return the URI of the first track found
	return searchResponse.Tracks.Items[0].URI, nil
}
