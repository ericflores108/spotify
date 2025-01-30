package spotify

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func (c *AuthClient) GetAlbum(albumID string) (*Album, error) {
	url := fmt.Sprintf("/albums/%s", albumID)

	req, err := http.NewRequest("GET", BaseURL+url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.AccessToken)

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var albumResponse Album
	if err := json.Unmarshal(body, &albumResponse); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w\nResponse Body: %s\nHTTP Status: %d - %s",
			err, string(body), resp.StatusCode, http.StatusText(resp.StatusCode))
	}

	return &albumResponse, nil
}

func (c *AuthClient) GetAlbumTracks(albumID string) (*AlbumResponse, error) {
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
