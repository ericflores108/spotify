package db

import (
	"context"
	"fmt"
	"log"
	"time"

	"cloud.google.com/go/bigquery"
	"github.com/ericflores108/spotify/spotify"
	"google.golang.org/api/googleapi"
)

const DatasetID = "spotify"

// Recommendation represents a flat structure for storing recommendations in BigQuery
type Recommendation struct {
	UserID      string `bigquery:"user_id"`
	TrackID     string `bigquery:"track_id"`
	TrackName   string `bigquery:"track_name"`
	AlbumName   string `bigquery:"album_name"`
	ArtistName  string `bigquery:"artist_name"`
	DurationMs  int64  `bigquery:"duration_ms"`
	ReleaseDate string `bigquery:"release_date"`
}

// createTableIfNotExists checks if the BigQuery table exists and creates it if it doesn't
func createTableIfNotExists(ctx context.Context, client *bigquery.Client, tableID string) error {
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
	if err := createTableIfNotExists(ctx, client, tableID); err != nil {
		return fmt.Errorf("failed to ensure table exists: %v", err)
	}

	// Prepare the data to insert
	var rows []*Recommendation
	for _, track := range recommendations.Tracks {
		if len(track.Artists) > 0 {
			// Parse release date and ensure it is in YYYY-MM-DD format
			releaseDate := parseSpotifyReleaseDate(track.Album.ReleaseDate)
			recommendation := &Recommendation{
				UserID:      userID,
				TrackID:     track.ID,
				TrackName:   track.Name,
				AlbumName:   track.Album.Name,
				ArtistName:  track.Artists[0].Name, // Assuming the primary artist
				DurationMs:  int64(track.DurationMs),
				ReleaseDate: releaseDate, // Set the parsed release date
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

// parseSpotifyReleaseDate parses a Spotify release date in "YEAR-MONTH" format to "YYYY-MM-DD"
// If only a year and month are provided, it defaults to the first day of the month (e.g., "2023-05-01").
func parseSpotifyReleaseDate(releaseDate string) string {
	// Try to parse "YEAR-MONTH" format and add "-01" as the day
	if len(releaseDate) == 7 {
		releaseDate += "-01"
	}
	// Attempt to parse as "YYYY-MM-DD" and return the formatted date or an empty string on failure
	parsedDate, err := time.Parse("2006-01-02", releaseDate)
	if err != nil {
		log.Printf("Warning: failed to parse release date '%s': %v", releaseDate, err)
		return "" // Default to an empty string if parsing fails
	}
	return parsedDate.Format("2006-01-02")
}
