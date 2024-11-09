package auth

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

// Config holds client credentials
type Config struct {
	ClientID     string
	ClientSecret string
}

// TokenResponse holds the token information
type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

// GetSpotifyToken retrieves an access token from Spotify using client credentials
func GetSpotifyToken(config Config) (string, error) {
	auth := base64.StdEncoding.EncodeToString([]byte(config.ClientID + ":" + config.ClientSecret))
	url := "https://accounts.spotify.com/api/token"
	data := []byte(`grant_type=client_credentials`)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		return "", fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Authorization", "Basic "+auth)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", errors.New("failed to retrieve access token: status " + resp.Status)
	}

	body, err := io.ReadAll(resp.Body) // Replaced ioutil.ReadAll with io.ReadAll
	if err != nil {
		return "", fmt.Errorf("error reading response: %w", err)
	}

	var tokenResponse TokenResponse
	if err := json.Unmarshal(body, &tokenResponse); err != nil {
		return "", fmt.Errorf("error parsing JSON response: %w", err)
	}

	return tokenResponse.AccessToken, nil
}
