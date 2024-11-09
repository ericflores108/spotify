package spotify

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

const (
	topURL = "https://api.spotify.com/v1/me/top/"
)

func GetTopItems(top TopType, accessToken string) {
	req, err := http.NewRequest("GET", topURL+string(top), nil)
	if err != nil {
		log.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
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
