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
	artists, err := a.GetTopItems("artists")
	if err != nil {
		log.Panicf("failed to get top artists: %v", err)
	}

	return a.GetRecommendations(generateIDString(artists), "", "")
}
