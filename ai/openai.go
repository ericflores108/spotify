package ai

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ericflores108/spotify/config"
	"github.com/openai/openai-go"
)

type AIClient struct {
	Client *openai.Client
}

func (ai *AIClient) FindTrackSamples(ctx context.Context, song, artist string, excludedSongs []string) (*config.SampledTrack, error) {
	// Formulate the question
	question := fmt.Sprintf(`For the song '%s' by '%s': 
	- Suggest one song and its artists that this song samples or draws inspiration from.
	- Exclude the following songs: %s`, song, artist, excludedSongs)

	// Define the schema parameter
	schemaParam := openai.ResponseFormatJSONSchemaJSONSchemaParam{
		Name:        openai.F("track"),
		Description: openai.F("Information about sampled tracks and general inspirations for the specified song"),
		Schema:      openai.F(SampleTrackResponseSchema),
		Strict:      openai.Bool(true),
	}

	// Query the Chat Completions API
	chat, err := ai.Client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: openai.F([]openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(question),
		}),
		ResponseFormat: openai.F[openai.ChatCompletionNewParamsResponseFormatUnion](
			openai.ResponseFormatJSONSchemaParam{
				Type:       openai.F(openai.ResponseFormatJSONSchemaTypeJSONSchema),
				JSONSchema: openai.F(schemaParam),
			},
		),
		Model: openai.F(openai.ChatModelGPT4o2024_08_06),
	})

	if err != nil {
		return nil, fmt.Errorf("failed to query Chat Completions API: %w", err)
	}

	// Parse the response into a SampledTrack struct
	var sampledTrack config.SampledTrack
	err = json.Unmarshal([]byte(chat.Choices[0].Message.Content), &sampledTrack)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Check if the response contains a valid song
	if sampledTrack.Artist == "" || sampledTrack.Name == "" {
		return nil, nil // Return nil if the response does not contain a valid song
	}

	// Return the structured response
	return &sampledTrack, nil
}
