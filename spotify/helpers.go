package spotify

import "strings"

// generateIDString extracts up to 5 artist IDs from a pointer to ArtistsResponse and creates a comma-separated string.
func generateIDString(idResponse *TopResponse) string {
	var ids []string
	for i, artist := range idResponse.Items {
		if i >= 5 {
			break
		}
		ids = append(ids, artist.ID)
	}
	return strings.Join(ids, ",")
}
