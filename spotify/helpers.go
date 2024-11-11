package spotify

import "strings"

// generateIDString takes a slice of strings (IDs) and returns a comma-separated string of up to 5 IDs.
func generateIDString(items []string) string {
	// Limit to the first 5 items, or fewer if items has less than 5 elements.
	if len(items) > 5 {
		items = items[:5]
	}
	return strings.Join(items, ",")
}
