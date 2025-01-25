package spotify

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func (c *AuthClient) GetAlbumDetails(albumID string) (*AlbumResponse, error) {
	// Construct the URL for the album endpoint
	url := fmt.Sprintf("/albums/%s/tracks", albumID)

	// Create the GET request
	req, err := http.NewRequest("GET", BaseURL+url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add Authorization header
	req.Header.Set("Authorization", "Bearer "+c.AccessToken)

	// Send the request
	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read and parse the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var albumResponse AlbumResponse
	if err := json.Unmarshal(body, &albumResponse); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return &albumResponse, nil
}

func (c *AuthClient) GetSimplifiedAlbumDetails(albumID string) (map[string]interface{}, error) {
	// Fetch album details using the existing GetAlbumDetails method
	albumDetails, err := c.GetAlbumDetails(albumID)
	if err != nil {
		return nil, fmt.Errorf("failed to get album details: %w", err)
	}

	// Prepare the simplified response
	simplifiedResponse := map[string]interface{}{
		"AlbumName":  albumDetails.Name,
		"TrackNames": []string{},
	}

	// Extract track names
	for _, track := range albumDetails.Tracks.Items {
		simplifiedResponse["TrackNames"] = append(simplifiedResponse["TrackNames"].([]string), track.Name)
	}

	return simplifiedResponse, nil
}
