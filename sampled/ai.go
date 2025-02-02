package sampled

import (
	"context"
	"fmt"

	"github.com/ericflores108/spotify/ai"
	"github.com/ericflores108/spotify/logger"
	"github.com/ericflores108/spotify/spotify"
)

type AIService struct {
	Spotify *spotify.AuthClient
	AI      *ai.AIClient
}

func (a *AIService) GetSample(ctx context.Context, song, artist string) (*SpotifyTrack, error) {
	aiSearch, err := a.AI.FindTrackSamples(ctx, song, artist)
	if err != nil {
		logger.LogError("Error occurred at AI aiSearch: %v", err)
		return nil, fmt.Errorf("Could not find aiSearch: %v", err)
	}

	if aiSearch == nil {
		logger.LogDebug("No AI sampledTrack found.")
		return nil, nil
	}

	spotifyTrack := &SpotifyTrack{
		Name:   aiSearch.Name,
		Artist: aiSearch.Artist,
	}

	// Get Spotify URI
	trackURI, err := a.Spotify.GetTrackURI(aiSearch.Name, aiSearch.Artist)
	if err != nil {
		logger.LogError("Error occurred at trackURI: %v", err)
		return nil, fmt.Errorf("Could not get trackURI: %v", err)
	}

	if trackURI == "" {
		logger.LogDebug("No trackURI found for - TRACK - %s - ARTIST - %s", aiSearch.Name, aiSearch.Artist)
		return nil, nil
	}

	spotifyTrack.URI = trackURI

	return spotifyTrack, nil
}
