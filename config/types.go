package config

// A struct that will be converted to a Structered Outputs response schema
type SampledTrack struct {
	Artist string `json:"artist" jsonschema_description:"The artist of the sampled song."`
	Name   string `json:"name" jsonschema_description:"The name of the sampled song."`
	Genre  string `json:"genre" jsonschema_description:"The genre of the sampled song."`
}
