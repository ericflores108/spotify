package db

import (
	"context"
	"fmt"
	"log"

	"cloud.google.com/go/bigquery"
	"github.com/ericflores108/spotify/spotify"
	"google.golang.org/api/googleapi"
)

// Recommendation represents a flat structure for storing recommendations in BigQuery
type Recommendation struct {
	UserID      string   `bigquery:"user_id"`
	TrackID     string   `bigquery:"track_id"`
	TrackName   string   `bigquery:"track_name"`
	AlbumName   string   `bigquery:"album_name"`
	ArtistName  string   `bigquery:"artist_name"`
	DurationMs  int64    `bigquery:"duration_ms"`
	ReleaseDate string   `bigquery:"release_date"`
	Genres      []string `bigquery:"genres"`
	Popularity  int      `bigquery:"popularity"` // New popularity field
}

// createTableIfNotExists checks if the BigQuery table exists and creates it if it doesn't
func createRecommendationTableIfNotExists(ctx context.Context, client *bigquery.Client, tableID string) error {
	dataset := client.Dataset(DatasetID)
	table := dataset.Table(tableID)

	// Check if the table already exists
	_, err := table.Metadata(ctx)
	if err == nil {
		// Table exists
		return nil
	}

	// Check if the error is a "not found" error
	if apiErr, ok := err.(*googleapi.Error); ok && apiErr.Code == 404 {
		// Define the schema for the table
		schema := bigquery.Schema{
			{Name: "user_id", Type: bigquery.StringFieldType},
			{Name: "track_id", Type: bigquery.StringFieldType},
			{Name: "track_name", Type: bigquery.StringFieldType},
			{Name: "album_name", Type: bigquery.StringFieldType},
			{Name: "artist_name", Type: bigquery.StringFieldType},
			{Name: "duration_ms", Type: bigquery.IntegerFieldType},
			{Name: "release_date", Type: bigquery.DateFieldType},
			{Name: "genres", Type: bigquery.StringFieldType, Repeated: true},
			{Name: "popularity", Type: bigquery.IntegerFieldType},
		}

		// Create the table with the defined schema
		if err := table.Create(ctx, &bigquery.TableMetadata{Schema: schema}); err != nil {
			return fmt.Errorf("failed to create table: %v", err)
		}

		log.Println("BigQuery table 'recommendation' created successfully.")
		return nil
	}

	// Return the original error if it's not a "not found" error
	return fmt.Errorf("failed to check if table exists: %v", err)
}

// StoreRecommendations stores a list of recommendations in the BigQuery "recommendation" table
func StoreRecommendations(ctx context.Context, client *bigquery.Client, userID string, recommendations *spotify.RecommendationsResponse) error {
	// Define the BigQuery dataset and table
	tableID := "recommendation"

	// Ensure the table exists
	if err := createRecommendationTableIfNotExists(ctx, client, tableID); err != nil {
		return fmt.Errorf("failed to ensure table exists: %v", err)
	}

	// Prepare the data to insert
	var rows []*Recommendation
	for _, track := range recommendations.Tracks {
		if len(track.Artists) > 0 {
			// Parse release date and ensure it is in YYYY-MM-DD format
			releaseDate := parseSpotifyReleaseDate(track.Album.ReleaseDate)

			// Check if genres are available; if not, set to ["Other"]
			genres := track.Artists[0].Genres
			if len(genres) == 0 {
				genres = []string{"Other"}
			}

			// Prepare recommendation with the new Popularity field
			recommendation := &Recommendation{
				UserID:      userID,
				TrackID:     track.ID,
				TrackName:   track.Name,
				AlbumName:   track.Album.Name,
				ArtistName:  track.Artists[0].Name, // Assuming the primary artist
				DurationMs:  int64(track.DurationMs),
				ReleaseDate: releaseDate, // Set the parsed release date
				Genres:      genres,
				Popularity:  track.Popularity,
			}
			rows = append(rows, recommendation)
		}
	}

	// Insert rows into BigQuery table
	inserter := client.Dataset(DatasetID).Table(tableID).Inserter()
	if err := inserter.Put(ctx, rows); err != nil {
		return fmt.Errorf("failed to insert rows into BigQuery: %v", err)
	}

	log.Println("Successfully stored recommendations in BigQuery.")
	return nil
}
