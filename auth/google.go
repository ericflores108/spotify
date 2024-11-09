package auth

import (
	"context"
	"fmt"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
)

// GetSecret retrieves the secret data from Google Secret Manager.
func GetSecret(ctx context.Context, client *secretmanager.Client, projectID, secretID string) (string, error) {
	// Define the resource name for the secret version to access.
	secretVersionName := fmt.Sprintf("projects/%s/secrets/%s/versions/1", projectID, secretID)

	// Build the request to access the secret version.
	accessRequest := &secretmanagerpb.AccessSecretVersionRequest{
		Name: secretVersionName,
	}

	// Call the API to access the secret version.
	result, err := client.AccessSecretVersion(ctx, accessRequest)
	if err != nil {
		return "", fmt.Errorf("failed to access secret version: %w", err)
	}

	// Return the secret data as a string.
	return string(result.Payload.Data), nil
}
