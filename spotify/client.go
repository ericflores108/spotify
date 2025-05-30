package spotify

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
}

func NewSpotifyClient(clientID, clientSecret string) (*AuthClient, error) {
	auth := base64.StdEncoding.EncodeToString([]byte(clientID + ":" + clientSecret))
	url := "https://accounts.spotify.com/api/token"
	data := []byte(`grant_type=client_credentials`)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Authorization", "Basic "+auth)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("failed to retrieve access token: status " + resp.Status)
	}

	body, err := io.ReadAll(resp.Body) // Replaced ioutil.ReadAll with io.ReadAll
	if err != nil {
		return nil, fmt.Errorf("error reading response: %w", err)
	}

	var tokenResponse TokenResponse
	if err := json.Unmarshal(body, &tokenResponse); err != nil {
		return nil, fmt.Errorf("error parsing JSON response: %w", err)
	}

	return &AuthClient{
		Client:      client,
		AccessToken: tokenResponse.AccessToken,
	}, nil
}

func NewSpotifyUserClient(refreshToken, clientID, clientSecret string) (string, error) {
	// Spotify token endpoint
	tokenURL := "https://accounts.spotify.com/api/token"

	// Set up the request body
	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", refreshToken)

	// Encode client credentials in Base64
	authHeader := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", clientID, clientSecret)))

	// Create a new HTTP POST request
	req, err := http.NewRequest("POST", tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Authorization", "Basic "+authHeader)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Make the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// Check for non-200 status code
	if resp.StatusCode != http.StatusOK {
		return "", errors.New("failed to refresh access token, status code: " + resp.Status)
	}

	// Read and parse the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	// Parse the access token from the response
	var token TokenResponse
	if err := json.Unmarshal(body, &token); err != nil {
		return "", fmt.Errorf("failed to parse token response: %w", err)
	}

	if token.AccessToken == "" {
		return "", errors.New("access token not found in response")
	}

	return token.AccessToken, nil
}
