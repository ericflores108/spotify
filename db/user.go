package db

import (
	"context"
	"fmt"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
)

// User represents a user document in Firestore
type User struct {
	ID           string `firestore:"id"`
	DisplayName  string `firestore:"display_name"`
	AccessToken  string `firestore:"access_token"`
	RefreshToken string `firestore:"refresh_token"`
}

// GetAllUsers retrieves all user documents from the SpotifyUser collection
func GetAllUsers(ctx context.Context, client *firestore.Client) ([]User, error) {
	var users []User

	// Reference the SpotifyUser collection
	iter := client.Collection("SpotifyUser").Documents(ctx)
	defer iter.Stop() // Close iterator after use

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate documents: %w", err)
		}

		// Map document data to a User struct
		var user User
		if err := doc.DataTo(&user); err != nil {
			return nil, fmt.Errorf("failed to map document data: %w", err)
		}
		user.ID = doc.Ref.ID // Optionally add Firestore document ID
		users = append(users, user)
	}

	return users, nil
}
