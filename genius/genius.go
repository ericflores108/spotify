package genius

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

type GeniusClient struct {
	Client      *http.Client
	AccessToken string
}

// NewClient generates a new GeniusClient by authenticating with the Genius API.
func NewClient(clientID, clientSecret string) (*GeniusClient, error) {
	tokenURL := "https://api.genius.com/oauth/token"

	// Prepare form data
	form := url.Values{}
	form.Set("client_id", clientID)
	form.Set("client_secret", clientSecret)
	form.Set("grant_type", "client_credentials")

	// Make the POST request
	resp, err := http.PostForm(tokenURL, form)
	if err != nil {
		return nil, fmt.Errorf("error making POST request: %w", err)
	}
	defer resp.Body.Close()

	// Check for unsuccessful status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get access token, status: %s", resp.Status)
	}

	// Parse the JSON response
	var responseData struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&responseData); err != nil {
		return nil, fmt.Errorf("error decoding JSON response: %w", err)
	}

	// Check if the token was successfully retrieved
	if responseData.AccessToken == "" {
		return nil, errors.New("access token not found in response")
	}

	// Return the GeniusClient
	return &GeniusClient{
		Client:      &http.Client{},
		AccessToken: responseData.AccessToken,
	}, nil
}

// Search searches for a track by title and artist.
func (g *GeniusClient) Search(track, artist string) (*SearchResponse, error) {
	baseURL := "https://api.genius.com/search"
	params := url.Values{}

	// Find the index of the parenthesis
	if idx := strings.Index(track, "("); idx != -1 {
		// Slice the string to exclude everything from the parenthesis onward
		track = strings.TrimSpace(track[:idx])
	}

	params.Add("q", fmt.Sprintf("%s %s", track, artist))

	req, err := http.NewRequest("GET", fmt.Sprintf("%s?%s", baseURL, params.Encode()), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", g.AccessToken))

	resp, err := g.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %s", resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var searchResponse SearchResponse
	if err := json.Unmarshal(body, &searchResponse); err != nil {
		return nil, err
	}

	return &searchResponse, nil
}

// Songs retrieves detailed information about a song by its ID.
func (g *GeniusClient) Songs(id string) (*SongResponse, error) {
	baseURL := fmt.Sprintf("https://api.genius.com/songs/%s", id)

	req, err := http.NewRequest("GET", baseURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", g.AccessToken))

	resp, err := g.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var songResponse SongResponse
	if err := json.Unmarshal(body, &songResponse); err != nil {
		return nil, err
	}

	return &songResponse, nil
}
