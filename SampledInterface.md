# Purpose

- Create custom Playlist for Spotify users (authenticating via OAuth). 
- Create a Cloud Run service to run Golang web application. 
- Demonstrate use of interfaces.
- Demonstrate use of OpenAI API Structured Outputs. 

[Check out this demo video](https://storage.googleapis.com/titled96demo/titled96.mp4)

[Check out the code](https://github.com/ericflores108/spotify)

## Sampled interface

### Purpose

The purpose of the **Sampled** interface is to create a function, GetSampled(song, artist string) (SpotifyURI, error), that can be implemented regardless of the source (ie, Genius, OpenAI, etc.). 

### Sampled Manager

The **Sampled Manager** will take a list of sources that implement the *Sampled* interface.

The order of the Sampled Manager will determine the priority of the source we want to get the sample from. 

```go
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

```

Use case: 

```go
// this can be genius, openai, etc. order matters when set in main
go func(index int, trackName, artist string) {
  defer wg.Done()
  for _, source := range s.SampledManager.Sources {
    spotifyTrack, err := source.GetSample(ctx, track.Name, artist)
    if err != nil {
      logger.LogError("Error getting %s by %s sample: %v", track.Name, artist, err)
      continue
    }

    if spotifyTrack == nil {
      continue
    }

    mu.Lock()
    // the original track will go before the sampled track
    spotifyTracks[index*2+1] = spotifyTrack.URI
    mu.Unlock()

    break
  }
}(index, track.Name, artist)
```