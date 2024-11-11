package spotify

import (
	"fmt"
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
		defer resp.Body.Close()
		return nil, fmt.Errorf("request to %s failed with status %d", endpoint, resp.StatusCode)
	}

	return resp, err
}
