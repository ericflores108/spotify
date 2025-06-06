package ai

import (
	"github.com/ericflores108/spotify/config"
	"github.com/invopop/jsonschema"
)

func GenerateSchema[T any]() interface{} {
	// Structured Outputs uses a subset of JSON schema
	// These flags are necessary to comply with the subset
	reflector := jsonschema.Reflector{
		AllowAdditionalProperties: false,
		DoNotReference:            true,
	}
	var v T
	schema := reflector.Reflect(v)
	return schema
}

// Generate the JSON schema at initialization time
var SampleTrackResponseSchema = GenerateSchema[config.SampledTrack]()
