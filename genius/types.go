package genius

type SearchResponse struct {
	Response struct {
		Hits []struct {
			Result struct {
				ID        int    `json:"id"`
				FullTitle string `json:"full_title"`
				APIPath   string `json:"api_path"`
			} `json:"result"`
		} `json:"hits"`
	} `json:"response"`
}

type SongResponse struct {
	Response struct {
		Song struct {
			ID                int    `json:"id"`
			Title             string `json:"title"`
			URL               string `json:"url"`
			SongRelationships []struct {
				RelationshipType string `json:"type"`
				Songs            []struct {
					Title  string `json:"title"`
					Artist string `json:"primary_artist_names"`
				} `json:"songs"`
			} `json:"song_relationships"`
		} `json:"song"`
	} `json:"response"`
}
