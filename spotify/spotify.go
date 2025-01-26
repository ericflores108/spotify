package spotify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const (
	BaseURL = "https://api.spotify.com/v1"
)

type AuthClient struct {
	Client      *http.Client
	AccessToken string
	UserID      string
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

func (c *AuthClient) Post(endpoint string, payload any) (*http.Response, error) {
	// Convert the payload to JSON
	var body io.Reader
	if payload != nil {
		jsonData, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal payload: %w", err)
		}
		body = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest("POST", BaseURL+endpoint, body)
	if err != nil {
		return nil, err
	}

	// Set necessary headers
	req.Header.Set("Authorization", "Bearer "+c.AccessToken)
	req.Header.Set("Content-Type", "application/json")

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
