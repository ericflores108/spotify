package sampled

import "context"

type SpotifyTrack struct {
	Name   string
	Artist string
	URI    string
}

type Sampled interface {
	GetSample(ctx context.Context, song, artist string) (*SpotifyTrack, error)
}

type SampledManager struct {
	Sources []Sampled
}

func NewSampledManager(sources ...Sampled) *SampledManager {
	return &SampledManager{
		Sources: sources,
	}
}
