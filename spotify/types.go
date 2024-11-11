package spotify

type TopResponse struct {
	Href     string   `json:"href"`
	Limit    int      `json:"limit"`
	Next     string   `json:"next"`
	Offset   int      `json:"offset"`
	Previous string   `json:"previous"`
	Total    int      `json:"total"`
	Items    []Artist `json:"items"`
}

type Artist struct {
	ExternalURLs ExternalURLs `json:"external_urls"`
	Followers    Followers    `json:"followers"`
	Genres       []string     `json:"genres"`
	Href         string       `json:"href"`
	ID           string       `json:"id"`
	Images       []Image      `json:"images"`
	Name         string       `json:"name"`
	Popularity   int          `json:"popularity"`
	Type         string       `json:"type"`
	URI          string       `json:"uri"`
}

type ExternalURLs struct {
	Spotify string `json:"spotify"`
}

type Followers struct {
	Href  string `json:"href"`
	Total int    `json:"total"`
}

type Image struct {
	URL    string `json:"url"`
	Height int    `json:"height"`
	Width  int    `json:"width"`
}

type TopType string

const (
	Artists TopType = "artists"
	Tracks  TopType = "tracks"
)

// RecommendationsResponse represents the main structure of the Spotify recommendations response.
type RecommendationsResponse struct {
	Seeds  []Seed  `json:"seeds"`
	Tracks []Track `json:"tracks"`
}

// Seed represents the information about a seed item.
type Seed struct {
	AfterFilteringSize int    `json:"afterFilteringSize"`
	AfterRelinkingSize int    `json:"afterRelinkingSize"`
	Href               string `json:"href"`
	ID                 string `json:"id"`
	InitialPoolSize    int    `json:"initialPoolSize"`
	Type               string `json:"type"`
}

// Track represents a track item in the recommendations.
type Track struct {
	Album            Album         `json:"album"`
	Artists          []Artist      `json:"artists"`
	AvailableMarkets []string      `json:"available_markets"`
	DiscNumber       int           `json:"disc_number"`
	DurationMs       int           `json:"duration_ms"`
	Explicit         bool          `json:"explicit"`
	ExternalIDs      ExternalIDs   `json:"external_ids"`
	ExternalURLs     ExternalURLs  `json:"external_urls"`
	Href             string        `json:"href"`
	ID               string        `json:"id"`
	IsPlayable       bool          `json:"is_playable"`
	Restrictions     *Restrictions `json:"restrictions,omitempty"`
	Name             string        `json:"name"`
	Popularity       int           `json:"popularity"`
	PreviewURL       string        `json:"preview_url"`
	TrackNumber      int           `json:"track_number"`
	Type             string        `json:"type"`
	URI              string        `json:"uri"`
	IsLocal          bool          `json:"is_local"`
}

// Album represents the album details for a track.
type Album struct {
	AlbumType            string        `json:"album_type"`
	TotalTracks          int           `json:"total_tracks"`
	AvailableMarkets     []string      `json:"available_markets"`
	ExternalURLs         ExternalURLs  `json:"external_urls"`
	Href                 string        `json:"href"`
	ID                   string        `json:"id"`
	Images               []Image       `json:"images"`
	Name                 string        `json:"name"`
	ReleaseDate          string        `json:"release_date"`
	ReleaseDatePrecision string        `json:"release_date_precision"`
	Restrictions         *Restrictions `json:"restrictions,omitempty"`
	Type                 string        `json:"type"`
	URI                  string        `json:"uri"`
	Artists              []Artist      `json:"artists"`
}

// ExternalIDs represents various external IDs for a track.
type ExternalIDs struct {
	ISRC string `json:"isrc"`
	EAN  string `json:"ean"`
	UPC  string `json:"upc"`
}

// Restrictions represents restrictions on a track or album.
type Restrictions struct {
	Reason string `json:"reason"`
}

type ArtistResponse struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	Genres     []string `json:"genres"`
	Popularity int      `json:"popularity"`
	Followers  struct {
		Total int `json:"total"`
	} `json:"followers"`
	Images []struct {
		URL    string `json:"url"`
		Height int    `json:"height"`
		Width  int    `json:"width"`
	} `json:"images"`
}
