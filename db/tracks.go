package db

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/ericflores108/spotify/logger"
	"google.golang.org/api/iterator"
)

type Tracks struct {
	ID     string    `firestore:"id"`
	Tracks []string  `firestore:"tracks"`
	TTL    time.Time `firestore:"ttl"`
}

const TrackCollection = "SpotifyTracks"

func GetTracks(ctx context.Context, client *firestore.Client, ID string) ([]string, error) {
	query := client.Collection(TrackCollection).Where("id", "==", ID).Limit(1)

	iter := query.Documents(ctx)
	defer iter.Stop()

	doc, err := iter.Next()
	if err != nil {
		if err == iterator.Done {
			return nil, fmt.Errorf("track with ID %s not found", ID)
		}
		logger.LogError("Error occurred at GetTracks iter.Next(): %v", err)
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	var tracks Tracks
	if err := doc.DataTo(&tracks); err != nil {
		logger.LogError("Error occurred at GetTracks doc.DataTo: %v", err)
		return nil, fmt.Errorf("failed to map document data: %w", err)
	}

	return tracks.Tracks, nil
}

func SetTracks(ctx context.Context, client *firestore.Client, ID string, tracks []string) error {
	_, _, err := client.Collection(TrackCollection).Add(ctx, Tracks{
		ID:     ID,
		Tracks: tracks,
		TTL:    time.Now().Add(7 * 24 * time.Hour),
	})
	if err != nil {
		logger.LogError("Error occurred at SetAlbumTracks: %v", err)
		return fmt.Errorf("failed to store tracks: %w", err)
	}

	return nil
}
