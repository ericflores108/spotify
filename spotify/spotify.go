package spotify

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"slices"
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
func (c *AuthClient) Get(endpoint string) (*http.Response, error) {
	req, err := http.NewRequest("GET", BaseURL+endpoint, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.AccessToken)
	resp, err := c.Client.Do(req)
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

func (c *AuthClient) Recommend() (*RecommendationsResponse, error) {
	// Step 1: Get top artists for the user
	topArtistsRes, err := c.GetTopItems(Artists)
	if err != nil {
		log.Panicf("failed to get top artists: %v", err)
	}

	var artistIDs []string
	for artist := range slices.Values(topArtistsRes.Items) {
		// Check if item is a map and contains an "id" key
		if artistMap, ok := artist.(map[string]any); ok {
			// Attempt to get the "id" value as a string
			if id, exists := artistMap["id"].(string); exists {
				artistIDs = append(artistIDs, id)
			} else {
				log.Printf("ID not found or is not a string in item: %v", artistMap)
			}
		} else {
			log.Printf("Item is not a map['id']any: %v", artist)
		}
	}

	// Step 2: Use top artists as seeds to get recommendations
	recommendations, err := c.GetRecommendations(generateIDString(artistIDs), "", "")
	if err != nil {
		return nil, fmt.Errorf("failed to get recommendations: %w", err)
	}

	// Step 3: Populate genres for each artist in recommendations
	for i, track := range recommendations.Tracks {
		if len(track.Artists) > 0 {
			// Fetch artist details by ID for the primary artist
			artistID := track.Artists[0].ID
			artistDetails, err := c.GetArtist(artistID)
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

// Tracks retrieves the top tracks for the user and converts them into a TopTracksResponse
func (c *AuthClient) TopTracks() (*TopTracksResponse, error) {
	// Step 1: Get top items with items as `[]any`
	topTracksRes, err := c.GetTopItems(Tracks)
	if err != nil {
		return nil, fmt.Errorf("failed to get top tracks: %v", err)
	}

	// Step 2: Convert TopResponse.Items to TopTracksResponse.Items
	topTracks := &TopTracksResponse{
		Href:     topTracksRes.Href,
		Limit:    topTracksRes.Limit,
		Next:     topTracksRes.Next,
		Offset:   topTracksRes.Offset,
		Previous: topTracksRes.Previous,
		Total:    topTracksRes.Total,
	}

	for _, item := range topTracksRes.Items {
		// Convert each item from map[string]interface{} to Track
		if trackMap, ok := item.(map[string]any); ok {
			var track Track
			trackJSON, err := json.Marshal(trackMap) // Convert map to JSON
			if err != nil {
				return nil, fmt.Errorf("failed to marshal track item: %v", err)
			}
			err = json.Unmarshal(trackJSON, &track) // Convert JSON to Track struct
			if err != nil {
				return nil, fmt.Errorf("failed to unmarshal track item: %v", err)
			}
			topTracks.Items = append(topTracks.Items, track)
		} else {
			return nil, fmt.Errorf("item is not a map[string]interface{}: %v", item)
		}
	}

	return topTracks, nil
}
