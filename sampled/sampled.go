package sampled

type SampledTrack struct {
	Artist string `json:"artist" jsonschema_description:"The artist of the sampled song."`
	Name   string `json:"name" jsonschema_description:"The name of the sampled song."`
}

type Sampled interface {
	GetSample(song, artist string) (*SampledTrack, error)
}
