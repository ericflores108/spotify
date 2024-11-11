package spotify

import "strings"

// generateIDString extracts up to 5 artist IDs from a pointer to ArtistsResponse and creates a comma-separated string.
func generateIDString(artistsResponse *ArtistsResponse) string {
	var ids []string
	for i, artist := range artistsResponse.Items {
		if i >= 5 {
			break
		}
		ids = append(ids, artist.ID)
	}
	return strings.Join(ids, ",")
}
