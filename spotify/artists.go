package spotify

import (
	"encoding/json"
	"fmt"
	"io"
)

// GetArtist retrieves artist information by artist ID.
// It makes an authenticated request to the "artists/{id}" endpoint and returns a pointer to ArtistResponse or an error.
func (c *AuthClient) GetArtist(artistID string) (*ArtistResponse, error) {
	// Build the endpoint with the artist ID
	endpoint := fmt.Sprintf("/artists/%s", artistID)

	// Make the GET request using the AuthClient's Get method
	resp, err := c.Get(endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to get artist info: %w", err)
	}
	defer resp.Body.Close()

	// Read and parse the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var artistResponse ArtistResponse

	// Unmarshal the JSON data into the ArtistResponse struct
	if err := json.Unmarshal(body, &artistResponse); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Return the parsed response and nil error on success
	return &artistResponse, nil
}
