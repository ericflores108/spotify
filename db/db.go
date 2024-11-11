package db

import (
	"log"
	"time"
)

const DatasetID = "spotify"

// parseSpotifyReleaseDate parses various Spotify release date formats to "YYYY-MM-DD".
// If only the year is provided, it defaults to the first day of that year ("YYYY-01-01").
// Returns an empty string if parsing fails, to represent NULL in BigQuery.
func parseSpotifyReleaseDate(releaseDate string) string {
	// Handle cases with only a year (e.g., "1972")
	if len(releaseDate) == 4 {
		releaseDate += "-01-01" // Default to January 1st of that year
	} else if len(releaseDate) == 7 {
		// Handle cases with year and month (e.g., "1980-05")
		releaseDate += "-01" // Default to the first day of the month
	}

	// Attempt to parse as "YYYY-MM-DD"
	parsedDate, err := time.Parse("2006-01-02", releaseDate)
	if err != nil {
		log.Printf("Warning: failed to parse release date '%s': %v", releaseDate, err)
		return "" // Use empty string if parsing fails
	}

	return parsedDate.Format("2006-01-02")
}
