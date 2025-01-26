
# AI Client with Structured Outputs and JSON Schema Integration

This project demonstrates how to use OpenAI's API with structured outputs using JSON schemas to ensure a predictable and reliable response format. The implementation uses the [openai-go](https://github.com/openai/openai-go) library and integrates with `jsonschema` for schema generation and validation.

---

## Key Concepts

### Structured Outputs

Structured outputs ensure that the API responses adhere to a predefined schema. This provides:
- Predictability: Responses follow a strict format, reducing parsing errors.
- Validation: Schemas enforce data integrity.
- Simplified Integration: Developers can confidently use the output knowing it adheres to the expected structure.

The application integrates OpenAI's structured outputs feature using JSON Schema definitions.

---

### JSON Schema Integration

JSON Schema is used to define the structure and properties of the expected response. The schema is generated dynamically using the `jsonschema` package in Go. 

For example, the `SampledTrack` struct is used to define the schema for the track sampling information:

```go
type SampledTrack struct {
	Artist string `json:"artist" jsonschema_description:"The artist of the sampled song."`
	Name   string `json:"name" jsonschema_description:"The name of the sampled song."`
	Genre  string `json:"genre" jsonschema_description:"The genre of the sampled song."`
}
```

The schema is generated with:

```go
func GenerateSchema[T any]() interface{} {
	reflector := jsonschema.Reflector{
		AllowAdditionalProperties: false,
		DoNotReference:            true,
	}
	var v T
	schema := reflector.Reflect(v)
	return schema
}

var SampleTrackResponseSchema = GenerateSchema[SampledTrack]()
```

This ensures that all responses match the defined structure.

---

### OpenAI Integration

The application leverages OpenAI's Chat Completions API to find sampled tracks for a given song and artist. The implementation includes:
1. **Question Formulation**: The user query is dynamically generated with the given song, artist, and excluded songs.
2. **Schema Validation**: The API enforces adherence to the JSON schema to ensure valid responses.
3. **Parsing Responses**: The structured response is unmarshaled into a `SampledTrack` object.

---

### Workflow

1. **Define the Schema**: Use the `SampledTrack` struct to define the expected output.
2. **Generate the Schema**: Use the `jsonschema` package to create the schema dynamically.
3. **Call OpenAI API**: Query the API with the schema and parse the response.
4. **Process Results**: Validate and return the structured response.

---

## Example Code

### Querying Sampled Tracks

Here is an example function to query sampled tracks using OpenAI's API:

```go
func (ai *AIClient) FindTrackSamples(ctx context.Context, song, artist string, excludedSongs []string) (*SampledTrack, error) {
	question := fmt.Sprintf("For the song '%s' by '%s', suggest one song and its artists that this song samples or draws inspiration from. Exclude: %s", song, artist, excludedSongs)

	chat, err := ai.Client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(question),
		},
		ResponseFormat: openai.ResponseFormatJSONSchemaParam{
			Schema: SampleTrackResponseSchema,
			Strict: true,
		},
		Model: openai.ChatModelGPT4o2024_08_06,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to query Chat Completions API: %w", err)
	}

	var sampledTrack SampledTrack
	err = json.Unmarshal([]byte(chat.Choices[0].Message.Content), &sampledTrack)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if sampledTrack.Artist == "" || sampledTrack.Name == "" {
		return nil, nil
	}

	return &sampledTrack, nil
}
```

---

## Structured Outputs Example

Explore OpenAI's demo on structured outputs here: [Structured Outputs Demo](https://github.com/openai/openai-go/tree/main/examples/structured-outputs).

---

## Benefits of This Approach

- **Reliable Responses**: Structured outputs ensure adherence to the schema.
- **Error Reduction**: Prevents unexpected formats and parsing issues.
- **Extensibility**: Schemas can be updated or extended as needed.
- **Easy Integration**: Compatible with existing workflows for predictable behavior.

---

# Spotify App

