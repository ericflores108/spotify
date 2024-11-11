package spotify

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
)

// GetTopItems retrieves the user's top artists or tracks from Spotify, based on the specified TopType.
// It makes an authenticated request to the "me/top/{type}" endpoint and outputs the list of top items.
// The function handles JSON parsing and prints out the names of the top items.
func (a *AuthClient) GetTopItems(top TopType) {
	resp, err := a.Get("/me/top/" + string(top))
	if err != nil {
		log.Fatalf("Failed to get response: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Failed to read response body: %v", err)
	}

	var artistsResponse ArtistsResponse
	if err := json.Unmarshal(body, &artistsResponse); err != nil {
		log.Fatalf("Failed to parse JSON: %v", err)
	}

	fmt.Println("Top Artists:", artistsResponse.Total)
	for _, artist := range artistsResponse.Items {
		fmt.Println(artist.Name)
	}
}
