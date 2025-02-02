package sampled

import (
	"context"
	"fmt"
	"strconv"

	"github.com/ericflores108/spotify/genius"
	"github.com/ericflores108/spotify/logger"
	"github.com/ericflores108/spotify/spotify"
)

type GeniusService struct {
	Spotify *spotify.AuthClient
	Genius  *genius.GeniusClient
}

func (g *GeniusService) GetSample(ctx context.Context, song, artist string) (*SpotifyTrack, error) {
	geniusSearch, err := g.Genius.Search(song, artist)
	if err != nil {
		return nil, fmt.Errorf("Could not search Genius: %v", err)
	}

	if len(geniusSearch.Response.Hits) == 0 {
		logger.LogDebug("Genius search has no hits")
		return nil, nil
	}

	geniusSong, err := g.Genius.Songs(strconv.Itoa(geniusSearch.Response.Hits[0].Result.ID))
	if err != nil {
		return nil, fmt.Errorf("Could not get Genius song: %v", err)
	}

	if len(geniusSong.Response.Song.SongRelationships) == 0 {
		logger.LogDebug("Song has no song relationships")
		return nil, nil
	}

	var spotifyTrack *SpotifyTrack

	for _, relation := range geniusSong.Response.Song.SongRelationships {
		if relation.RelationshipType == "samples" && len(relation.Songs) > 0 {
			spotifyTrack = &SpotifyTrack{
				Artist: relation.Songs[0].Artist,
				Name:   relation.Songs[0].Title,
			}
			break
		}
	}

	if spotifyTrack == nil {
		logger.LogDebug("Song has no Genius samples")
		return nil, nil
	}

	// Get Spotify URI
	trackURI, err := g.Spotify.GetTrackURI(spotifyTrack.Name, spotifyTrack.Artist)
	if err != nil {
		logger.LogError("Error occurred at trackURI: %v", err)
		return nil, fmt.Errorf("Could not get trackURI: %v", err)
	}

	if trackURI == "" {
		logger.LogDebug("No trackURI found for - TRACK - %s - ARTIST - %s", spotifyTrack.Name, spotifyTrack.Artist)
		return nil, nil
	}

	spotifyTrack.URI = trackURI

	return spotifyTrack, nil
}