The Spotify Authentication App is a Go-based web application that facilitates Spotify API authentication and interaction. It uses OAuth 2.0 for secure user authentication and integrates with Google Cloud services such as Firestore and Secret Manager. The app is designed to be lightweight and customizable for different environments, including production and local development.

![Beta](https://img.shields.io/badge/status-beta-yellow)

This project is currently in beta. Features and functionality are subject to change.

## Features

- **Spotify OAuth 2.0 Authentication**: Secure user authentication using Spotify's API.
- **Local and Production Modes**: Easily toggle between localhost and production redirect URLs.
- **Google Cloud Integration**: Leverages Firestore and Secret Manager for secure data storage and retrieval.
- **Extensible Service Layer**: Built with modular components to support additional integrations (e.g., OpenAI).

---

## Prerequisites

To use this application, ensure the following dependencies are installed and configured:

### Spotify API Credentials

1. Create a Spotify Developer account at [Spotify for Developers](https://developer.spotify.com/).
2. Create an application and obtain the following credentials:
   - **Client ID**
   - **Client Secret**
3. Set up redirect URIs:
   - For local development: `http://localhost:8080/callback`

### Google Cloud Services

1. **Secret Manager**: Store sensitive data like the Spotify Client ID and Secret.
2. **Firestore**: Used for persistent storage of user data and application state.

### Environment Variables

Define the following environment variables:

- `PORT`: The port on which the app runs (default: `8080`).
- `GOOGLE_APPLICATION_CREDENTIALS`: Path to the Google Cloud service account JSON file.

---

## Installation

### Clone the Repository

```bash
git clone https://github.com/ericflores108/spotify.git
cd spotify
```

### Install Dependencies

Ensure you have Go installed. Then, run:

```bash
go mod tidy
```

---

## Usage

### Start the Application

To start the application, run:

```bash
go run main.go
```

### Available Flags

- `-useLocalHost`: Toggle the redirect URL between production and localhost. 
  - Default: Production URL (`https://spotify-123259034538.us-west1.run.app/callback`)
  - Example for localhost:

    ```bash
    go run main.go -useLocalHost
    ```

### Local Development Example

To test the application locally, ensure the following steps:

1. Set up a local Spotify redirect URI (`http://localhost:8080/callback`) in your Spotify Developer application.
2. Run the app with the `-useLocalHost` flag.
3. Visit `http://localhost:8080` to initiate authentication.

---

## Configuration

The application uses the following components:

### Spotify API

The Spotify service is initialized using:

- **Client ID**: Retrieved from Secret Manager.
- **Client Secret**: Retrieved from Secret Manager.
- **Redirect URL**: Set dynamically based on the `-useLocalHost` flag.

### Google Cloud Services

- **Firestore**: Used for storing application data and state.
- **Secret Manager**: Used for securely storing sensitive credentials.

---

## API Endpoints

The app provides the following HTTP endpoints:

1. `/`: Home page to initiate Spotify authentication.
2. `/callback`: Handles the Spotify OAuth callback and finalizes authentication.
3. `/refresh`: Refreshes the Spotify access token.

---

## Logging

The application uses a structured logger to provide detailed logs. Logs are categorized into:

- **Info**: General application events.
- **Error**: Errors and issues encountered during runtime.

Example log output:

```
INFO: starting app
INFO: listening on port 8080
ERROR: Failed to start server on port 8080: <error_message>
```

---

## Deployment

### Local Development

Run the application locally:

```bash
go run main.go -useLocalHost
```

### Production Deployment

The application can be deployed to Google Cloud Run or any containerized environment. Example:

1. Build the Docker image:

    ```bash
    docker build -t spotify-auth-app .
    ```

2. Deploy the image to Google Cloud Run or any other hosting platform.

---

## Example

Here’s an example of how to authenticate a user and retrieve their Spotify profile data:

1. Start the app locally.
2. Navigate to `http://localhost:8080`.
3. Log in with your Spotify account.
4. Once authenticated, view your Spotify profile data.

---

## Contact

For questions or suggestions, contact the maintainer:

- **Eric Flores**
- Email: eflorty108@gmail.com
