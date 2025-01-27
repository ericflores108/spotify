package spotify

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"
)

func (c *AuthClient) GetTrackURI(trackName, artistName string) (string, error) {
	query := url.Values{}

	if artistName != "" {
		query.Set("q", fmt.Sprintf("track:%s artist:%s", trackName, artistName))
	} else {
		query.Set("q", fmt.Sprintf("track:%s", trackName))
	}

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
