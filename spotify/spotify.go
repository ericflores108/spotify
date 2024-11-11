package spotify

import (
	"fmt"
	"io"
	"log"
	"net/http"
)

const (
	BaseURL = "https://api.spotify.com/v1"
)

type AuthClient struct {
	Client      *http.Client
	AccessToken string
}

// Get creates and sends an authenticated GET request to the Spotify API at the specified endpoint.
// It returns the HTTP response or an error if the request fails.
func (a *AuthClient) Get(endpoint string) (*http.Response, error) {
	req, err := http.NewRequest("GET", BaseURL+endpoint, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+a.AccessToken)
	resp, err := a.Client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		// Read the error body to get more details
		bodyBytes, readErr := io.ReadAll(resp.Body)
		defer resp.Body.Close()

		if readErr != nil {
			return nil, fmt.Errorf("request to %s failed with status %d, and failed to read error body: %v", endpoint, resp.StatusCode, readErr)
		}
		errorBody := string(bodyBytes) // Convert body to string for better display

		return nil, fmt.Errorf("request to %s failed with status %d. \nError: %v. \nResponse body: %s", endpoint, resp.StatusCode, resp.Status, errorBody)
	}

	return resp, err
}

func (a *AuthClient) Recommend() (*RecommendationsResponse, error) {
	// Step 1: Get top artists for the user
	artists, err := a.GetTopItems("artists")
	if err != nil {
		log.Panicf("failed to get top artists: %v", err)
	}

	// Step 2: Use top artists as seeds to get recommendations
	recommendations, err := a.GetRecommendations(generateIDString(artists), "", "")
	if err != nil {
		return nil, fmt.Errorf("failed to get recommendations: %w", err)
	}

	// Step 3: Populate genres for each artist in recommendations
	for i, track := range recommendations.Tracks {
		if len(track.Artists) > 0 {
			// Fetch artist details by ID for the primary artist
			artistID := track.Artists[0].ID
			artistDetails, err := a.GetArtist(artistID)
			if err != nil {
				log.Printf("Warning: failed to get details for artist %s: %v", artistID, err)
				continue // Skip this artist if fetching details fails
			}

			// Populate genres for the primary artist of this track
			recommendations.Tracks[i].Artists[0].Genres = artistDetails.Genres
		}
	}

	// Return the recommendations with populated genres
	return recommendations, nil
}
